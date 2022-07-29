package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/aqueducthq/aqueduct/lib/cronjob"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/dag"
	dag_utils "github.com/aqueducthq/aqueduct/lib/workflow/dag"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type AqueductTimeConfig struct {
	// Configures exactly long we wait before polling again on an in-progress operator.
	OperatorPollInterval time.Duration

	// Configures the maximum amount of time we wait for execution to finish before aborting the run.
	ExecTimeout time.Duration

	// Configures the maximum amount of time we want for any leftover, in-progress operators to complete,
	// after execution has already finished. Once this time is exceeded, we'll give up.
	CleanupTimeout time.Duration
}

type aqEngine struct {
	database             database.Database
	githubManager        github.Manager
	vault                vault.Vault
	jobManager           job.JobManager
	storageConfig        *shared.StorageConfig
	timeConfig           *AqueductTimeConfig
	shouldPersistResults bool

	//Writers and readers needed for workflow execution
	workflowDagResultWriter workflow_dag_result.Writer
	operatorResultWriter    operator_result.Writer
	artifactResultWriter    artifact_result.Writer
	notificationWriter      notification.Writer
	workflowReader          workflow.Reader
	userReader              user.Reader
}

type workflowRunMetadata struct {
	// Maps every operator to the number of its immediate dependencies
	// that still needs to be computed. When this hits 0 during execution,
	// then the operator is ready to be scheduled.
	opToDependencyCount map[uuid.UUID]int
	inProgressOps       map[uuid.UUID]operator.Operator
	completedOps        map[uuid.UUID]operator.Operator
	status              shared.ExecutionStatus
}

func NewAqEngine(
	database database.Database,
	githubManager github.Manager,
	vault vault.Vault,
	jobManager job.JobManager,
	storageConfig *shared.StorageConfig,
	timeConfig AqueductTimeConfig,
	shouldPersistResults bool,
	workflowDagResultWriter workflow_dag_result.Writer,
	operatorResultWriter operator_result.Writer,
	artifactResultWriter artifact_result.Writer,
	notificationWriter notification.Writer,
	workflowReader workflow.Reader,
	userReader user.Reader,
) (Engine, error) {

	return &aqEngine{
		database:                database,
		githubManager:           githubManager,
		vault:                   vault,
		workflowDagResultWriter: workflowDagResultWriter,
		operatorResultWriter:    operatorResultWriter,
		artifactResultWriter:    artifactResultWriter,
		notificationWriter:      notificationWriter,
		workflowReader:          workflowReader,
		userReader:              userReader,
		storageConfig:           storageConfig,
		timeConfig:              &timeConfig,
		shouldPersistResults:    shouldPersistResults,
	}, nil
}

func (eng *aqEngine) CleanupWorkflow(ctx context.Context, workflowDag dag.WorkflowDag) {
	for _, op := range workflowDag.Operators() {
		op.Finish(ctx)
	}
}

func (eng *aqEngine) ScheduleWorkflow(ctx context.Context, dbWorkflowDag *workflow_dag.DBWorkflowDag, name string, period string) error {

	cronjobManager := cronjob.NewProcessCronjobManager()
	err := cronjobManager.DeployCronJob(
		ctx,
		name,
		period,
		eng.generateWorkflowCronFunction(context.Background(), name, dbWorkflowDag),
	)
	if err != nil {
		return errors.Wrap(err, "Unable to schedule workflow.")
	}
	return nil

	// spec := job.NewWorkflowSpec(
	// 	name,
	// 	workflowId,
	// 	eng.database.Config(),
	// 	eng.vault.Config(),
	// 	eng.jobManager.Config(),
	// 	eng.githubManager.Config(),
	// 	nil, /* parameters */
	// )

	// // Note: Change implementation of Schedule to not rely on JobManager
	// err := eng.jobManager.DeployCronJob(
	// 	ctx,
	// 	name,
	// 	period,
	// 	spec,
	// )
	// if err != nil {
	// 	return errors.Wrap(err, "Unable to schedule workflow.")
	// }

	// return nil
}

func (eng *aqEngine) SyncWorkflow(ctx context.Context, dbWorkflowDag *workflow_dag.DBWorkflowDag) {}

func (eng *aqEngine) ExecuteWorkflow(
	ctx context.Context,
	dbWorkflowDag *workflow_dag.DBWorkflowDag,
) (shared.ExecutionStatus, error) {

	dag, err := dag_utils.NewWorkflowDag(
		ctx,
		dbWorkflowDag,
		eng.workflowDagResultWriter,
		eng.operatorResultWriter,
		eng.artifactResultWriter,
		eng.workflowReader,
		eng.notificationWriter,
		eng.userReader,
		eng.jobManager,
		eng.vault,
		eng.storageConfig,
		eng.database,
	)
	if err != nil {
		return shared.FailedExecutionStatus, errors.Wrap(err, "Unable to create NewWorkflowDag.")
	}

	opToDependencyCount := make(map[uuid.UUID]int, len(dag.Operators()))
	for _, op := range dag.Operators() {
		inputs, err := dag.OperatorInputs(op)
		if err != nil {
			return shared.FailedExecutionStatus, errors.Wrap(err, "Unable to initialize operator inputs.")
		}
		opToDependencyCount[op.ID()] = len(inputs)
	}

	workflowRunMetadata := &workflowRunMetadata{
		opToDependencyCount: opToDependencyCount,
		inProgressOps:       make(map[uuid.UUID]operator.Operator, len(dag.Operators())),
		completedOps:        make(map[uuid.UUID]operator.Operator, len(dag.Operators())),
		status:              shared.PendingExecutionStatus,
	}

	if eng.shouldPersistResults {
		err := dag.InitializeResults(ctx)
		if err != nil {
			return shared.FailedExecutionStatus, errors.Wrap(err, "Unable to initialize dag results.")
		}

		// Make sure to persist the dag results on exit.
		defer func() {
			err = dag.PersistResult(ctx, workflowRunMetadata.status)
			if err != nil {
				log.Errorf("Error when persisting dag resutls: %v", err)
			}
		}()
	}

	workflowRunMetadata.status = shared.RunningExecutionStatus
	err = eng.execute(
		ctx,
		dag,
		workflowRunMetadata,
		eng.timeConfig,
		eng.jobManager,
		eng.shouldPersistResults,
	)
	if err != nil {
		workflowRunMetadata.status = shared.FailedExecutionStatus
	} else {
		workflowRunMetadata.status = shared.SucceededExecutionStatus
	}
	return workflowRunMetadata.status, err
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
			// since this is called after engestration exits.
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
		log.Errorf("Failed due to user error. Operator name %s, id %s.", op.Name(), op.ID())
		return ErrOpExecBlockingUserFailure
	}
	return errors.Newf("Internal error: Unsupported failure type %v", failureType)
}

func (eng *aqEngine) execute(
	ctx context.Context,
	workflowDag dag.WorkflowDag,
	workflowRunMetadata *workflowRunMetadata,
	timeConfig *AqueductTimeConfig,
	jobManager job.JobManager,
	shouldPersistResults bool,
) error {
	// These are the operators of immediate interest. They either need to be scheduled or polled on.
	inProgressOps := workflowRunMetadata.inProgressOps
	completedOps := workflowRunMetadata.completedOps
	dag := workflowDag
	opToDependencyCount := workflowRunMetadata.opToDependencyCount

	// Kick off execution by starting all operators that don't have any inputs.
	for _, op := range dag.Operators() {
		if opToDependencyCount[op.ID()] == 0 {
			inProgressOps[op.ID()] = op
		}
	}

	if len(inProgressOps) == 0 {
		return errors.Newf("No initial operators to schedule.")
	}

	// Wait a little bit for all active operators to finish before exiting on failure.
	defer waitForInProgressOperators(ctx, inProgressOps, timeConfig.OperatorPollInterval, timeConfig.CleanupTimeout)

	start := time.Now()

	for len(inProgressOps) > 0 {
		if time.Since(start) > timeConfig.ExecTimeout {
			return errors.New("Reached timeout waiting for workflow to complete.")
		}

		for _, op := range inProgressOps {
			execState, err := op.GetExecState(ctx)
			if err != nil {
				return err
			}

			if execState.Status == shared.PendingExecutionStatus {
				spec := op.JobSpec()
				err = jobManager.Launch(ctx, spec.JobName(), spec)
				if err != nil {
					return errors.Wrapf(err, "Unable to schedule operator %s.", op.Name())
				}
				continue
			} else if execState.Status == shared.RunningExecutionStatus {
				continue
			}
			if execState.Status != shared.FailedExecutionStatus && execState.Status != shared.SucceededExecutionStatus {
				return errors.Newf("Internal error: the operator is expected to have terminated, but instead has status %s", execState.Status)
			}

			// From here on we can assume that the operator has terminated.
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

			outputArtifacts, err := dag.OperatorOutputs(op)
			if err != nil {
				return err
			}
			for _, outputArtifact := range outputArtifacts {
				nextOps, err := dag.OperatorsOnArtifact(outputArtifact)
				if err != nil {
					return err
				}

				for _, nextOp := range nextOps {
					// Decrement the active dependency count for every downstream operator.
					// Once this count reaches zero, we can schedule the next operator.
					opToDependencyCount[nextOp.ID()] -= 1

					if opToDependencyCount[nextOp.ID()] < 0 {
						return errors.Newf("Internal error: operator %s has a negative dependnecy count.", op.Name())
					}

					if opToDependencyCount[nextOp.ID()] == 0 {
						// Defensive check: do not reschedule an already in-progress operator. This shouldn't actually
						// matter because we only keep and update a single copy an on operator.
						if _, ok := inProgressOps[nextOp.ID()]; !ok {
							inProgressOps[nextOp.ID()] = nextOp
						}
					}
				}
			}

			time.Sleep(timeConfig.OperatorPollInterval)
		}
	}

	if len(completedOps) != len(dag.Operators()) {
		return errors.Newf("Internal error: %d operators were provided but only %d completed.", len(dag.Operators()), len(completedOps))
	}

	for opID, depCount := range opToDependencyCount {
		if depCount != 0 {
			return errors.Newf("Internal error: operator %s has a non-zero dep count %d.", opID, depCount)
		}
	}
	return nil
}

func (eng *aqEngine) generateWorkflowCronFunction(ctx context.Context, name string, dbWorkflowDag *workflow_dag.DBWorkflowDag) func() {
	return func() {
		jobName := fmt.Sprintf("%s-%d", name, time.Now().Unix())
		_, err := eng.ExecuteWorkflow(ctx, dbWorkflowDag)
		if err != nil {
			log.Errorf("Error running cron job %s: %v", jobName, err)
		} else {
			log.Infof("Launched cron job %s", jobName)
		}
	}
}
