package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/airflow"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/errors"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/job"
	lambda_utils "github.com/aqueducthq/aqueduct/lib/lambda"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/notification"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/storage_migration"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
)

const (
	pollAuthenticateInterval = 500 * time.Millisecond
	pollAuthenticateTimeout  = 2 * time.Minute
)

var pathConfigKeys = map[string]bool{
	"config_file_path":    true, // AWS, S3, Athena credentials path
	"kubeconfig_path":     true, // K8s credentials path
	"s3_credentials_path": true, // Airflow S3 credentials path
	"database":            true, // SQLite database path
}

// Route: /integration/connect
// Method: POST
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//		`integration-name`: the name for the integration
//		`integration-service`: the service type for the integration
//		`integration-config`: the json-serialized integration config
//
// Response: none
type ConnectIntegrationHandler struct {
	PostHandler

	Database   database.Database
	JobManager job.JobManager

	ArtifactRepo         repos.Artifact
	ArtifactResultRepo   repos.ArtifactResult
	DAGRepo              repos.DAG
	IntegrationRepo      repos.Integration
	StorageMigrationRepo repos.StorageMigration
	OperatorRepo         repos.Operator

	PauseServer   func()
	RestartServer func()
}

func (*ConnectIntegrationHandler) Headers() []string {
	return []string{
		routes.IntegrationNameHeader,
		routes.IntegrationServiceHeader,
		routes.IntegrationConfigHeader,
	}
}

type ConnectIntegrationArgs struct {
	*aq_context.AqContext
	Name         string         // User specified name for the integration
	Service      shared.Service // Name of the service to connect (e.g. Snowflake, Postgres)
	Config       auth.Config    // Integration config
	UserOnly     bool           // Whether the integration is only accessible by the user or the entire org
	SetAsStorage bool           // Whether the integration should be used as the storage layer
}

type ConnectIntegrationResponse struct{}

func (*ConnectIntegrationHandler) Name() string {
	return "ConnectIntegration"
}

func (h *ConnectIntegrationHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Unable to connect integration.")
	}

	service, userOnly, err := request.ParseIntegrationServiceFromRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to connect integration.")
	}

	name, configMap, err := request.ParseIntegrationConfigFromRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to connect integration.")
	}

	if name == "" {
		return nil, http.StatusBadRequest, errors.New("Integration name is not provided")
	}

	if service == shared.Github || service == shared.GoogleSheets {
		return nil, http.StatusBadRequest, errors.Newf("%s integration type is currently not supported", service)
	}

	if err = convertToAbsolutePath(configMap); err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Error getting server's home directory path")
	}

	config := auth.NewStaticConfig(configMap)

	// Check if this integration should be used as the new storage layer
	setStorage, err := checkIntegrationSetStorage(service, config)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to connect integration.")
	}

	return &ConnectIntegrationArgs{
		AqContext:    aqContext,
		Service:      service,
		Name:         name,
		Config:       config,
		UserOnly:     userOnly,
		SetAsStorage: setStorage,
	}, http.StatusOK, nil
}

func (h *ConnectIntegrationHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*ConnectIntegrationArgs)

	emptyResp := ConnectIntegrationResponse{}

	statusCode, err := ValidatePrerequisites(
		ctx,
		args.Service,
		args.Name,
		args.ID,
		args.OrgID,
		h.IntegrationRepo,
		h.Database,
	)
	if err != nil {
		return emptyResp, statusCode, err
	}

	// Validate integration config
	statusCode, err = ValidateConfig(
		ctx,
		args.RequestID,
		args.Config,
		args.Service,
		h.JobManager,
		args.StorageConfig,
	)
	if err != nil {
		return emptyResp, statusCode, err
	}

	txn, err := h.Database.BeginTx(ctx)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	// Assumption: we are always ADDING a new integration, so `integrationObj` must be a freshly created
	// integration entry on success.
	integrationObj, statusCode, err := ConnectIntegration(ctx, args, h.IntegrationRepo, txn)
	if err != nil {
		return emptyResp, statusCode, err
	}

	if args.Service == shared.AWS {
		if statusCode, err := setupCloudIntegration(
			ctx,
			args,
			h,
			txn,
		); err != nil {
			return emptyResp, statusCode, err
		}
	}

	if err := txn.Commit(ctx); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}

	if args.SetAsStorage {

		// TODO: REMOVE AQPATH
		confData, err := args.Config.Marshal()
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		newStorageConfig, err := storage.ConvertIntegrationConfigToStorageConfig(args.Service, confData)
		if err != nil {
			return emptyResp, http.StatusBadRequest, errors.Wrap(err, "Integration config is malformed.")
		}

		err = storage_migration.PerformStorageMigration(
			ctx,
			args.OrgID,
			integrationObj,
			newStorageConfig,
			h.PauseServer,
			h.RestartServer,
			h.ArtifactRepo,
			h.ArtifactResultRepo,
			h.DAGRepo,
			h.IntegrationRepo,
			h.OperatorRepo,
			h.StorageMigrationRepo,
			h.Database,
		)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to migrate storage layer.")
		}

		//// This integration should be used as the new storage layer.
		//// In order to do so, we need to migrate all content from the old store
		//// to the new store. This requires pausing the server and then restarting it.
		//
		//// The migration logic is performed asynchronously, and it's process is tracked
		//// as a new entry in the `storage_migration` table:
		//storageMigrationObj, err := h.StorageMigrationRepo.Create(
		//	ctx,
		//	&integrationObj.ID,
		//	h.Database,
		//)
		//if err != nil {
		//	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to migrate storage.")
		//}
		//
		//// If the migration is successful, the new entry is given a success execution status, along with `current=True`.
		//// If the migration is unsuccessful, the error is recorded on the new entry in `storage_migration`.
		//go func() {
		//	// Shadows the context in the outer scope on purpose.
		//	ctx := context.Background()
		//
		//	log.Info("Starting storage migration process...")
		//	// Wait until the server is paused
		//	h.PauseServer()
		//	// Makes sure that the server is restarted
		//	defer h.RestartServer()
		//
		//	execState := storageMigrationObj.ExecState
		//
		//	var err error
		//	defer func() {
		//		if err != nil {
		//			// Regardless of whether the migration succeeded, we should track the finished timestamp.
		//			finishedAt := time.Now()
		//			execState.Timestamps.FinishedAt = &finishedAt
		//
		//			execState.UpdateWithFailure(
		//				// TODO: this can be a system error too. But no one cares right now.
		//				shared.UserFatalFailure,
		//				&shared.Error{
		//					Tip:     fmt.Sprintf("Failure occurred when migrating to the new storage integration %s.", integrationObj.Name),
		//					Context: err.Error(),
		//				},
		//			)
		//			err = h.updateStorageMigrationExecState(ctx, storageMigrationObj.ID, &execState)
		//			if err != nil {
		//				log.Errorf("Unexpected error when updating the storage migration entry to FAILED: %v", err)
		//				return
		//			}
		//
		//		}
		//	}()
		//
		//	// Mark the migration explicitly as RUNNING.
		//	runningAt := time.Now()
		//	execState.Timestamps.RunningAt = &runningAt
		//	err = h.updateStorageMigrationExecState(ctx, storageMigrationObj.ID, &execState)
		//	if err != nil {
		//		log.Errorf("Unexpected error when updating the storage migration entry to RUNNING: %v", err)
		//		return
		//	}
		//
		//	// TODO: REMOVE
		//	if strings.HasPrefix(integrationObj.Name, "failing") {
		//		err = errors.Newf("I REALLY DONT LIKE YOUR NAME %s SIR", integrationObj.Name)
		//		return
		//	}
		//
		//	// Actually perform the storage migration.
		//	storageConfig, storageCleanupConfig, err := h.performStorageMigration(args.Service, args.Config, args.OrgID)
		//	// We let the defer() handle the failure case appropriately.
		//	if err != nil {
		//		return
		//	}
		//
		//	log.Info("Successfully migrated the storage layer!")
		//	finishedAt := time.Now()
		//	execState.Timestamps.FinishedAt = &finishedAt
		//	execState.Status = shared.SucceededExecutionStatus
		//
		//	// The update of the storage config and storage migration entry should happen together.
		//	// While we don't enforce this atomically, we can make the two update together to minimize the risk.
		//	err = h.updateStorageMigrationExecState(ctx, storageMigrationObj.ID, &execState)
		//	if err != nil {
		//		log.Errorf("Unexpected error when updating the storage migration entry to SUCCESS: %v", err)
		//		return
		//	}
		//
		//	err = config.UpdateStorage(storageConfig)
		//	if err != nil {
		//		log.Errorf("Unexpected error when updating the global storage layer config: %v", err)
		//		return
		//	} else {
		//		log.Info("Successfully updated the global storage layer config!")
		//	}
		//
		//	// We only perform best-effort deletion the old storage layer files here, after everything else has succeede.
		//	for _, key := range storageCleanupConfig.StoreKeys {
		//		if err := storageCleanupConfig.Store.Delete(ctx, key); err != nil {
		//			log.Errorf("Unexpected error when deleting the old storage file %s: %v", key, err)
		//		}
		//	}
		//
		//	for _, key := range storageCleanupConfig.VaultKeys {
		//		if err := storageCleanupConfig.Vault.Delete(ctx, key); err != nil {
		//			log.Errorf("Unexpected error when deleting the old vault file %s: %v", key, err)
		//		}
		//	}
		//}()
	}

	return emptyResp, http.StatusOK, nil
}

// Also updates `current=True` if the execution state is marked as SUCCESS!
//func (h *ConnectIntegrationHandler) updateStorageMigrationExecState(
//	ctx context.Context,
//	storageMigrationID uuid.UUID,
//	execState *shared.ExecutionState,
//) error {
//	updates := map[string]interface{}{
//		models.StorageMigrationExecutionState: execState,
//	}
//
//	// This is updated to a transaction if we also need to mark an old entry as current=False.
//	db := h.Database
//	if execState.Status == shared.SucceededExecutionStatus {
//		updates[models.StorageMigrationCurrent] = true
//
//		// If there was a previous storage migration, update that entry to be `current=False`.
//		oldStorageMigrationObj, err := h.StorageMigrationRepo.Current(ctx, h.Database)
//		if err == nil {
//			// Updating the old storage migration entry must be done in the same transaction
//			// as the update of the new storage migration entry, so that there is at most one
//			// current=True entry in the storage_migration table.
//			txn, err := h.Database.BeginTx(context.Background())
//			if err != nil {
//				return errors.Wrap(err, "Unable to start transaction for updating storage state.")
//			}
//			defer database.TxnRollbackIgnoreErr(ctx, txn)
//
//			db = txn
//			_, err = h.StorageMigrationRepo.Update(
//				ctx,
//				oldStorageMigrationObj.ID,
//				map[string]interface{}{
//					models.StorageMigrationCurrent: false,
//				},
//				db,
//			)
//			if err != nil {
//				return errors.Wrap(err, "Unexpected error when updating old storage migration entry to be non-current")
//			}
//		} else if !aq_errors.Is(err, database.ErrNoRows()) {
//			return errors.Wrap(err, "Unexpected error when fetching current storage state.")
//		}
//		// Continue without doing anything if there was no previous storage migration.
//	}
//
//	// Perform the actual intended execution state update.
//	_, err := h.StorageMigrationRepo.Update(
//		ctx,
//		storageMigrationID,
//		updates,
//		db,
//	)
//	if err != nil {
//		return errors.Wrap(err, "Unexpected error when updating storage migration execution state.")
//	}
//
//	// Only need to do this if we're committing a transaction.
//	if txn, ok := db.(database.Transaction); ok {
//		err = txn.Commit(ctx)
//		if err != nil {
//			return errors.Wrap(err, "Unexpected error when committing storage migration execution state update.")
//		}
//	}
//	return nil
//}
//
//// Migrate the storage (and vault) context to the new storage layer.
//// The global config for the server is *NOT* updated here.
//// Returns the new storage config, along with any cleanup information that should be done by the caller.
//func (h *ConnectIntegrationHandler) performStorageMigration(
//	svc shared.Service,
//	conf auth.Config,
//	orgID string,
//) (*shared.StorageConfig, *storage_migration.StorageCleanupConfig, error) {
//	// Wait until there are no more workflow runs in progress
//	lock := utils.NewExecutionLock()
//	if err := lock.Lock(); err != nil {
//		return nil, nil, errors.Wrap(err, "Unexpected error when acquiring workflow execution lock.")
//	}
//	defer func() {
//		if lockErr := lock.Unlock(); lockErr != nil {
//			log.Errorf("Unexpected error when unlocking workflow execution lock: %v", lockErr)
//		}
//	}()
//
//	data, err := conf.Marshal()
//	if err != nil {
//		return nil, nil, err
//	}
//
//	var storageConfig *shared.StorageConfig
//
//	switch svc {
//	case shared.S3:
//		var c shared.S3IntegrationConfig
//		if err := json.Unmarshal(data, &c); err != nil {
//			return nil, nil, err
//		}
//
//		storageConfig, err = convertS3IntegrationtoStorageConfig(&c)
//		if err != nil {
//			return nil, nil, err
//		}
//	case shared.GCS:
//		var c shared.GCSIntegrationConfig
//		if err := json.Unmarshal(data, &c); err != nil {
//			return nil, nil, err
//		}
//
//		storageConfig = convertGCSIntegrationtoStorageConfig(&c)
//	default:
//		return nil, nil, errors.Newf("%v cannot be used as the storage layer", svc)
//	}
//
//	// Migrate all storage content to the new storage config
//	currentStorageConfig := config.Storage()
//	storageCleanupConfig, err := storage_migration.MigrateStorageAndVault(
//		context.Background(),
//		&currentStorageConfig,
//		storageConfig,
//		orgID,
//		h.DAGRepo,
//		h.ArtifactRepo,
//		h.ArtifactResultRepo,
//		h.OperatorRepo,
//		h.IntegrationRepo,
//		h.Database,
//	)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	return storageConfig, storageCleanupConfig, nil
//}

// ConnectIntegration connects a new integration specified by `args`.
// It returns the integration object, the status code for the request and an error, if any.
// If an error is returns, the integration object is guaranteed to be nil. Conversely, the integration
// object is always well-formed on success.
func ConnectIntegration(
	ctx context.Context,
	args *ConnectIntegrationArgs,
	integrationRepo repos.Integration,
	DB database.Database,
) (*models.Integration, int, error) {
	// Extract non-confidential config
	publicConfig := args.Config.PublicConfig()

	var integrationObject *models.Integration
	var err error
	if args.UserOnly {
		// This is a user-specific integration
		integrationObject, err = integrationRepo.CreateForUser(
			ctx,
			args.OrgID,
			args.ID,
			args.Service,
			args.Name,
			(*shared.IntegrationConfig)(&publicConfig),
			true,
			DB,
		)
	} else {
		integrationObject, err = integrationRepo.Create(
			ctx,
			args.OrgID,
			args.Service,
			args.Name,
			(*shared.IntegrationConfig)(&publicConfig),
			true,
			DB,
		)
	}
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to initialize vault.")
	}

	// Store config (including confidential information) in vault
	if err := auth.WriteConfigToSecret(
		ctx,
		integrationObject.ID,
		args.Config,
		vaultObject,
	); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}

	// TODO(ENG-2523): move base conda env creation outside of ConnectIntegration.
	if args.Service == shared.Conda {
		go func() {
			DB, err = database.NewDatabase(DB.Config())
			if err != nil {
				log.Errorf("Error creating DB in go routine: %v", err)
				return
			}

			exec_env.InitializeConda(
				context.Background(),
				integrationObject.ID,
				integrationRepo,
				DB,
			)
		}()
	}

	if args.Service == shared.Lambda {
		go func() {
			DB, err = database.NewDatabase(DB.Config())
			if err != nil {
				log.Errorf("Error creating DB in go routine: %v", err)
				return
			}

			lambda_utils.ConnectToLambda(
				context.Background(),
				args.Config,
				integrationObject.ID,
				integrationRepo,
				DB,
			)
		}()
	}

	return integrationObject, http.StatusOK, nil
}

// ValidateConfig authenticates the config provided.
// It returns a status code and an error, if any.
func ValidateConfig(
	ctx context.Context,
	requestId string,
	config auth.Config,
	service shared.Service,
	jobManager job.JobManager,
	storageConfig *shared.StorageConfig,
) (int, error) {
	if service == shared.Airflow {
		// Airflow authentication is performed via the Go client
		// instead of the Python client, so we don't launch a job for it.
		return validateAirflowConfig(ctx, config)
	}

	if service == shared.Kubernetes {
		// Kuerbnetes authentication is performed via initializing a k8s client
		// instead of the Python client, so we don't launch a job for it.
		return validateKubernetesConfig(ctx, config)
	}

	if service == shared.Lambda {
		// Lambda authentication is performed in ConnectToLambda()
		// by creating Lambda jobs instead of the Python client,
		// so we don't launch a job for it.
		return http.StatusOK, nil
	}

	if service == shared.Databricks {
		// Databricks authentication is performed by posting a ListJobs
		// request, so we don't launch a job for it.
		return validateDatabricksConfig(ctx, config)
	}

	if service == shared.Spark {
		return validateSparkConfig(ctx, config)
	}

	if service == shared.Email {
		return validateEmailConfig(config)
	}

	if service == shared.Slack {
		return validateSlackConfig(config)
	}

	if service == shared.AWS {
		return validateAWSConfig(config)
	}

	jobName := fmt.Sprintf("authenticate-operator-%s", uuid.New().String())
	if service == shared.Conda {
		return validateConda()
	}

	// Schedule authenticate job
	jobMetadataPath := fmt.Sprintf("authenticate-%s", requestId)

	defer func() {
		// Delete storage files created for authenticate job metadata
		go utils.CleanupStorageFiles(context.Background(), storageConfig, []string{jobMetadataPath})
	}()

	jobSpec := job.NewAuthenticateSpec(
		jobName,
		storageConfig,
		jobMetadataPath,
		service,
		config,
	)

	if err := jobManager.Launch(ctx, jobName, jobSpec); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to launch authenticate job.")
	}

	jobStatus, err := job.PollJob(ctx, jobName, jobManager, pollAuthenticateInterval, pollAuthenticateTimeout)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}

	if jobStatus == shared.SucceededExecutionStatus {
		// Authentication was successful
		return http.StatusOK, nil
	}

	// Authentication failed, so we need to fetch the error message from storage
	var execState shared.ExecutionState
	if err := utils.ReadFromStorage(
		ctx,
		storageConfig,
		jobMetadataPath,
		&execState,
	); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}

	if execState.Error != nil {
		return http.StatusBadRequest, errors.Newf(
			"Unable to authenticate.\n%s\n%s",
			execState.Error.Tip,
			execState.Error.Context,
		)
	}

	return http.StatusInternalServerError, errors.New(
		"Unable to authenticate credentials, we couldn't obtain more context at this point.",
	)
}

// validateAirflowConfig authenticates the Airflow config provided.
// It returns a status code and an error, if any.
func validateAirflowConfig(
	ctx context.Context,
	config auth.Config,
) (int, error) {
	if err := airflow.Authenticate(ctx, config); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

// checkIntegrationSetStorage returns whether this integration should be used as the storage layer.
func checkIntegrationSetStorage(svc shared.Service, conf auth.Config) (bool, error) {
	if svc != shared.S3 && svc != shared.GCS {
		// Only S3 and GCS can be used for storage
		return false, nil
	}

	data, err := conf.Marshal()
	if err != nil {
		return false, err
	}

	switch svc {
	case shared.S3:
		var c shared.S3IntegrationConfig
		if err := json.Unmarshal(data, &c); err != nil {
			return false, err
		}
		return bool(c.UseAsStorage), nil
	case shared.GCS:
		var c shared.GCSIntegrationConfig
		if err := json.Unmarshal(data, &c); err != nil {
			return false, err
		}
		return bool(c.UseAsStorage), nil
	default:
		return false, errors.Newf("%v cannot be used as the metadata storage layer", svc)
	}
}

func convertS3IntegrationtoStorageConfig(c *shared.S3IntegrationConfig) (*shared.StorageConfig, error) {
	// Users provide AWS credentials for an S3 integration via one of the following:
	//  1. AWS Access Key and Secret Key
	//  2. Credentials file content
	//  3. Credentials filepath and profile name
	// The S3 Storage implementation expects the AWS credentials to be specified via a
	// filepath and profile name, so we must convert the above to the correct format.
	storageConfig := &shared.StorageConfig{
		Type: shared.S3StorageType,
		S3Config: &shared.S3Config{
			Bucket: fmt.Sprintf("s3://%s", c.Bucket),
			Region: c.Region,
		},
	}
	switch c.Type {
	case shared.AccessKeyS3ConfigType:
		// AWS access and secret keys need to be written to a credentials file
		path := filepath.Join(config.AqueductPath(), "storage", uuid.NewString())
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		credentialsContent := fmt.Sprintf(
			"[default]\naws_access_key_id=%s\naws_secret_access_key=%s\n",
			c.AccessKeyId,
			c.SecretAccessKey,
		)
		if _, err := f.WriteString(credentialsContent); err != nil {
			return nil, err
		}

		storageConfig.S3Config.CredentialsPath = path
		storageConfig.S3Config.CredentialsProfile = "default"
	case shared.ConfigFileContentS3ConfigType:
		// The credentials content needs to be written to a credentials file
		path := filepath.Join(config.AqueductPath(), "storage", uuid.NewString())
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		// Determine profile name by looking for [profile_name]
		i := strings.Index(c.ConfigFileContent, "[")
		if i < 0 {
			return nil, errors.New("Unable to determine AWS credentials profile name.")
		}

		j := strings.Index(c.ConfigFileContent, "]")
		if j < 0 {
			return nil, errors.New("Unable to determine AWS credentials profile name.")
		}

		profileName := c.ConfigFileContent[i+1 : j]

		if _, err := f.WriteString(c.ConfigFileContent); err != nil {
			return nil, err
		}

		storageConfig.S3Config.CredentialsPath = path
		storageConfig.S3Config.CredentialsProfile = profileName
	case shared.ConfigFilePathS3ConfigType:
		// The credentials are already in the form of a filepath and profile, so no changes
		// need to be made
		storageConfig.S3Config.CredentialsPath = c.ConfigFilePath
		storageConfig.S3Config.CredentialsProfile = c.ConfigFileProfile
	default:
		return nil, errors.Newf("Unknown S3ConfigType: %v", c.Type)
	}

	return storageConfig, nil
}

func convertGCSIntegrationtoStorageConfig(c *shared.GCSIntegrationConfig) *shared.StorageConfig {
	return &shared.StorageConfig{
		Type: shared.GCSStorageType,
		GCSConfig: &shared.GCSConfig{
			Bucket:                    c.Bucket,
			ServiceAccountCredentials: c.ServiceAccountCredentials,
		},
	}
}

func validateKubernetesConfig(
	ctx context.Context,
	config auth.Config,
) (int, error) {
	if err := engine.AuthenticateK8sConfig(ctx, config); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

func validateDatabricksConfig(
	ctx context.Context,
	config auth.Config,
) (int, error) {
	if err := engine.AuthenticateDatabricksConfig(ctx, config); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

func validateSparkConfig(
	ctx context.Context,
	config auth.Config,
) (int, error) {
	// Validate that we are able to connect to the Spark cluster via Livy.
	if err := engine.AuthenticateSparkConfig(ctx, config); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

func validateEmailConfig(config auth.Config) (int, error) {
	emailConfig, err := lib_utils.ParseEmailConfig(config)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if err := notification.AuthenticateEmail(emailConfig); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

func validateSlackConfig(config auth.Config) (int, error) {
	slackConfig, err := lib_utils.ParseSlackConfig(config)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if err := notification.AuthenticateSlack(slackConfig); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

func validateAWSConfig(
	config auth.Config,
) (int, error) {
	if err := engine.AuthenticateAWSConfig(config); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

// ValidatePrerequisites validates if the integration for the given service can be connected at all.
// 1) Checks if an integration already exists for unique integrations including conda, email, and slack.
// 2) Checks if the name has already been taken.
func ValidatePrerequisites(
	ctx context.Context,
	svc shared.Service,
	name string,
	userID uuid.UUID,
	orgID string,
	integrationRepo repos.Integration,
	DB database.Database,
) (int, error) {
	// We expect the new name to be unique.
	_, err := integrationRepo.GetByNameAndUser(ctx, name, userID, orgID, DB)
	if err == nil {
		return http.StatusBadRequest, errors.Newf("Cannot connect to an integration %s, since it already exists.", name)
	}
	if !errors.Is(err, database.ErrNoRows()) {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to query for existing integrations.")
	}

	if svc == shared.Conda {
		condaIntegration, err := exec_env.GetCondaIntegration(
			ctx, userID, integrationRepo, DB,
		)
		if err != nil {
			return http.StatusInternalServerError, errors.Wrap(err, "Unable to verify if conda is connected.")
		}

		if condaIntegration != nil {
			return http.StatusBadRequest, errors.Newf(
				"You already have conda integration %s connected.",
				condaIntegration.Name,
			)
		}

		if err = exec_env.ValidateCondaDevelop(); err != nil {
			return http.StatusBadRequest, errors.Wrap(
				err,
				"Failed to run `conda develop`. We use this to help set up conda environments. Please install the dependency before connecting Aqueduct to Conda. Typically, this can be done by running `conda install conda-build`.",
			)
		}

		return http.StatusOK, nil
	}

	if svc != shared.Conda && shared.IsComputeIntegration(svc) {
		// For all non-conda compute integrations, we require the metadata store to be cloud storage.
		if config.Storage().Type == shared.FileStorageType {
			return http.StatusBadRequest, errors.Newf("You need to setup cloud storage as metadata store before registering compute integration of type %s.", svc)
		}
	}

	// These integrations should be unique.
	if svc == shared.Email || svc == shared.Slack {
		integrations, err := integrationRepo.GetByServiceAndUser(ctx, svc, userID, DB)
		if err != nil {
			return http.StatusInternalServerError, errors.Wrap(err, "Unable to verify if email is connected.")
		}

		if len(integrations) > 0 {
			return http.StatusBadRequest, errors.Newf(
				"You already have an %s integration %s connected.",
				svc,
				integrations[0].Name,
			)
		}

		return http.StatusOK, nil
	}

	// For AWS integration, we require the user to have AWS CLI and Terraform installed.
	// We also require env (GNU coreutils) executable to set the env variables when using the AWS CLI.
	if svc == shared.AWS {
		if _, _, err := lib_utils.RunCmd("terraform", []string{"--version"}, "", false); err != nil {
			return http.StatusNotFound, errors.Wrap(err, "terraform executable not found. Please go to https://developer.hashicorp.com/terraform/downloads to install terraform")
		}

		awsVersionString, _, err := lib_utils.RunCmd("aws", []string{"--version"}, "", false)
		if err != nil {
			return http.StatusNotFound, errors.Wrap(err, "AWS CLI executable not found. Please go to https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html to install AWS CLI")
		}

		awsVersion, err := version.NewVersion(strings.Split(strings.Split(awsVersionString, " ")[0], "/")[1])
		if err != nil {
			return http.StatusUnprocessableEntity, errors.Wrap(err, "Error parsing AWS CLI version")
		}

		requiredVersion, _ := version.NewVersion("2.11.5")
		if awsVersion.LessThan(requiredVersion) {
			return http.StatusUnprocessableEntity, errors.Wrapf(err, "AWS CLI version 2.11.5 and above is required, but you got %s. Please update!", awsVersion.String())
		}

		if _, _, err := lib_utils.RunCmd("env", []string{"--version"}, "", false); err != nil {
			return http.StatusNotFound, errors.Wrap(err, "env (GNU coreutils) executable not found.")
		}
	}

	return http.StatusOK, nil
}

func validateConda() (int, error) {
	errMsg := "Unable to validate conda installation. Do you have conda installed?"
	_, _, err := lib_utils.RunCmd(exec_env.CondaCmdPrefix, []string{"--version"}, "", false)
	if err != nil {
		return http.StatusBadRequest, errors.Wrap(err, errMsg)
	}

	return http.StatusOK, nil
}

func convertToAbsolutePath(configMap map[string]string) error {
	for key, path := range configMap {
		if _, ok := pathConfigKeys[key]; ok {
			if strings.HasPrefix(path, "~") {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				configMap[key] = strings.Replace(path, "~", homeDir, 1)
			}
		}
	}

	return nil
}
