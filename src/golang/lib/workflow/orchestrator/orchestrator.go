package orchestrator

import (
	"context"
	"fmt"
	dag "github.com/aqueducthq/aqueduct/lib/workflow"
	operator2 "github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/scheduler"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	defaultTimeout = 15 * time.Minute
)

var (
	ErrIncorrectOperatorsScheduled = errors.New("Incorrect number of operators scheduled.")
	ErrOpExecSystemFailure         = errors.New("Operator execution failed due to system error.")
	ErrOpExecBlockingUserFailure   = errors.New("Operator execution failed due to user error.")
	ErrInvalidOpId                 = errors.New("Invalid operator ID.")
	ErrInvalidArtifactId           = errors.New("Invalid artifact ID.")
)

type Orchestrator interface {
	Execute(ctx context.Context, dag dag.WorkflowDag) (shared.ExecutionStatus, error)
}

type orchestratorImpl struct {
	jobManager   job.JobManager
	pollInterval time.Duration

	shouldPersistResults bool
}

func NewOrchestrator(
	jobManager job.JobManager,
	shouldPersistResults bool,
) Orchestrator {
	return &orchestratorImpl{
		jobManager:           jobManager,
		pollInterval:         time.Millisecond * 500,
		shouldPersistResults: shouldPersistResults,
	}
}

func (orch *orchestratorImpl) Execute(
	ctx context.Context,
	dag dag.WorkflowDag,
) (shared.ExecutionStatus, error) {
	defer dag.Finish(ctx)

	status := shared.SucceededExecutionStatus
	err := execute(
		ctx,
		dag,
		orch.pollInterval,
		defaultTimeout, /* timeout */
		2*time.Minute,  /* cleanupTimeout */
		orch.jobManager,
		orch.shouldPersistResults,
	)
	if err != nil {
		status = shared.FailedExecutionStatus
	}

	if orch.shouldPersistResults {
		err = dag.PersistResult(ctx, status)
		if err != nil {
			log.Errorf("Error when persisting dag resutls: %v", err)
		}
	}
	return status, err
}

func waitForInProgressOperators(
	ctx context.Context,
	inProgressOps map[uuid.UUID]operator2.Operator,
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

func opFailureError(failureType shared.FailureType, op operator2.Operator) error {
	if failureType == shared.SystemFailure {
		return ErrOpExecSystemFailure
	} else if failureType == shared.UserFailure {
		log.Errorf("Failed due to user error. Operator name %s, id %s", op.Name(), op.ID())
		return ErrOpExecBlockingUserFailure
	}
	return errors.Newf("Internal error: Unsupported failure type %s", failureType)
}

func execute(
	ctx context.Context,
	dag dag.WorkflowDag,
	pollInterval time.Duration,
	timeout time.Duration,
	cleanupTimeout time.Duration,
	jobManager job.JobManager,
	shouldPersistResults bool,
) error {
	// These are the operators of immediate interest. They either need to be scheduled or polled on.
	inProgressOps := make(map[uuid.UUID]operator2.Operator, len(dag.Operators()))
	completedOps := make(map[uuid.UUID]operator2.Operator, len(dag.Operators()))

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
	defer waitForInProgressOperators(ctx, inProgressOps, pollInterval, cleanupTimeout)

	start := time.Now()

	// TODO(kenxu): documentation
	for len(inProgressOps) > 0 {
		if time.Since(start) > timeout {
			return errors.New("Reached timeout waiting for workflow to complete.")
		}

		for _, op := range inProgressOps {
			execState, err := op.GetExecState(ctx)
			if err != nil {
				return err
			}

			if execState.Status == shared.PendingExecutionStatus {
				err = scheduler.ScheduleOperator(ctx, op, jobManager)
				if err != nil {
					return errors.Wrapf(err, "Unable to schedule operator %s.", op.Name())
				}
			} else if execState.Status == shared.RunningExecutionStatus {
				continue
			} else if execState.Status != shared.FailedExecutionStatus && execState.Status != shared.SucceededExecutionStatus {
				return errors.Newf("Internal error: a scheduled operator has unsupported status %s", execState.Status)
			}

			// The operator must have finished executing and is in either a success or failed state.
			if shouldPersistResults {
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
					// Before scheduling the next operator, check that all upstream artifacts to that operator
					// have been computed.
					if !nextOp.Ready(ctx) {
						continue
					}

					// Defensive check: do not reschedule an already in-progress operator. This shouldn't actually
					// matter because we only keep and update a single copy an on operator.
					if _, ok := inProgressOps[nextOp.ID()]; !ok {
						inProgressOps[nextOp.ID()] = nextOp
					}
				}
			}

			time.Sleep(pollInterval)
		}
	}

	if len(completedOps) != len(dag.Operators()) {
		return errors.Newf(fmt.Sprintf("Internal error: %d operators were provided but only %d completed.", len(dag.Operators()), len(completedOps)))
	}
	return nil
}

func Preview(
	ctx context.Context,
	dag *workflow_dag.DBWorkflowDag,
	workflowStoragePaths *utils.WorkflowStoragePaths,
	pollIntervalMillisec time.Duration,
	jobManager job.JobManager,
	vaultObject vault.Vault,
) (shared.ExecutionStatus, error) {
	return deprecatedOrchestrate(
		ctx,
		dag,
		workflowStoragePaths,
		pollIntervalMillisec,
		workflow.NewNoopReader(true),
		workflow_dag_result.NewNoopWriter(true),
		operator_result.NewNoopWriter(true),
		artifact_result.NewNoopWriter(true),
		notification.NewNoopWriter(true),
		user.NewNoopReader(true),
		database.NewNoopDatabase(),
		jobManager,
		vaultObject,
		true,
	)
}

func Execute(
	ctx context.Context,
	dag *workflow_dag.DBWorkflowDag,
	workflowStoragePaths *utils.WorkflowStoragePaths,
	pollIntervalMillisec time.Duration,
	workflowReader workflow.Reader,
	workflowDagResultWriter workflow_dag_result.Writer,
	operatorResultWriter operator_result.Writer,
	artifactResultWriter artifact_result.Writer,
	notificationWriter notification.Writer,
	userReader user.Reader,
	db database.Database,
	jobManager job.JobManager,
	vaultObject vault.Vault,
) (shared.ExecutionStatus, error) {
	return deprecatedOrchestrate(
		ctx,
		dag,
		workflowStoragePaths,
		pollIntervalMillisec,
		workflowReader,
		workflowDagResultWriter,
		operatorResultWriter,
		artifactResultWriter,
		notificationWriter,
		userReader,
		db,
		jobManager,
		vaultObject,
		false,
	)
}
