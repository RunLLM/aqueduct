package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/airflow"
	"github.com/aqueducthq/aqueduct/lib/container_registry"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/errors"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/execution_state"
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
//
// If this route finishes successfully, then an integration entry is guaranteed to have been created
// in the database.
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

	// Sanitize the root directory path for S3. We remove any leading slash, but force there to always
	// be a trailing slash. eg: `path/to/root/`.
	if service == shared.S3 {
		if root_dir, ok := configMap["root_dir"]; ok && root_dir != "" {
			if root_dir[len(root_dir)-1] != '/' {
				root_dir += "/"
			}
			configMap["root_dir"] = strings.TrimLeft(root_dir, "/")
		}
	}

	staticConfig := auth.NewStaticConfig(configMap)

	// Check if this integration should be used as the new storage layer
	setStorage, err := checkIntegrationSetStorage(service, staticConfig)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to connect integration.")
	}

	return &ConnectIntegrationArgs{
		AqContext:    aqContext,
		Service:      service,
		Name:         name,
		Config:       staticConfig,
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

	// Assumption: we are always ADDING a new integration, so `integrationObj` must be a freshly created integration entry.
	// Note that the config of this returned `integrationObj` may be outdated.
	integrationObj, statusCode, err := ConnectIntegration(ctx, h, args, h.IntegrationRepo, h.Database)
	if err != nil {
		return emptyResp, statusCode, err
	}

	if args.SetAsStorage {
		confData, err := args.Config.Marshal()
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		newStorageConfig, err := storage.ConvertIntegrationConfigToStorageConfig(args.Service, confData)
		if err != nil {
			return emptyResp, http.StatusBadRequest, errors.Wrap(err, "Integration config is malformed.")
		}

		err = storage_migration.Perform(
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
	}

	return emptyResp, http.StatusOK, nil
}

// ConnectIntegration connects a new integration specified by `args`.
// It returns the integration object, the status code for the request and an error, if any.
// If an error is returns, the integration object is guaranteed to be nil. Conversely, the integration
// object is always well-formed on success.
func ConnectIntegration(
	ctx context.Context,
	h *ConnectIntegrationHandler, // This only needs to be non-nil if the integration can be AWS.
	args *ConnectIntegrationArgs,
	integrationRepo repos.Integration,
	DB database.Database,
) (_ *models.Integration, _ int, err error) {
	// Extract non-confidential config
	publicConfig := args.Config.PublicConfig()

	// Always create the integration entry with a running state to start.
	runningAt := time.Now()
	publicConfig[exec_env.ExecStateKey] = execution_state.SerializedRunning(&runningAt)

	// Must open a transaction to write the initial integration state, because the AWS integration
	// may need to perform multiple writes.
	txn, err := DB.BeginTx(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	var integrationObject *models.Integration
	if args.UserOnly {
		// This is a user-specific integration
		integrationObject, err = integrationRepo.CreateForUser(
			ctx,
			args.OrgID,
			args.ID,
			args.Service,
			args.Name,
			(*shared.IntegrationConfig)(&publicConfig),
			txn,
		)
	} else {
		integrationObject, err = integrationRepo.Create(
			ctx,
			args.OrgID,
			args.Service,
			args.Name,
			(*shared.IntegrationConfig)(&publicConfig),
			txn,
		)
	}
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}

	if args.Service == shared.AWS {
		if h == nil {
			return nil, http.StatusInternalServerError, errors.New("Internal error: No route handler present when registering an AWS integration.")
		}
		if statusCode, err := setupCloudIntegration(
			ctx,
			args,
			h,
			txn,
		); err != nil {
			return nil, statusCode, err
		}
	}
	if err := txn.Commit(ctx); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}

	// The initial integration entry has been written. Any errors from this point on will need to update
	// the that entry to reflect the failure. Note that this defer is only relevant for q
	defer func() {
		if err != nil {
			execution_state.UpdateOnFailure(
				ctx,
				"", // outputs
				err.Error(),
				string(args.Service),
				(*shared.IntegrationConfig)(&publicConfig),
				&runningAt,
				integrationObject.ID,
				integrationRepo,
				DB,
			)
		}
	}()

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

	// For those integrations that require asynchronous setup, we spin those up here. When those goroutines are
	// complete, they write their results back to the config column of their integration entry.
	// Note that kicking off any asynchronous setup is the last thing this method does. This ensures that there
	// will never be any status update races between the goroutines and the main thread.
	// TODO(ENG-2523): move base conda env creation outside of ConnectIntegration.
	if args.Service == shared.Conda {
		go func() {
			// We must copy the Database inside the goroutine, because the underlying DB connection
			// will error if passed between the main thread and goroutine.
			condaDB, err := database.NewDatabase(DB.Config())
			if err != nil {
				log.Errorf("Error creating DB for Conda: %v", err)
				return
			}

			condaErr := setupCondaAsync(integrationRepo, integrationObject.ID, publicConfig, runningAt, condaDB)
			if condaErr != nil {
				log.Errorf("Conda setup failed: %v", condaErr)
			}
		}()
	} else if args.Service == shared.Lambda {
		go func() {
			// We must copy the Database inside the goroutine, because the underlying DB connection
			// will error if passed between the main thread and goroutine.
			lambdaDB, err := database.NewDatabase(DB.Config())
			if err != nil {
				log.Errorf("Error creating DB for Lambda: %v", err)
				return
			}

			lambdaErr := setupLambdaAsync(integrationRepo, integrationObject.ID, publicConfig, runningAt, lambdaDB)
			if lambdaErr != nil {
				log.Errorf("Lambda setup failed: %v", lambdaErr)
			}
		}()
	} else {
		// No asynchronous setup is needed for these services, so we can simply mark the connection entries as successful.
		err = execution_state.UpdateOnSuccess(
			ctx,
			string(args.Service),
			(*shared.IntegrationConfig)(&publicConfig),
			&runningAt,
			integrationObject.ID,
			integrationRepo,
			DB,
		)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
	}
	return integrationObject, http.StatusOK, nil
}

// Asynchronously setup the lambda integration.
func setupLambdaAsync(
	integrationRepo repos.Integration,
	integrationID uuid.UUID,
	publicConfig map[string]string,
	runningAt time.Time,
	DB database.Database,
) (err error) {
	defer func() {
		if err != nil {
			execution_state.UpdateOnFailure(
				context.Background(),
				"", // outputs
				err.Error(),
				string(shared.Lambda),
				(*shared.IntegrationConfig)(&publicConfig),
				&runningAt,
				integrationID,
				integrationRepo,
				DB,
			)
		} else {
			_ = execution_state.UpdateOnSuccess(
				context.Background(),
				string(shared.Lambda),
				(*shared.IntegrationConfig)(&publicConfig),
				&runningAt,
				integrationID,
				integrationRepo,
				DB,
			)
		}
	}()

	return lambda_utils.ConnectToLambda(
		context.Background(),
		publicConfig[lambda_utils.RoleArnKey],
	)
}

// Asynchronously setup the conda integration.
func setupCondaAsync(
	integrationRepo repos.Integration,
	integrationID uuid.UUID,
	publicConfig map[string]string,
	runningAt time.Time,
	DB database.Database,
) (err error) {
	var condaPath string
	var output string
	defer func() {
		// Update both the conda path and execution state of the integration's config.
		publicConfig[exec_env.CondaPathKey] = condaPath

		if err != nil {
			execution_state.UpdateOnFailure(
				context.Background(),
				output,
				err.Error(),
				string(shared.Conda),
				(*shared.IntegrationConfig)(&publicConfig),
				&runningAt,
				integrationID,
				integrationRepo,
				DB,
			)
		} else {
			// Update the conda execution state to be successful.
			_ = execution_state.UpdateOnSuccess(
				context.Background(),
				string(shared.Conda),
				(*shared.IntegrationConfig)(&publicConfig),
				&runningAt,
				integrationID,
				integrationRepo,
				DB,
			)
		}
	}()

	condaPath, output, err = exec_env.InitializeConda()
	return err
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

	if service == shared.ECR {
		return validateECRConfig(config)
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

func validateECRConfig(
	config auth.Config,
) (int, error) {
	if err := container_registry.AuthenticateAndUpdateECRConfig(config); err != nil {
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
	}

	// For ECR integration, we require the user to have AWS CLI installed.
	if svc == shared.ECR {
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
