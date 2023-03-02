package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/airflow"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/dynamic"
	"github.com/aqueducthq/aqueduct/lib/engine"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/job"
	lambda_utils "github.com/aqueducthq/aqueduct/lib/lambda"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/notification"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	pollAuthenticateInterval = 500 * time.Millisecond
	pollAuthenticateTimeout  = 2 * time.Minute
)

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

	ArtifactRepo       repos.Artifact
	ArtifactResultRepo repos.ArtifactResult
	DAGRepo            repos.DAG
	IntegrationRepo    repos.Integration
	OperatorRepo       repos.Operator

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

	if statusCode, err := ConnectIntegration(ctx, args, h.IntegrationRepo, txn); err != nil {
		return emptyResp, statusCode, err
	}

	if args.Service == shared.AWS {
		cloudIntegration, err := h.IntegrationRepo.GetByNameAndUser(
			ctx,
			args.Name,
			uuid.Nil,
			args.OrgID,
			txn,
		)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve cloud integration.")
		}

		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "dynamic", "kube_config")
		// Register a dynamic k8s integration.
		connectIntegrationArgs := &ConnectIntegrationArgs{
			AqContext: args.AqContext,
			Name:      fmt.Sprintf("%s:%s", args.Name, "k8s"),
			Service:   shared.Kubernetes,
			Config: auth.NewStaticConfig(
				map[string]string{
					shared.K8sKubeconfigPathKey:     kubeconfigPath,
					shared.K8sClusterNameKey:        shared.DynamicK8sClusterName,
					shared.K8sDynamicKey:            strconv.FormatBool(true),
					shared.K8sCloudIntegrationIdKey: cloudIntegration.ID.String(),
					shared.K8sUseSameClusterKey:     strconv.FormatBool(false),
					shared.K8sStatusKey:             string(shared.K8sClusterTerminatedStatus),
					shared.K8sKeepaliveKey:          strconv.FormatInt(int64(shared.DefaultKeepalive), 10),
				},
			),
			UserOnly:     false,
			SetAsStorage: false,
		}

		_, _, err = (&ConnectIntegrationHandler{
			Database:   txn,
			JobManager: h.JobManager,

			ArtifactRepo:       h.ArtifactRepo,
			ArtifactResultRepo: h.ArtifactResultRepo,
			DAGRepo:            h.DAGRepo,
			IntegrationRepo:    h.IntegrationRepo,
			OperatorRepo:       h.OperatorRepo,
		}).Perform(ctx, connectIntegrationArgs)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to register dynamic k8s integration.")
		}

		if _, _, err := lib_utils.RunCmd("terraform", []string{"init"}, dynamic.TerraformDir, true); err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Error initializing Terraform")
		}
	}

	if err := txn.Commit(ctx); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}

	if args.SetAsStorage {
		// This integration should be used as the new storage layer.
		// In order to do so, we need to migrate all content from the old store
		// to the new store. This requires pausing the server and then restarting it.
		// All of this logic is performed asynchronously so that the user knows that
		// the connect integration request has succeeded and that the migration is now
		// under way.
		go func() {
			log.Info("Starting storage migration process...")
			// Wait until the server is paused
			h.PauseServer()
			// Makes sure that the server is restarted
			defer h.RestartServer()

			// Wait until there are no more workflow runs in progress
			lock := utils.NewExecutionLock()
			if err := lock.Lock(); err != nil {
				log.Errorf("Unexpected error when acquiring workflow execution lock: %v. Aborting storage migration!", err)
				return
			}
			defer func() {
				if err := lock.Unlock(); err != nil {
					log.Errorf("Unexpected error when unlocking workflow execution lock: %v", err)
				}
			}()

			if err := setIntegrationAsStorage(
				context.Background(),
				args.Service,
				args.Config,
				args.OrgID,
				h.DAGRepo,
				h.ArtifactRepo,
				h.ArtifactResultRepo,
				h.OperatorRepo,
				h.IntegrationRepo,
				h.Database,
			); err != nil {
				log.Errorf("Unexpected error when setting the new storage layer: %v", err)
			}

			log.Info("Successfully migrated the storage layer!")
		}()
	}

	return emptyResp, http.StatusOK, nil
}

// ConnectIntegration connects a new integration specified by `args`. It returns a status code for the request
// and an error, if any.
func ConnectIntegration(
	ctx context.Context,
	args *ConnectIntegrationArgs,
	integrationRepo repos.Integration,
	DB database.Database,
) (int, error) {
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
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to initialize vault.")
	}

	// Store config (including confidential information) in vault
	if err := auth.WriteConfigToSecret(
		ctx,
		integrationObject.ID,
		args.Config,
		vaultObject,
	); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
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

	return http.StatusOK, nil
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
		// Lambda authentication is performed by creating Lambda jobs
		// instead of the Python client, so we don't launch a job for it.
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
		go utils.CleanupStorageFiles(ctx, storageConfig, []string{jobMetadataPath})
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

// setIntegrationAsStorage use the integration config `conf` and updates the global
// storage config with it. This involves migrating the storage (and vault) content to the new
// storage layer.
func setIntegrationAsStorage(
	ctx context.Context,
	svc shared.Service,
	conf auth.Config,
	orgID string,
	dagRepo repos.DAG,
	artifactRepo repos.Artifact,
	artifactResultRepo repos.ArtifactResult,
	operatorRepo repos.Operator,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	data, err := conf.Marshal()
	if err != nil {
		return err
	}

	var storageConfig *shared.StorageConfig

	switch svc {
	case shared.S3:
		var c shared.S3IntegrationConfig
		if err := json.Unmarshal(data, &c); err != nil {
			return err
		}

		storageConfig, err = convertS3IntegrationtoStorageConfig(&c)
		if err != nil {
			return err
		}
	case shared.GCS:
		var c shared.GCSIntegrationConfig
		if err := json.Unmarshal(data, &c); err != nil {
			return err
		}

		storageConfig = convertGCSIntegrationtoStorageConfig(&c)
	default:
		return errors.Newf("%v cannot be used as the metadata storage layer", svc)
	}

	currentStorageConfig := config.Storage()

	// Migrate all storage content to the new storage config
	if err := utils.MigrateStorageAndVault(
		ctx,
		&currentStorageConfig,
		storageConfig,
		orgID,
		dagRepo,
		artifactRepo,
		artifactResultRepo,
		operatorRepo,
		integrationRepo,
		db,
	); err != nil {
		return err
	}

	// Change global storage config
	return config.UpdateStorage(storageConfig)
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

// func validateLambdaConfig(
// 	ctx context.Context,
// 	config auth.Config,
// ) (int, error) {
// 	if err := engine.AuthenticateLambdaConfig(ctx, config); err != nil {
// 		return http.StatusBadRequest, err
// 	}

// 	return http.StatusOK, nil
// }

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
	if err != database.ErrNoRows {
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
			return http.StatusBadRequest, errors.Wrap(err, "terraform executable not found. Please go to https://developer.hashicorp.com/terraform/downloads to install terraform")
		}

		if _, _, err := lib_utils.RunCmd("aws", []string{"--version"}, "", false); err != nil {
			return http.StatusBadRequest, errors.Wrap(err, "AWS CLI executable not found. Please go to https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html to install AWS CLI")
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
