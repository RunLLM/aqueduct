package airflow

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

const (
	pollCompileAirflowInterval = 500 * time.Millisecond
	pollCompileAirflowTimeout  = 30 * time.Second
)

// ScheduleWorkflow schedules a workflow with an Airflow backend.
// It returns the Airflow DAG Python spec file and an error, if any.
//
// Assumptions:
//   - All of the in-memory fields of `dag` are set.
//   - All of the in-memory fields of each Operator in `dag.Operators` are set.
func ScheduleWorkflow(
	ctx context.Context,
	dag *models.DAG,
	dagRepo repos.DAG,
	jobManager job.JobManager,
	vault vault.Vault,
	DB database.Database,
) ([]byte, error) {
	// Generate an Airflow DAG ID
	dagId, err := generateDagId(dag.Metadata.Name)
	if err != nil {
		return nil, err
	}

	// Prepare the storage credentials, so it can be accessed from Airflow
	airflowStorageConfig, err := prepareStorageConfig(ctx, dag, &dag.StorageConfig, vault)
	if err != nil {
		return nil, err
	}

	artifactIDs := make([]uuid.UUID, 0, len(dag.Artifacts))
	for id := range dag.Artifacts {
		artifactIDs = append(artifactIDs, id)
	}

	operatorIDs := make([]uuid.UUID, 0, len(dag.Operators))
	for id := range dag.Operators {
		operatorIDs = append(operatorIDs, id)
	}

	// Generate storage path prefixes for artifact content and artifact metadata.
	// At runtime, the Airflow DAG run ID is appended to
	// the relevant path prefix to form a unique storage path without requiring
	// coordination between Airflow and the Aqueduct server.
	operatorToMetadataPathPrefix := generateStoragePathPrefixes(operatorIDs)
	artifactToContentPathPrefix := generateStoragePathPrefixes(artifactIDs)
	artifactToMetadataPathPrefix := generateStoragePathPrefixes(artifactIDs)

	// Convert the format of these paths into `ExecPaths`, which are used to construct
	// to Artifact/Operator objects. Relies on the fact that an operator -> output artifact
	// is always a unique one-to-one mapping.
	artifactIDToExecPaths := make(map[uuid.UUID]*utils.ExecPaths, len(artifactIDs))
	for _, dbOperator := range dag.Operators {
		for _, outputArtifactID := range dbOperator.Outputs {
			artifactIDToExecPaths[outputArtifactID] = &utils.ExecPaths{
				ArtifactContentPath:  artifactToContentPathPrefix[outputArtifactID],
				ArtifactMetadataPath: artifactToMetadataPathPrefix[outputArtifactID],
				OpMetadataPath:       operatorToMetadataPathPrefix[dbOperator.ID],
			}
		}
	}
	// Take an additional pass over the artifacts to fill in the paths for those that start workflows.
	for _, artifactId := range artifactIDs {
		if _, ok := artifactIDToExecPaths[artifactId]; !ok {
			artifactIDToExecPaths[artifactId] = &utils.ExecPaths{
				ArtifactContentPath:  artifactToContentPathPrefix[artifactId],
				ArtifactMetadataPath: artifactToMetadataPathPrefix[artifactId],
				OpMetadataPath:       "", // Artifacts with no input operators have no operator metadata path.
			}
		}
	}

	operatorToTask := make(map[uuid.UUID]string, len(dag.Operators))
	taskToJobSpec := make(map[string]job.Spec, len(dag.Operators))

	// For each operator, generate a job spec that can be used to turn an operator
	// into an Airflow task.
	for _, op := range dag.Operators {

		// An Airflow DAG cannot have any custom operator engine specs, the entire DAG
		// must be executed all on Airflow.
		if op.Spec.EngineConfig() != nil {
			return nil, errors.Newf("Custom engine set on operator, which is disallowed for Airflow.")
		}

		// Prepare `op`'s input artifacts
		inputArtifacts := make([]artifact.Artifact, 0, len(op.Inputs))
		inputExecPaths := make([]*utils.ExecPaths, 0, len(op.Inputs))
		for _, artifactId := range op.Inputs {
			dbInputArtifact, ok := dag.Artifacts[artifactId]
			if !ok {
				return nil, errors.Newf("cannot find artifact with ID %v", artifactId)
			}

			inputArtifact, err := artifact.NewArtifact(
				uuid.Nil, /* Airflow does not use the preview cache */
				dbInputArtifact,
				artifactIDToExecPaths[artifactId],
				nil, /* artifactWriter */
				nil, /* artifactResultWriter */
				&dag.StorageConfig,
				nil, /* artifactCacheManager */
				nil, /* db */
			)
			if err != nil {
				return nil, err
			}

			inputArtifacts = append(inputArtifacts, inputArtifact)
			inputExecPaths = append(inputExecPaths, artifactIDToExecPaths[artifactId])
		}

		// Prepare `op`'s output artifacts
		outputArtifacts := make([]artifact.Artifact, 0, len(op.Outputs))
		outputExecPaths := make([]*utils.ExecPaths, 0, len(op.Outputs))
		for _, artifactId := range op.Outputs {
			dbOutputArtifact, ok := dag.Artifacts[artifactId]
			if !ok {
				return nil, errors.Newf("cannot find artifact with ID %v", artifactId)
			}

			outputArtifact, err := artifact.NewArtifact(
				uuid.Nil, /* Airflow does not use the preview cache */
				dbOutputArtifact,
				artifactIDToExecPaths[artifactId],
				nil, /* artifactWriter */
				nil, /* artifactResultWriter */
				&dag.StorageConfig,
				nil, /* previewCacheManager */
				nil, /* db */
			)
			if err != nil {
				return nil, err
			}

			outputArtifacts = append(outputArtifacts, outputArtifact)
			outputExecPaths = append(outputExecPaths, artifactIDToExecPaths[artifactId])
		}

		airflowOperator, err := operator.NewOperator(
			ctx,
			op,
			inputArtifacts,
			outputArtifacts,
			inputExecPaths,
			outputExecPaths,
			nil,
			dag.EngineConfig,
			vault,
			&airflowStorageConfig,
			nil,              /* previewCacheManager */
			operator.Publish, // airflow operator will never run in preview mode
			nil,              /* ExecEnv */
			"",               /* aqPath */
			DB,
		)
		if err != nil {
			return nil, err
		}

		// Generate the job spec for this operator
		jobSpec := airflowOperator.JobSpec()

		taskId, err := generateTaskId(airflowOperator.Name())
		if err != nil {
			return nil, err
		}

		operatorToTask[airflowOperator.ID()] = taskId
		taskToJobSpec[taskId] = jobSpec
	}

	// Airflow needs to know the explicit dependencies between operators, so there can
	// only be operator to operator edges.
	taskEdges, err := computeEdges(dag.Operators, operatorToTask)
	if err != nil {
		return nil, err
	}

	operatorMetadataPath := fmt.Sprintf("compile-airflow-metadata-%s", uuid.New().String())
	operatorOutputPath := fmt.Sprintf("compile-airflow-output-%s", uuid.New().String())

	defer func() {
		go utils.CleanupStorageFiles(ctx, &dag.StorageConfig, []string{operatorMetadataPath, operatorOutputPath})
	}()

	jobName := fmt.Sprintf("compile-airflow-operator-%s", uuid.New().String())
	jobSpec, err := job.NewCompileAirflowSpec(
		jobName,
		dag.ID,
		&dag.StorageConfig,
		operatorMetadataPath,
		operatorOutputPath,
		dagId,
		string(dag.Metadata.Schedule.CronSchedule),
		taskToJobSpec,
		taskEdges,
	)
	if err != nil {
		return nil, err
	}

	if err := jobManager.Launch(ctx, jobSpec.JobName(), jobSpec); err != nil {
		return nil, err
	}

	jobStatus, err := job.PollJob(ctx, jobSpec.JobName(), jobManager, pollCompileAirflowInterval, pollCompileAirflowTimeout)
	if err != nil {
		return nil, err
	}

	if jobStatus == shared.FailedExecutionStatus {
		return nil, errors.New("Compile Airflow job failed.")
	}

	var execState shared.ExecutionState
	if err := utils.ReadFromStorage(
		ctx,
		&dag.StorageConfig,
		operatorMetadataPath,
		&execState,
	); err != nil {
		return nil, err
	}

	if execState.Status == shared.FailedExecutionStatus {
		return nil, errors.Newf("Compile Airflow job failed: %v \n logs: %v \n failure type: %v", execState.Error, execState.UserLogs, execState.FailureType)
	}

	airflowDagFile, err := storage.NewStorage(&dag.StorageConfig).Get(ctx, operatorOutputPath)
	if err != nil {
		return nil, err
	}

	// Update the AirflowRuntimeConfig for `dag`
	newRuntimeConfig := dag.EngineConfig
	newRuntimeConfig.AirflowConfig.DagId = dagId
	// The DAGs will not match until the user has copied over the newly generated `airflowDagFile`
	newRuntimeConfig.AirflowConfig.MatchesAirflow = false
	newRuntimeConfig.AirflowConfig.OperatorToTask = operatorToTask
	newRuntimeConfig.AirflowConfig.OperatorMetadataPathPrefix = operatorToMetadataPathPrefix
	newRuntimeConfig.AirflowConfig.ArtifactContentPathPrefix = artifactToContentPathPrefix
	newRuntimeConfig.AirflowConfig.ArtifactMetadataPathPrefix = artifactToMetadataPathPrefix

	_, err = dagRepo.Update(
		ctx,
		dag.ID,
		map[string]interface{}{
			models.DagEngineConfig: &newRuntimeConfig,
		},
		DB,
	)
	if err != nil {
		return nil, err
	}

	return airflowDagFile, nil
}

// prepareStorageConfig returns the StorageConfig so operators can access storage from the
// Airflow engine.
func prepareStorageConfig(
	ctx context.Context,
	dag *models.DAG,
	storageConfig *shared.StorageConfig,
	vault vault.Vault,
) (shared.StorageConfig, error) {
	emptyStorageConf := shared.StorageConfig{}

	authConf, err := auth.ReadConfigFromSecret(ctx, dag.EngineConfig.AirflowConfig.IntegrationId, vault)
	if err != nil {
		return emptyStorageConf, err
	}

	airflowConf, err := parseConfig(authConf)
	if err != nil {
		return emptyStorageConf, err
	}

	if storageConfig.Type != shared.S3StorageType {
		return emptyStorageConf, errors.New("The StorageType must be S3 to use the Airflow engine.")
	}

	return shared.StorageConfig{
		Type: storageConfig.Type,
		S3Config: &shared.S3Config{
			Region:             storageConfig.S3Config.Region,
			Bucket:             storageConfig.S3Config.Bucket,
			CredentialsPath:    airflowConf.S3CredentialsPath,
			CredentialsProfile: airflowConf.S3CredentialsProfile,
		},
	}, nil
}

// generateStoragePathPrefixes generates a storage path prefix for each ID.
func generateStoragePathPrefixes(ids []uuid.UUID) map[uuid.UUID]string {
	paths := make(map[uuid.UUID]string, len(ids))
	for _, id := range ids {
		paths[id] = uuid.NewString()
	}
	return paths
}

func computeEdges(operators map[uuid.UUID]models.Operator, operatorToTask map[uuid.UUID]string) (map[string][]string, error) {
	artifactToSrc := map[uuid.UUID]string{}
	artifactToDests := map[uuid.UUID][]string{}

	for _, op := range operators {
		taskId, ok := operatorToTask[op.ID]
		if !ok {
			return nil, errors.Newf("Unable to find task ID for operator %v", op.ID)
		}

		for _, outputArtifact := range op.Outputs {
			artifactToSrc[outputArtifact] = taskId
		}

		for _, inputArtifact := range op.Inputs {
			artifactToDests[inputArtifact] = append(artifactToDests[inputArtifact], taskId)
		}
	}

	taskEdges := map[string][]string{}

	for artifactId, srcTask := range artifactToSrc {
		destTasks, ok := artifactToDests[artifactId]
		if !ok {
			// This artifact has no downstream operators
			continue
		}

		// There is an implicit edge between `srcTask` and each task in `destTasks` via
		// the artifact `artifactId`.
		taskEdges[srcTask] = append(taskEdges[srcTask], destTasks...)
	}

	return taskEdges, nil
}
