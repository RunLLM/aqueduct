package engine

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"time"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/airflow"
	artifact_db "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/param"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/cronjob"
	"github.com/aqueducthq/aqueduct/lib/database"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/job"
	shared_utils "github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	mdl_shared "github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	dag_utils "github.com/aqueducthq/aqueduct/lib/workflow/dag"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/aqueducthq/aqueduct/lib/workflow/preview_cache"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	workflow_utils "github.com/aqueducthq/aqueduct/lib/workflow/utils"
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

// Repos contains the repos needed by the Engine
type Repos struct {
	ArtifactRepo             repos.Artifact
	ArtifactResultRepo       repos.ArtifactResult
	DAGRepo                  repos.DAG
	DAGEdgeRepo              repos.DAGEdge
	DAGResultRepo            repos.DAGResult
	ExecutionEnvironmentRepo repos.ExecutionEnvironment
	IntegrationRepo          repos.Integration
	NotificationRepo         repos.Notification
	OperatorRepo             repos.Operator
	OperatorResultRepo       repos.OperatorResult
	WatcherRepo              repos.Watcher
	WorkflowRepo             repos.Workflow
}

type aqEngine struct {
	Database       database.Database
	GithubManager  github.Manager
	CronjobManager cronjob.CronjobManager
	AqPath         string

	// Only used for previews.
	PreviewCacheManager preview_cache.CacheManager

	*Repos
}

type WorkflowRunMetadata struct {
	// Maps every operator to the number of its immediate dependencies
	// that still needs to be computed. When this hits 0 during execution,
	// then the operator is ready to be scheduled.
	OpToDependencyCount map[uuid.UUID]int
	InProgressOps       map[uuid.UUID]operator.Operator
	CompletedOps        map[uuid.UUID]operator.Operator
	Status              shared.ExecutionStatus
}

type WorkflowPreviewResult struct {
	Status    shared.ExecutionStatus
	Operators map[uuid.UUID]shared.ExecutionState
	Artifacts map[uuid.UUID]PreviewArtifactResult
}

type PreviewArtifactResult struct {
	SerializationType artifact_result.SerializationType `json:"serialization_type"`
	ArtifactType      artifact_db.Type                  `json:"artifact_type"`
	Content           []byte                            `json:"content"`
}

func NewAqEngine(
	database database.Database,
	githubManager github.Manager,
	previewCacheManager preview_cache.CacheManager,
	aqPath string,
	repos *Repos,
) (*aqEngine, error) {
	cronjobManager := cronjob.NewProcessCronjobManager()

	return &aqEngine{
		Database:            database,
		GithubManager:       githubManager,
		PreviewCacheManager: previewCacheManager,
		CronjobManager:      cronjobManager,
		AqPath:              aqPath,
		Repos:               repos,
	}, nil
}

// TODO ENG-1444: Remove jobSpec/ creation once we get rid of executor
func (eng *aqEngine) ScheduleWorkflow(
	ctx context.Context,
	workflowId uuid.UUID,
	name string,
	period string,
) error {
	jobSpec := job.NewWorkflowSpec(
		name,
		workflowId.String(),
		eng.Database.Config(),
		&job.ProcessConfig{
			BinaryDir:          path.Join(eng.AqPath, job.BinaryDir),
			OperatorStorageDir: path.Join(eng.AqPath, job.OperatorStorageDir),
		},
		eng.GithubManager.Config(),
		eng.AqPath,
		nil,
	)
	err := eng.CronjobManager.DeployCronJob(
		ctx,
		name,
		period,
		eng.generateCronFunction(name, jobSpec),
	)
	if err != nil {
		return errors.Wrap(err, "Unable to schedule workflow.")
	}
	return nil
}

func (eng *aqEngine) ExecuteWorkflow(
	ctx context.Context,
	workflowID uuid.UUID,
	timeConfig *AqueductTimeConfig,
	parameters map[string]param.Param,
) (shared.ExecutionStatus, error) {
	dbDAG, err := workflow_utils.ReadLatestDAGFromDatabase(
		ctx,
		workflowID,
		eng.WorkflowRepo,
		eng.DAGRepo,
		eng.OperatorRepo,
		eng.ArtifactRepo,
		eng.DAGEdgeRepo,
		eng.Database,
	)
	if err != nil {
		return shared.FailedExecutionStatus, errors.Wrap(err, "Error reading latest workflowDag.")
	}

	pendingAt := time.Now()
	execState := &mdl_shared.ExecutionState{
		Status: mdl_shared.PendingExecutionStatus,
		Timestamps: &mdl_shared.ExecutionTimestamps{
			PendingAt: &pendingAt,
		},
	}

	dagResult, err := eng.DAGResultRepo.Create(
		ctx,
		dbDAG.ID,
		execState,
		eng.Database,
	)
	if err != nil {
		return shared.FailedExecutionStatus, errors.Wrap(err, "Error initializing workflowDagResult.")
	}

	// Any errors after this point should be persisted to the WorkflowDagResult created above
	defer func() {
		if err != nil {
			// Mark the workflow dag result as failed
			execState.Status = mdl_shared.FailedExecutionStatus
			now := time.Now()
			execState.Timestamps.FinishedAt = &now
		}

		if updateErr := workflow_utils.UpdateDAGResultMetadata(
			ctx,
			dagResult.ID,
			execState,
			eng.DAGResultRepo,
			eng.WorkflowRepo,
			eng.NotificationRepo,
			eng.Database,
		); updateErr != nil {
			log.Errorf("Unable to update DAGResult metadata for %v", dagResult.ID)
		}
	}()

	githubClient, err := eng.GithubManager.GetClient(ctx, dbDAG.Metadata.UserID)
	if err != nil {
		return shared.FailedExecutionStatus, errors.Wrap(err, "Error getting github client.")
	}

	dbDAG, err = workflow_utils.UpdateWorkflowDagToLatest(
		ctx,
		githubClient,
		dbDAG,
		eng.WorkflowRepo,
		eng.DAGRepo,
		eng.OperatorRepo,
		eng.DAGEdgeRepo,
		eng.ArtifactRepo,
		eng.Database,
	)
	if err != nil {
		return shared.FailedExecutionStatus, errors.Wrap(err, "Error updating workflowDag to latest.")
	}

	// Overwrite the parameter specs for all custom parameters defined by the user.
	for name, param := range parameters {
		var op *models.Operator
		for _, dagOp := range dbDAG.Operators {
			if dagOp.Name == name {
				op = &dagOp
				break
			}
		}

		if op == nil {
			continue
		}

		if !op.Spec.IsParam() {
			return shared.FailedExecutionStatus, errors.Wrap(err, "Cannot set parameters on a non-parameter operator.")
		}
		dbDAG.Operators[op.ID].Spec.Param().Val = param.Val
		dbDAG.Operators[op.ID].Spec.Param().SerializationType = param.SerializationType
	}

	opIds := make([]uuid.UUID, 0, len(dbDAG.Operators))
	for _, op := range dbDAG.Operators {
		opIds = append(opIds, op.ID)
	}

	execEnvsByOpId, err := exec_env.GetActiveExecutionEnvironmentsByOperatorIDs(
		ctx,
		opIds,
		eng.ExecutionEnvironmentRepo,
		eng.Database,
	)
	if err != nil {
		return shared.FailedExecutionStatus, errors.Wrap(err, "Unable to read operator environments.")
	}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return shared.FailedExecutionStatus, errors.Wrap(err, "Unable to initialize vault.")
	}

	dag, err := dag_utils.NewWorkflowDag(
		ctx,
		dagResult.ID,
		dbDAG,
		eng.OperatorResultRepo,
		eng.ArtifactRepo,
		eng.ArtifactResultRepo,
		vaultObject,
		nil, /* artifactCacheManager */
		execEnvsByOpId,
		operator.Publish,
		eng.AqPath,
		eng.Database,
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

	wfRunMetadata := &WorkflowRunMetadata{
		OpToDependencyCount: opToDependencyCount,
		InProgressOps:       make(map[uuid.UUID]operator.Operator, len(dag.Operators())),
		CompletedOps:        make(map[uuid.UUID]operator.Operator, len(dag.Operators())),
	}

	err = dag.InitOpAndArtifactResults(ctx)
	if err != nil {
		return shared.FailedExecutionStatus, errors.Wrap(err, "Unable to initialize dag results.")
	}

	execState.Status = mdl_shared.RunningExecutionStatus
	runningAt := time.Now()
	execState.Timestamps.RunningAt = &runningAt

	err = eng.executeWithEngine(
		ctx,
		dag,
		dbDAG.Metadata.Name,
		dbDAG.EngineConfig,
		dbDAG.StorageConfig,
		wfRunMetadata,
		timeConfig,
		operator.Publish,
		vaultObject,
	)
	if err != nil {
		execState.Status = mdl_shared.FailedExecutionStatus
		now := time.Now()
		execState.Timestamps.FinishedAt = &now
		return shared.FailedExecutionStatus, errors.Wrapf(err, "Error executing workflow")
	} else {
		execState.Status = mdl_shared.SucceededExecutionStatus
		now := time.Now()
		execState.Timestamps.FinishedAt = &now
	}

	return shared.SucceededExecutionStatus, nil
}

func (eng *aqEngine) PreviewWorkflow(
	ctx context.Context,
	dbDAG *models.DAG,
	execEnvByOperatorId map[uuid.UUID]exec_env.ExecutionEnvironment,
	timeConfig *AqueductTimeConfig,
) (*WorkflowPreviewResult, error) {
	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return nil, errors.Wrap(err, "Unable to initialize vault.")
	}

	dag, err := dag_utils.NewWorkflowDag(
		ctx,
		uuid.Nil, /* workflowDagResultID */
		dbDAG,
		eng.OperatorResultRepo,
		eng.ArtifactRepo,
		eng.ArtifactResultRepo,
		vaultObject,
		eng.PreviewCacheManager,
		execEnvByOperatorId,
		operator.Preview,
		eng.AqPath,
		eng.Database,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create NewWorkflowDag.")
	}

	defer eng.cleanupWorkflow(ctx, dag)

	opToDependencyCount := make(map[uuid.UUID]int, len(dag.Operators()))
	for _, op := range dag.Operators() {
		inputs, err := dag.OperatorInputs(op)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to initialize operator inputs.")
		}
		opToDependencyCount[op.ID()] = len(inputs)
	}

	wfRunMetadata := &WorkflowRunMetadata{
		OpToDependencyCount: opToDependencyCount,
		InProgressOps:       make(map[uuid.UUID]operator.Operator, len(dag.Operators())),
		CompletedOps:        make(map[uuid.UUID]operator.Operator, len(dag.Operators())),
		Status:              shared.PendingExecutionStatus,
	}

	wfRunMetadata.Status = shared.RunningExecutionStatus
	err = eng.executeWithEngine(
		ctx,
		dag,
		fmt.Sprintf("PREVIEW_%s", uuid.New().String()),
		dbDAG.EngineConfig,
		dbDAG.StorageConfig,
		wfRunMetadata,
		timeConfig,
		operator.Preview,
		vaultObject,
	)
	if err != nil {
		log.Errorf("Workflow failed with error: %v", err)
		wfRunMetadata.Status = shared.FailedExecutionStatus
	} else {
		wfRunMetadata.Status = shared.SucceededExecutionStatus
	}

	execStateByOp := make(map[uuid.UUID]shared.ExecutionState, len(dag.Operators()))
	for _, op := range dag.Operators() {
		execState := op.ExecState()
		execStateByOp[op.ID()] = *execState
	}

	// Only include artifact results that were successfully computed.
	artifactResults := make(map[uuid.UUID]PreviewArtifactResult)
	for _, artf := range dag.Artifacts() {
		if artf.Computed(ctx) {
			artifact_metadata, err := artf.GetMetadata(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to get artifact metadata.")
			}

			content, err := artf.GetContent(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to get artifact content.")
			}
			artifactResults[artf.ID()] = PreviewArtifactResult{
				SerializationType: artifact_metadata.SerializationType,
				ArtifactType:      artifact_metadata.ArtifactType,
				Content:           content,
			}
		}
	}

	return &WorkflowPreviewResult{
		Status:    wfRunMetadata.Status,
		Operators: execStateByOp,
		Artifacts: artifactResults,
	}, nil
}

func (eng *aqEngine) DeleteWorkflow(
	ctx context.Context,
	workflowID uuid.UUID,
) error {
	txn, err := eng.Database.BeginTx(ctx)
	if err != nil {
		return errors.Wrap(err, "Unable to delete workflow.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	// We first retrieve all relevant records from the database.
	workflowObj, err := eng.WorkflowRepo.Get(ctx, workflowID, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while retrieving workflow.")
	}

	dagsToDelete, err := eng.DAGRepo.GetByWorkflow(ctx, workflowObj.ID, txn)
	if err != nil || len(dagsToDelete) == 0 {
		return errors.Wrap(err, "Unexpected error occurred while retrieving workflow dags.")
	}

	dagIDs := make([]uuid.UUID, 0, len(dagsToDelete))
	for _, dag := range dagsToDelete {
		dagIDs = append(dagIDs, dag.ID)
	}

	dagResultsToDelete, err := eng.DAGResultRepo.GetByWorkflow(ctx, workflowObj.ID, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while retrieving workflow dag results.")
	}

	dagResultIDs := make([]uuid.UUID, 0, len(dagResultsToDelete))
	for _, dagResult := range dagResultsToDelete {
		dagResultIDs = append(dagResultIDs, dagResult.ID)
	}

	dagEdgesToDelete, err := eng.DAGEdgeRepo.GetByDAGBatch(ctx, dagIDs, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while retrieving workflow dag edges.")
	}

	operatorIDs := make([]uuid.UUID, 0, len(dagEdgesToDelete))
	artifactIDs := make([]uuid.UUID, 0, len(dagEdgesToDelete))

	operatorIdMap := make(map[uuid.UUID]bool)
	artifactIDMap := make(map[uuid.UUID]bool)

	for _, dagEdge := range dagEdgesToDelete {
		var operatorId uuid.UUID
		var artifactID uuid.UUID

		if dagEdge.Type == mdl_shared.OperatorToArtifactDAGEdge {
			operatorId = dagEdge.FromID
			artifactID = dagEdge.ToID
		} else {
			operatorId = dagEdge.ToID
			artifactID = dagEdge.FromID
		}

		if _, ok := operatorIdMap[operatorId]; !ok {
			operatorIdMap[operatorId] = true
			operatorIDs = append(operatorIDs, operatorId)
		}

		if _, ok := artifactIDMap[artifactID]; !ok {
			artifactIDMap[artifactID] = true
			artifactIDs = append(artifactIDs, artifactID)
		}
	}

	operatorsToDelete, err := eng.OperatorRepo.GetBatch(ctx, operatorIDs, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while retrieving operators.")
	}

	operatorResultsToDelete, err := eng.OperatorResultRepo.GetByDAGResultBatch(
		ctx,
		dagResultIDs,
		txn,
	)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while retrieving operator results.")
	}

	operatorResultIDs := make([]uuid.UUID, 0, len(operatorResultsToDelete))
	for _, operatorResult := range operatorResultsToDelete {
		operatorResultIDs = append(operatorResultIDs, operatorResult.ID)
	}

	artifactResultsToDelete, err := eng.ArtifactResultRepo.GetByDAGResults(
		ctx,
		dagResultIDs,
		txn,
	)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while retrieving artifact results.")
	}

	artifactResultIDs := make([]uuid.UUID, 0, len(artifactResultsToDelete))
	for _, artifactResult := range artifactResultsToDelete {
		artifactResultIDs = append(artifactResultIDs, artifactResult.ID)
	}

	targetWorkflowIDs, err := eng.WorkflowRepo.GetTargets(ctx, workflowID, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while retrieving target workflows.")
	}

	// Start deleting database records.
	err = eng.WatcherRepo.DeleteByWorkflow(ctx, workflowID, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while deleting workflow watchers.")
	}

	err = eng.OperatorResultRepo.DeleteBatch(ctx, operatorResultIDs, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while deleting operator results.")
	}

	err = eng.ArtifactResultRepo.DeleteBatch(ctx, artifactResultIDs, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while deleting artifact results.")
	}

	err = eng.DAGResultRepo.DeleteBatch(ctx, dagResultIDs, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while deleting workflow dag results.")
	}

	err = eng.DAGEdgeRepo.DeleteByDAGBatch(ctx, dagIDs, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while deleting workflow dag edges.")
	}

	err = eng.OperatorRepo.DeleteBatch(ctx, operatorIDs, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while deleting operators.")
	}

	err = eng.ArtifactRepo.DeleteBatch(ctx, artifactIDs, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while deleting artifacts.")
	}

	err = eng.DAGRepo.DeleteBatch(ctx, dagIDs, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while deleting workflow dags.")
	}

	err = eng.WorkflowRepo.Delete(ctx, workflowObj.ID, txn)
	if err != nil {
		return errors.Wrap(err, "Unexpected error occurred while deleting workflow.")
	}

	// For each target workflow, update its schedule to have a ManualTrigger
	for _, targetWorkflowID := range targetWorkflowIDs {
		targetWorkflow, err := eng.WorkflowRepo.Get(ctx, targetWorkflowID, txn)
		if err != nil {
			return errors.Wrap(err, "Unexpected error occurred while retrieving target workflow.")
		}

		schedule := targetWorkflow.Schedule
		schedule.SourceID = uuid.Nil
		schedule.Trigger = workflow.ManualUpdateTrigger

		if _, err := eng.WorkflowRepo.Update(
			ctx,
			targetWorkflowID,
			map[string]interface{}{
				models.WorkflowSchedule: &schedule,
			},
			txn,
		); err != nil {
			return errors.Wrap(err, "Unexpected error occurred while updating target workflow trigger from cascading to manual.")
		}
	}

	if err := txn.Commit(ctx); err != nil {
		return errors.Wrap(err, "Failed to delete workflow.")
	}

	// Delete storage files (artifact content and function files)
	storagePaths := make([]string, 0, len(operatorIDs)+len(artifactResultIDs))
	for _, op := range operatorsToDelete {
		if op.Spec.IsFunction() || op.Spec.IsMetric() || op.Spec.IsCheck() {
			storagePaths = append(storagePaths, op.Spec.Function().StoragePath)
		}
	}

	for _, art := range artifactResultsToDelete {
		storagePaths = append(storagePaths, art.ContentPath)
	}

	// Note: for now we assume all workflow dags have the same storage config.
	// This assumption will stay true until we allow users to configure custom storage config to store stuff.
	storageConfig := dagsToDelete[0].StorageConfig
	for _, workflowDag := range dagsToDelete {
		if !reflect.DeepEqual(workflowDag.StorageConfig, storageConfig) {
			return errors.New("Workflow Dags have mismatching storage config.")
		}
	}

	workflow_utils.CleanupStorageFiles(ctx, &storageConfig, storagePaths)

	// Delete the cron job if it had one.
	if workflowObj.Schedule.CronSchedule != "" {
		cronjobName := shared_utils.AppendPrefix(workflowID.String())
		err = eng.CronjobManager.DeleteCronJob(ctx, cronjobName)
		if err != nil {
			return errors.Wrap(err, "Failed to delete workflow's cronjob.")
		}
	}
	return nil
}

func (eng *aqEngine) EditWorkflow(
	ctx context.Context,
	txn database.Database,
	workflowID uuid.UUID,
	workflowName string,
	workflowDescription string,
	schedule *workflow.Schedule,
	retentionPolicy *workflow.RetentionPolicy,
	notificationSettings *mdl_shared.NotificationSettings,
) error {
	changes := map[string]interface{}{}
	if workflowName != "" {
		changes[models.WorkflowName] = workflowName
	}

	if workflowDescription != "" {
		changes[models.WorkflowDescription] = workflowDescription
	}

	if retentionPolicy != nil {
		changes[models.WorkflowRetentionPolicy] = retentionPolicy
	}

	if notificationSettings != nil {
		changes[models.WorkflowNotificationSettings] = notificationSettings
	}

	if schedule.Trigger != "" {
		cronjobName := shared_utils.AppendPrefix(workflowID.String())
		err := eng.updateWorkflowSchedule(ctx, workflowID, cronjobName, schedule)
		if err != nil {
			return errors.Wrap(err, "Unable to update workflow schedule.")
		}
		changes[models.WorkflowSchedule] = schedule
	}

	_, err := eng.WorkflowRepo.Update(ctx, workflowID, changes, txn)
	if err != nil {
		return errors.Wrap(err, "Unable to update workflow.")
	}

	return nil
}

// TODO ENG-1444: This function is only used to trigger a Workflow.
// Remove once executor is done.
func (eng *aqEngine) TriggerWorkflow(
	ctx context.Context,
	workflowID uuid.UUID,
	name string,
	timeConfig *AqueductTimeConfig,
	parameters map[string]param.Param,
) (shared.ExecutionStatus, error) {
	dag, err := utils.ReadLatestDAGFromDatabase(
		ctx,
		workflowID,
		eng.WorkflowRepo,
		eng.DAGRepo,
		eng.OperatorRepo,
		eng.ArtifactRepo,
		eng.DAGEdgeRepo,
		eng.Database,
	)
	if err != nil {
		return shared.FailedExecutionStatus, err
	}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return shared.FailedExecutionStatus, errors.Wrap(err, "Unable to initialize vault.")
	}

	if dag.EngineConfig.Type == shared.AirflowEngineType {
		// This is an Airflow workflow so the executor binary is not used
		if err := airflow.TriggerWorkflow(ctx, dag, vaultObject); err != nil {
			return shared.FailedExecutionStatus, errors.Wrap(
				err,
				"Unable to trigger a new workflow run on Airflow",
			)
		}
		return shared.SucceededExecutionStatus, nil
	}

	jobManager, err := job.NewProcessJobManager(
		&job.ProcessConfig{
			BinaryDir:          path.Join(eng.AqPath, job.BinaryDir),
			OperatorStorageDir: path.Join(eng.AqPath, job.OperatorStorageDir),
		},
	)
	if err != nil {
		log.Errorf("Unable to create JobManager: %v", err)
	}

	jobSpec := job.NewWorkflowSpec(
		name,
		workflowID.String(),
		eng.Database.Config(),
		&job.ProcessConfig{
			BinaryDir:          path.Join(eng.AqPath, job.BinaryDir),
			OperatorStorageDir: path.Join(eng.AqPath, job.OperatorStorageDir),
		},
		eng.GithubManager.Config(),
		eng.AqPath,
		parameters,
	)

	jobName := fmt.Sprintf("%s-%d", name, time.Now().Unix())
	err = jobManager.Launch(context.Background(), jobName, jobSpec)
	if err != nil {
		log.Errorf("Error running job %s: %v", jobName, err)
		return shared.FailedExecutionStatus, errors.Wrap(err, "Error running job.")
	} else {
		log.Infof("Launched job %s", jobName)
		return shared.PendingExecutionStatus, nil
	}
}

func (eng *aqEngine) cleanupWorkflow(ctx context.Context, workflowDag dag_utils.WorkflowDag) {
	for _, op := range workflowDag.Operators() {
		op.Finish(ctx)
	}
}

func (eng *aqEngine) executeWithEngine(
	ctx context.Context,
	dag dag_utils.WorkflowDag,
	workflowName string,
	engineConfig shared.EngineConfig,
	storageConfig shared.StorageConfig,
	workflowRunMetadata *WorkflowRunMetadata,
	timeConfig *AqueductTimeConfig,
	opExecMode operator.ExecutionMode,
	vaultObject vault.Vault,
) error {
	switch engineConfig.Type {
	case shared.DatabricksEngineType:
		jobConfig, err := operator.GenerateJobManagerConfig(
			ctx,
			engineConfig,
			&storageConfig,
			eng.AqPath,
			vaultObject,
		)
		if err != nil {
			return errors.Wrap(err, "Unable to generate JobManagerConfig.")
		}

		jobManager, err := job.NewJobManager(jobConfig)
		if err != nil {
			return errors.Wrap(err, "Unable to create JobManager.")
		}

		databricksJobManager, ok := jobManager.(*job.DatabricksJobManager)
		if !ok {
			return errors.Wrap(err, "Unable to create DatabricksJobManager.")
		}

		return ExecuteDatabricks(
			ctx,
			dag,
			workflowName,
			workflowRunMetadata,
			timeConfig,
			opExecMode,
			databricksJobManager,
			vaultObject,
			eng.IntegrationRepo,
			eng.Database,
		)
	default:
		return eng.execute(
			ctx,
			dag,
			workflowRunMetadata,
			timeConfig,
			vaultObject,
			opExecMode,
		)
	}
}

func onFinishExecution(
	ctx context.Context,
	inProgressOps map[uuid.UUID]operator.Operator,
	pollInterval time.Duration,
	cleanupTimeout time.Duration,
	curErr error,
	notificationContent *notificationContentStruct,
	dag dag_utils.WorkflowDag,
	execMode operator.ExecutionMode,
	vaultObject vault.Vault,
	integrationRepo repos.Integration,
	DB database.Database,
) {
	// Wait a little bit for all active operators to finish before exiting on failure.
	waitForInProgressOperators(ctx, inProgressOps, pollInterval, cleanupTimeout)
	if curErr != nil && notificationContent == nil {
		notificationContent = &notificationContentStruct{
			level:      mdl_shared.ErrorNotificationLevel,
			contextMsg: curErr.Error(),
		}
	}

	// Send notifications
	if notificationContent != nil && execMode == operator.Publish {
		err := sendNotifications(
			ctx,
			dag,
			notificationContent,
			vaultObject,
			integrationRepo,
			DB,
		)
		if err != nil {
			log.Errorf("Error sending notifications: %s", err)
		}
	}
}

func (eng *aqEngine) execute(
	ctx context.Context,
	workflowDag dag_utils.WorkflowDag,
	workflowRunMetadata *WorkflowRunMetadata,
	timeConfig *AqueductTimeConfig,
	vaultObject vault.Vault,
	opExecMode operator.ExecutionMode,
) (err error) {
	// These are the operators of immediate interest. They either need to be scheduled or polled on.
	inProgressOps := workflowRunMetadata.InProgressOps
	completedOps := workflowRunMetadata.CompletedOps
	dag := workflowDag
	opToDependencyCount := workflowRunMetadata.OpToDependencyCount

	var notificationContent *notificationContentStruct = nil
	err = nil

	// Kick off execution by starting all operators that don't have any inputs.
	for _, op := range dag.Operators() {
		if opToDependencyCount[op.ID()] == 0 {
			inProgressOps[op.ID()] = op
		}
	}

	if len(inProgressOps) == 0 {
		return errors.Newf("No initial operators to schedule.")
	}

	defer func() {
		onFinishExecution(
			ctx,
			inProgressOps,
			timeConfig.OperatorPollInterval,
			timeConfig.CleanupTimeout,
			err,
			notificationContent,
			workflowDag,
			opExecMode,
			vaultObject,
			eng.IntegrationRepo,
			eng.Database,
		)
	}()

	start := time.Now()

	for len(inProgressOps) > 0 {
		if time.Since(start) > timeConfig.ExecTimeout {
			return errors.New("Reached timeout waiting for workflow to complete.")
		}

		for _, op := range inProgressOps {
			execState, err := op.Poll(ctx)
			if err != nil {
				return err
			}

			if execState.Status == shared.PendingExecutionStatus {
				err = op.Launch(ctx)
				if err != nil {
					return errors.Wrapf(err, "Unable to schedule operator %s.", op.Name())
				}
				continue
			} else if execState.Status == shared.RunningExecutionStatus {
				continue
			}

			if !execState.Terminated() {
				return errors.Newf("Internal error: the operator is expected to have terminated, but instead has status %s", execState.Status)
			}

			// From here on we can assume that the operator has terminated.
			if opExecMode == operator.Publish {
				err = op.PersistResult(ctx)
				if err != nil {
					return errors.Wrapf(err, "Error when finishing execution of operator %s", op.Name())
				}
			}

			// We can continue orchestration on non-fatal errors; currently, this only allows through succeeded operators
			// and check operators with warning severity.
			if shouldStopExecution(execState) {
				log.Infof("Stopping execution of operator %v", op.ID())
				for id, dagOp := range workflowDag.Operators() {
					log.Infof("Checking status of operator %v", id)
					// Skip if this operator has already been completed or is in progress.
					if _, ok := completedOps[id]; ok {
						continue
					}
					if _, ok := inProgressOps[id]; ok {
						continue
					}

					dagOp.Cancel()
					if opExecMode == operator.Publish {
						err = dagOp.PersistResult(ctx)
						if err != nil {
							return errors.Wrapf(err, "Error when finishing execution of operator %s", op.Name())
						}
					}
				}

				notificationCtxMsg := ""
				if execState.Error != nil {
					notificationCtxMsg = execState.Error.Message()
				}

				notificationContent = &notificationContentStruct{
					level:      mdl_shared.ErrorNotificationLevel,
					contextMsg: notificationCtxMsg,
				}

				return opFailureError(*execState.FailureType, op)
			} else if execState.Status == shared.FailedExecutionStatus {
				notificationCtxMsg := ""
				if execState.Error != nil {
					notificationCtxMsg = execState.Error.Message()
				}

				notificationContent = &notificationContentStruct{
					level:      mdl_shared.WarningNotificationLevel,
					contextMsg: notificationCtxMsg,
				}
			}

			// Add the operator to the completed stack, and remove it from the in-progress one.
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

	// avoid overriding an existing notification (in practice, this is a warning)
	if notificationContent == nil {
		notificationContent = &notificationContentStruct{
			level: mdl_shared.SuccessNotificationLevel,
		}
	}
	return nil
}

func (eng *aqEngine) generateCronFunction(name string, jobSpec job.Spec) func() {
	// TODO ENG-1444: Creating this process job manager just to
	// launch the executor which calls ExecuteWorkflow().
	// Replace with call to ExecuteWorkflow() once executor is removed.
	jobManager, err := job.NewProcessJobManager(
		&job.ProcessConfig{
			BinaryDir:          path.Join(eng.AqPath, job.BinaryDir),
			OperatorStorageDir: path.Join(eng.AqPath, job.OperatorStorageDir),
		},
	)
	if err != nil {
		log.Errorf("Unable to create JobManager: %v", err)
	}

	return func() {
		jobName := fmt.Sprintf("%s-%d", name, time.Now().Unix())
		err := jobManager.Launch(context.Background(), jobName, jobSpec)
		if err != nil {
			log.Errorf("Error running cron job %s: %v", jobName, err)
		} else {
			log.Infof("Launched cron job %s", jobName)
		}
	}
}

func (eng *aqEngine) updateWorkflowSchedule(
	ctx context.Context,
	workflowId uuid.UUID,
	cronjobName string,
	newSchedule *workflow.Schedule,
) error {
	// How we update the workflow schedule depends on whether a cron job already exists.
	// A manually triggered workflow does not have a cron job. If we're editing it to have a periodic
	// schedule, we'll need to create a new cron job.
	if !eng.CronjobManager.CronJobExists(ctx, cronjobName) {
		if newSchedule.CronSchedule != "" {

			err := eng.ScheduleWorkflow(
				ctx,
				workflowId,
				cronjobName,
				string(newSchedule.CronSchedule),
			)
			if err != nil {
				return errors.Wrap(err, "Unable to deploy new cron job.")
			}
		}
		// We will no-op if the workflow continues to be manually triggered.
	} else {
		// Here, we can blindly set the cron job to be paused without any other
		// modification. The pausedness of the workflow will be written to the
		// database by the changes map above, and `prepare` guarantees us that
		// if `Paused` is true, then the workflow type is `Periodic`, which in
		// turn means a schedule must be set.
		newCronSchedule := string(newSchedule.CronSchedule)
		if newSchedule.Paused {
			// The `EditCronJob` helper automatically pauses a workflow when
			// you set the cron job schedule to an empty string.
			newCronSchedule = ""
		}
		// TODO ENG-1444: Remove jobSpec once executor is removed.
		jobSpec := job.NewWorkflowSpec(
			cronjobName,
			workflowId.String(),
			eng.Database.Config(),
			&job.ProcessConfig{
				BinaryDir:          path.Join(eng.AqPath, job.BinaryDir),
				OperatorStorageDir: path.Join(eng.AqPath, job.OperatorStorageDir),
			},
			eng.GithubManager.Config(),
			eng.AqPath,
			nil,
		)

		err := eng.CronjobManager.EditCronJob(
			ctx,
			cronjobName,
			newCronSchedule,
			eng.generateCronFunction(cronjobName, jobSpec),
		)
		if err != nil {
			return errors.Wrap(err, "Unable to change workflow schedule.")
		}
	}
	return nil
}

func (eng *aqEngine) InitEnv(
	ctx context.Context,
	env *exec_env.ExecutionEnvironment,
) error {
	return env.CreateEnv()
}
