package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/workflow/dag"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultExecutionTimeout = 15 * time.Minute
	DefaultCleanupTimeout   = 2 * time.Minute
)

var (
	ErrOpExecSystemFailure       = errors.New("Operator execution failed due to system error.")
	ErrOpExecBlockingUserFailure = errors.New("Operator execution failed due to user error.")
)

type Orchestrator interface {
	Execute(ctx context.Context, dag dag.WorkflowDag) (shared.ExecutionStatus, error)

	// Finish is an end-of-orchestration hook meant to do any final cleanup work, after Execute completes.
	Finish(ctx context.Context)
}

type AqueductTimeConfig struct {
	// Configures exactly long we wait before polling again on an in-progress operator.
	OperatorPollInterval time.Duration

	// Configures the maximum amount of time we wait for execution to finish before aborting the run.
	ExecTimeout time.Duration

	// Configures the maximum amount of time we want for any leftover, in-progress operators to complete,
	// after execution has already finished. Once this time is exceeded, we'll give up.
	CleanupTimeout time.Duration
}

type aqOrchestrator struct {
	dag        dag.WorkflowDag
	jobManager job.JobManager
	timeConfig *AqueductTimeConfig

	inProgressOps map[uuid.UUID]operator.Operator
	completedOps  map[uuid.UUID]operator.Operator
	status        shared.ExecutionStatus

	shouldPersistResults bool
}

func NewAqOrchestrator(
	dag dag.WorkflowDag,
	jobManager job.JobManager,
	timeConfig AqueductTimeConfig,
	shouldPersistResults bool,
) Orchestrator {
	return &aqOrchestrator{
		dag:                  dag,
		jobManager:           jobManager,
		timeConfig:           &timeConfig,
		inProgressOps:        make(map[uuid.UUID]operator.Operator, len(dag.Operators())),
		completedOps:         make(map[uuid.UUID]operator.Operator, len(dag.Operators())),
		status:               shared.PendingExecutionStatus,
		shouldPersistResults: shouldPersistResults,
	}
}

func (orch *aqOrchestrator) Execute(
	ctx context.Context,
	dag dag.WorkflowDag,
) (shared.ExecutionStatus, error) {
	orch.status = shared.RunningExecutionStatus
	log.Errorf("Should persist results! %s", orch.shouldPersistResults)
	err := orch.execute(
		ctx,
		dag,
		orch.timeConfig,
		orch.jobManager,
		orch.shouldPersistResults,
	)
	if err != nil {
		orch.status = shared.FailedExecutionStatus
	} else {
		orch.status = shared.SucceededExecutionStatus
	}

	if orch.shouldPersistResults {
		err = dag.PersistResult(ctx, orch.status)
		if err != nil {
			log.Errorf("Error when persisting dag resutls: %v", err)
		}
	}
	return orch.status, err
}

func waitForInProgressOperators(
	ctx context.Context,
	inProgressOps map[uuid.UUID]operator.Operator,
	pollInterval time.Duration,
	timeout time.Duration,
) {
	start := time.Now()
	for len(inProgressOps) > 0 {
		if time.Since(start) > timeout {
			return
		}

		for opID, op := range inProgressOps {
			execState, err := op.GetExecState(ctx)

			// Resolve any jobs that aren't actively running or failed. We don't are if they succeeded or failed,
			// since this is called after orchestration exits.
			if err != nil || execState.Status != shared.RunningExecutionStatus {
				delete(inProgressOps, opID)
			}
		}
		time.Sleep(pollInterval)
	}
}

func opFailureError(failureType shared.FailureType, op operator.Operator) error {
	if failureType == shared.SystemFailure {
		return ErrOpExecSystemFailure
	} else if failureType == shared.UserFailure {
		log.Errorf("Failed due to user error. Operator name %s, id %s", op.Name(), op.ID())
		return ErrOpExecBlockingUserFailure
	}
	return errors.Newf("Internal error: Unsupported failure type %s", failureType)
}

func (orch *aqOrchestrator) execute(
	ctx context.Context,
	dag dag.WorkflowDag,
	timeConfig *AqueductTimeConfig,
	jobManager job.JobManager,
	shouldPersistResults bool,
) error {
	// These are the operators of immediate interest. They either need to be scheduled or polled on.
	inProgressOps := orch.inProgressOps
	completedOps := orch.completedOps

	// Kick off execution by starting all operators that don't have any inputs.
	for _, op := range dag.Operators() {
		if op.Ready(ctx) {
			inProgressOps[op.ID()] = op
		}
	}

	if len(inProgressOps) == 0 {
		return errors.Newf("No initial operators to schedule.")
	}

	// Wait a little bit for all active operators to finish before exiting on failure.
	defer waitForInProgressOperators(ctx, inProgressOps, timeConfig.OperatorPollInterval, timeConfig.CleanupTimeout)

	start := time.Now()

	// TODO(kenxu): documentation
	for len(inProgressOps) > 0 {
		if time.Since(start) > timeConfig.ExecTimeout {
			return errors.New("Reached timeout waiting for workflow to complete.")
		}

		for _, op := range inProgressOps {
			execState, err := op.GetExecState(ctx)
			if err != nil {
				return err
			}

			log.Errorf("Exec Status %s.", execState.Status)
			if execState.Status == shared.PendingExecutionStatus {
				log.Errorf("Scheduling op %s.", op.Name())

				spec := op.JobSpec()
				err = jobManager.Launch(ctx, spec.JobName(), spec)
				if err != nil {
					return errors.Wrapf(err, "Unable to schedule operator %s.", op.Name())
				}
				continue
			} else if execState.Status == shared.RunningExecutionStatus {
				log.Errorf("Op %s is running.", op.Name())
				continue
			}
			if execState.Status != shared.FailedExecutionStatus && execState.Status != shared.SucceededExecutionStatus {
				return errors.Newf("Internal error: the operator is expected to have terminated, but instead has status %s", execState.Status)
			}

			// From here on we can assume that the operator has terminated.
			log.Errorf("Op %s has terminated with status %s.", op.Name(), execState.Status)
			if shouldPersistResults {
				log.Errorf("Op %s is persisting results", op.Name())
				err = op.PersistResult(ctx)
				if err != nil {
					return errors.Wrapf(err, "Error when finishing execution of operator %s", op.Name())
				}
			}

			if execState.Status == shared.FailedExecutionStatus {
				return opFailureError(*execState.FailureType, op)
			}

			// The operator has succeeded! Add the operator to the completed stack, and remove it from the in-progress one.
			if _, ok := completedOps[op.ID()]; ok {
				return errors.Newf("Internal error: operator %s was completed twice.", op.Name())
			}
			completedOps[op.ID()] = op
			delete(inProgressOps, op.ID())

			outputArtifacts, err := dag.ArtifactsFromOperator(op)
			if err != nil {
				return err
			}
			for _, outputArtifact := range outputArtifacts {
				nextOps, err := dag.OperatorsOnArtifact(outputArtifact)
				if err != nil {
					return err
				}

				for _, nextOp := range nextOps {
					log.Errorf("Looking at scheduling %s.", nextOp.Name())
					// Before scheduling the next operator, check that all upstream artifacts to that operator
					// have been computed.
					if !nextOp.Ready(ctx) {
						log.Errorf("%s is not ready yet.", nextOp.Name())
						continue
					}

					// Defensive check: do not reschedule an already in-progress operator. This shouldn't actually
					// matter because we only keep and update a single copy an on operator.
					if _, ok := inProgressOps[nextOp.ID()]; !ok {
						inProgressOps[nextOp.ID()] = nextOp
					}
				}
			}

			time.Sleep(timeConfig.OperatorPollInterval)
		}
	}

	if len(completedOps) != len(dag.Operators()) {
		return errors.Newf(fmt.Sprintf("Internal error: %d operators were provided but only %d completed.", len(dag.Operators()), len(completedOps)))
	}
	return nil
}

func (orch *aqOrchestrator) Finish(ctx context.Context) {
	for _, op := range orch.dag.Operators() {
		op.Finish(ctx)
	}
}
