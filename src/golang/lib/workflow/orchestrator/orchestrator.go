package orchestrator

import (
	"context"
	"fmt"
	dag "github.com/aqueducthq/aqueduct/lib/workflow"
	operator2 "github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
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
			execState, err := op.GetExecState()

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
		if op.Ready() {
			inProgressOps[op.ID()] = op
		}
	}

	if len(inProgressOps) == 0 {
		return errors.Newf("No initial operators to schedule.")
	}

	// Wait a little bit for all active operators to finish before exiting on failure.
	defer waitForInProgressOperators(inProgressOps, pollInterval, cleanupTimeout)

	start := time.Now()

	// TODO(kenxu): documentation
	for len(inProgressOps) > 0 {
		if time.Since(start) > timeout {
			return errors.New("Reached timeout waiting for workflow to complete.")
		}

		for _, op := range inProgressOps {
			execState, err := op.GetExecState()
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
				err = op.PersistResult()
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
					if !nextOp.Ready() {
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
		return errors.Newf(fmt.Sprintf("Internal error: %d operators were provided but only %d completed.", len(dag.Operators), len(completedOps)))
	}
	return nil
}

func initializeOrchestration(
	operators map[uuid.UUID]operator.DBOperator,
	ready map[uuid.UUID]bool,
	operatorDependencies map[uuid.UUID]map[uuid.UUID]bool,
	artifactToDownstreamOperatorIds map[uuid.UUID][]uuid.UUID,
) {
	// This block does an initial scan of all operators and set the initial state for the orchestration.
	// Input data structures should be empty and will be updated throughout initialization:
	// - set upstream artifact counts for all operators
	// - put all operators without any upstream to 'ready to schedule' set
	// - set downstream operator list for all artifacts
	for id, operator := range operators {
		operatorDependencies[id] = make(map[uuid.UUID]bool, len(operator.Inputs))
		for _, artifactId := range operator.Inputs {
			operatorDependencies[id][artifactId] = true
		}

		if len(operator.Inputs) == 0 {
			ready[id] = true
		}

		for _, artifactId := range operator.Inputs {
			downstreamOps, ok := artifactToDownstreamOperatorIds[artifactId]
			if !ok {
				downstreamOps = make([]uuid.UUID, 0, len(operators))
				artifactToDownstreamOperatorIds[artifactId] = downstreamOps
			}

			artifactToDownstreamOperatorIds[artifactId] = append(downstreamOps, id)
		}
	}
}

// `updateCompletedOp` checks the status of actively running operators and updates
// their status according to the execution result. It returns a bool indicating
// whether the workflow execution should stop due to an error in the user code, and
// any internal system error occurred during the execution.
func updateCompletedOp(
	ctx context.Context,
	operators map[uuid.UUID]operator.DBOperator,
	ready map[uuid.UUID]bool,
	active map[uuid.UUID]bool,
	operatorDependencies map[uuid.UUID]map[uuid.UUID]bool,
	artifactToDownstreamOperatorIds map[uuid.UUID][]uuid.UUID,
	operatorIdToJobId map[uuid.UUID]string,
	storageConfig *shared.StorageConfig,
	artifactMetadataPaths map[uuid.UUID]string,
	operatorMetadataPaths map[uuid.UUID]string,
	operatorToOperatorResult map[uuid.UUID]uuid.UUID,
	artifactToArtifactResult map[uuid.UUID]uuid.UUID,
	operatorResultWriter operator_result.Writer,
	artifactResultWriter artifact_result.Writer,
	db database.Database,
	jobManager job.JobManager,
	isPreview bool,
) error {
	completedIds := make([]uuid.UUID, 0, len(active))
	for id := range active {
		jobStatus, err := jobManager.Poll(ctx, operatorIdToJobId[id])
		if err != nil {
			return err
		}

		if jobStatus != shared.PendingExecutionStatus {
			op, ok := operators[id]
			if !ok {
				return errors.Newf("Error during executino, invalid operator ID %s", id)
			}

			execStatus := scheduler.CheckOperatorExecutionStatus(
				ctx,
				storageConfig,
				operatorMetadataPaths[op.Id],
			)

			if !isPreview {
				utils.UpdateOperatorAndArtifactResults(
					ctx,
					&op,
					storageConfig,
					execStatus,
					artifactMetadataPaths,
					operatorToOperatorResult,
					artifactToArtifactResult,
					operatorResultWriter,
					artifactResultWriter,
					db,
				)
			}

			if execStatus.Status == shared.FailedExecutionStatus && *execStatus.FailureType == shared.SystemFailure {
				return ErrOpExecSystemFailure
			}

			if execStatus.Status == shared.FailedExecutionStatus && *execStatus.FailureType == shared.UserFailure {
				// There is an user error when executing the operator.
				log.Errorf("Failed due to user error. Operator name %s, id %s", op.Name, op.Id)
				return ErrOpExecBlockingUserFailure
			}

			completedIds = append(completedIds, id)

			for _, artifactId := range op.Outputs {
				if downstreampOps, ok := artifactToDownstreamOperatorIds[artifactId]; ok {
					for _, downstreamOpId := range downstreampOps {
						if _, ok := operatorDependencies[downstreamOpId][artifactId]; !ok {
							return ErrInvalidArtifactId
						}

						delete(operatorDependencies[downstreamOpId], artifactId)
						if len(operatorDependencies[downstreamOpId]) == 0 {
							ready[downstreamOpId] = true
						}
					}
				}
			}
		}
	}

	for _, id := range completedIds {
		delete(active, id)
	}

	return nil
}

func scheduleOperators(
	ctx context.Context,
	operators map[uuid.UUID]operator.DBOperator,
	artifacts map[uuid.UUID]artifact.DBArtifact,
	ready map[uuid.UUID]bool,
	active map[uuid.UUID]bool,
	operatorIdToJobId map[uuid.UUID]string,
	storageConfig *shared.StorageConfig,
	artifactContentPaths map[uuid.UUID]string,
	artifactMetadataPaths map[uuid.UUID]string,
	operatorMetadataPaths map[uuid.UUID]string,
	jobManager job.JobManager,
	vaultObject vault.Vault,
) error {
	for id := range ready {
		op, ok := operators[id]
		if !ok {
			return ErrInvalidOpId
		}

		operatorMetadataPath, ok := operatorMetadataPaths[id]
		if !ok {
			return ErrInvalidOpId
		}

		inputArtifacts := make([]artifact.DBArtifact, 0, len(op.Inputs))
		inputContentPaths := make([]string, 0, len(op.Inputs))
		inputMetadataPaths := make([]string, 0, len(op.Inputs))
		for _, inputArtifactId := range op.Inputs {
			inputArtifact, ok := artifacts[inputArtifactId]
			if !ok {
				return errors.Newf("Cannot find artifact with ID %v", inputArtifactId)
			}

			inputArtifacts = append(inputArtifacts, inputArtifact)
			inputContentPaths = append(inputContentPaths, artifactContentPaths[inputArtifact.Id])
			inputMetadataPaths = append(inputMetadataPaths, artifactMetadataPaths[inputArtifact.Id])
		}

		outputArtifacts := make([]artifact.DBArtifact, 0, len(op.Outputs))
		outputContentPaths := make([]string, 0, len(op.Outputs))
		outputMetadataPaths := make([]string, 0, len(op.Outputs))
		for _, outputArtifactId := range op.Outputs {
			outputArtifact, ok := artifacts[outputArtifactId]
			if !ok {
				return errors.Newf("Cannot find artifact with ID %v", outputArtifactId)
			}

			outputArtifacts = append(outputArtifacts, outputArtifact)
			outputContentPaths = append(outputContentPaths, artifactContentPaths[outputArtifact.Id])
			outputMetadataPaths = append(outputMetadataPaths, artifactMetadataPaths[outputArtifact.Id])
		}

		jobId, err := scheduler.ScheduleOperator(
			ctx,
			op,
			inputArtifacts,
			outputArtifacts,
			operatorMetadataPath,
			inputContentPaths,
			inputMetadataPaths,
			outputContentPaths,
			outputMetadataPaths,
			storageConfig,
			jobManager,
			vaultObject,
		)
		if err != nil {
			return err
		}

		active[id] = true
		operatorIdToJobId[id] = jobId
	}

	return nil
}

// `waitForActiveOperators` is deferred till the end of `Orchestrate`.
// If the workflow is successfully executed, this should return immediately
// as `active` would be empty. But if the workflow failed in the middle,
// there may be some in-progress operators that we want to wait till they finish.
// After they finish, we can then perform cleanup operations such as clearing
// storage files.
func waitForActiveOperators(
	ctx context.Context,
	active map[uuid.UUID]bool,
	operatorIdToJobId map[uuid.UUID]string,
	jobManager job.JobManager,
) {
	for len(active) != 0 {
		completedIds := make([]uuid.UUID, 0, len(active))
		for id := range active {
			jobStatus, err := jobManager.Poll(ctx, operatorIdToJobId[id])
			if err != nil {
				// If the err is job doesn't exist, then it means it already finished and was garbage-collected.
				if err != job.ErrJobNotExist {
					log.Errorf(
						"Unexpected error occurred when checking the job status during failed workflow cleanup. %v",
						err,
					)
					return
				}
			}

			if jobStatus != shared.PendingExecutionStatus {
				completedIds = append(completedIds, id)
			}
		}

		for _, id := range completedIds {
			delete(active, id)
		}
	}
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

func deprecatedOrchestrate(
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
	isPreview bool,
) (shared.ExecutionStatus, error) {
	numOperators := len(dag.Operators)
	artifactToDownstreamOperatorIds := make(map[uuid.UUID][]uuid.UUID, len(dag.Artifacts))
	operatorIdToJobId := make(map[uuid.UUID]string, numOperators)
	// Maps from operator ID to its upstream artifact dependencies.
	operatorDependencies := make(map[uuid.UUID]map[uuid.UUID]bool, numOperators)
	ready := make(map[uuid.UUID]bool, numOperators)
	active := make(map[uuid.UUID]bool, numOperators)

	defer func() {
		waitForActiveOperators(ctx, active, operatorIdToJobId, jobManager)
	}()

	initializeOrchestration(
		dag.Operators,
		ready,
		operatorDependencies,
		artifactToDownstreamOperatorIds,
	)

	status := shared.FailedExecutionStatus

	// These are only relevant for non-preview execution.
	var workflowDagResultId uuid.UUID
	operatorToOperatorResult := make(map[uuid.UUID]uuid.UUID, len(dag.Operators))
	artifactToArtifactResult := make(map[uuid.UUID]uuid.UUID, len(dag.Artifacts))

	if !isPreview {
		// First, we create a database record of workflow dag result and set its status to `pending`.
		// TODO: wrap these writes into a transaction.
		// eng-599-adding-transaction-support-to-our-database-reader-and-writer
		workflowDagResult, err := workflowDagResultWriter.CreateWorkflowDagResult(ctx, dag.Id, db)
		if err != nil {
			return shared.FailedExecutionStatus, errors.Wrap(err, "Unable to create workflow dag result record.")
		}

		workflowDagResultId = workflowDagResult.Id

		defer func() {
			// We `defer` this call to ensure that the WorkflowDagResult metadata is always updated.
			utils.UpdateWorkflowDagResultMetadata(
				ctx,
				workflowDagResultId,
				status,
				workflowDagResultWriter,
				workflowReader,
				notificationWriter,
				userReader,
				db,
			)
		}()

		// Initialize all operator results and artifact results.
		for operatorId := range dag.Operators {
			operatorResult, err := operatorResultWriter.CreateOperatorResult(
				ctx,
				workflowDagResultId,
				operatorId,
				db,
			)
			if err != nil {
				return shared.FailedExecutionStatus, errors.Wrap(err, "Failed to create operator result record.")
			}

			operatorToOperatorResult[operatorId] = operatorResult.Id
		}

		for artifactId := range dag.Artifacts {
			storagePath := workflowStoragePaths.ArtifactPaths[artifactId]
			artifactResult, err := artifactResultWriter.CreateArtifactResult(
				ctx,
				workflowDagResultId,
				artifactId,
				storagePath,
				db,
			)
			if err != nil {
				return shared.FailedExecutionStatus, errors.Wrap(err, "Failed to create artifact result record.")
			}

			artifactToArtifactResult[artifactId] = artifactResult.Id
		}
	}

	start := time.Now()

	// We keep orchestrating while there's any active or ready-to-schedule operators.
	// While such case, we do the following:
	// - poll the state of all active operators
	// - if any operator is completed, we:
	//   - for each of output artifacts, decrement their downstream operators' 'upstream artifact' count by 1
	//   - if any operator's upstream count become 0, mark it as 'ready'
	// - remove all completed operator from the list of active ones
	// - schedule all operators in ready list
	//   - mark each of them as active
	//   - clear the ready list after all operators are scheduled
	for len(ready) > 0 || len(active) > 0 {
		if time.Since(start) > defaultTimeout {
			return shared.FailedExecutionStatus, errors.New("Reached timeout waiting for workflow to complete.")
		}

		err := updateCompletedOp(
			ctx,
			dag.Operators,
			ready,
			active,
			operatorDependencies,
			artifactToDownstreamOperatorIds,
			operatorIdToJobId,
			&dag.StorageConfig,
			workflowStoragePaths.ArtifactMetadataPaths,
			workflowStoragePaths.OperatorMetadataPaths,
			operatorToOperatorResult,
			artifactToArtifactResult,
			operatorResultWriter,
			artifactResultWriter,
			db,
			jobManager,
			isPreview,
		)
		if err != nil {
			return shared.FailedExecutionStatus, err
		}

		// Schedule all operators in ready state.
		err = scheduleOperators(
			ctx,
			dag.Operators,
			dag.Artifacts,
			ready,
			active,
			operatorIdToJobId,
			&dag.StorageConfig,
			workflowStoragePaths.ArtifactPaths,
			workflowStoragePaths.ArtifactMetadataPaths,
			workflowStoragePaths.OperatorMetadataPaths,
			jobManager,
			vaultObject,
		)
		if err != nil {
			return shared.FailedExecutionStatus, err
		}

		ready = map[uuid.UUID]bool{}

		time.Sleep(pollIntervalMillisec)
	}

	status = shared.SucceededExecutionStatus

	return status, nil
}
