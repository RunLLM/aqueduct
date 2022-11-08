package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/airflow"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	postgres_utils "github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/job"
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

var ErrNoIntegrationName = errors.New("Integration name is not provided")

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
	Vault      vault.Vault
	JobManager job.JobManager

	WorkflowDagReader    workflow_dag.Reader
	ArtifactReader       artifact.Reader
	ArtifactResultReader artifact_result.Reader
	OperatorReader       operator.Reader
	IntegrationReader    integration.Reader
	WorkflowDagWriter    workflow_dag.Writer
	IntegrationWriter    integration.Writer

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
	Name         string              // User specified name for the integration
	Service      integration.Service // Name of the service to connect (e.g. Snowflake, Postgres)
	Config       auth.Config         // Integration config
	UserOnly     bool                // Whether the integration is only accessible by the user or the entire org
	SetAsStorage bool                // Whether the integration should be used as the storage layer
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
		return nil, http.StatusBadRequest, ErrNoIntegrationName
	}

	if service == integration.Github || service == integration.GoogleSheets {
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

	// Validate integration config
	statusCode, err := ValidateConfig(
		ctx,
		args.RequestId,
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

	if statusCode, err := ConnectIntegration(ctx, args, h.IntegrationWriter, txn, h.Vault); err != nil {
		return emptyResp, statusCode, err
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
				args.OrganizationId,
				h.WorkflowDagReader,
				h.WorkflowDagWriter,
				h.ArtifactReader,
				h.ArtifactResultReader,
				h.OperatorReader,
				h.IntegrationReader,
				h.Database,
			); err != nil {
				log.Errorf("Unexpected error when setting the new storage layer: %v", err)
			}
		}()
	}

	return emptyResp, http.StatusOK, nil
}

// ConnectIntegration connects a new integration specified by `args`. It returns a status code for the request
// and an error, if any.
func ConnectIntegration(
	ctx context.Context,
	args *ConnectIntegrationArgs,
	integrationWriter integration.Writer,
	db database.Database,
	vaultObject vault.Vault,
) (int, error) {
	// Extract non-confidential config
	publicConfig := args.Config.PublicConfig()

	var integrationObject *integration.Integration
	var err error
	if args.UserOnly {
		// This is a user-specific integration
		integrationObject, err = integrationWriter.CreateIntegrationForUser(
			ctx,
			args.OrganizationId,
			args.Id,
			args.Service,
			args.Name,
			(*postgres_utils.Config)(&publicConfig),
			true,
			db,
		)
	} else {
		integrationObject, err = integrationWriter.CreateIntegration(
			ctx,
			args.OrganizationId,
			args.Service,
			args.Name,
			(*postgres_utils.Config)(&publicConfig),
			true,
			db,
		)
	}
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}

	// Store config (including confidential information) in vault
	if err := auth.WriteConfigToSecret(
		ctx,
		integrationObject.Id,
		args.Config,
		vaultObject,
	); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}

	return http.StatusOK, nil
}

// ValidateConfig authenticates the config provided.
// It returns a status code and an error, if any.
func ValidateConfig(
	ctx context.Context,
	requestId string,
	config auth.Config,
	service integration.Service,
	jobManager job.JobManager,
	storageConfig *shared.StorageConfig,
) (int, error) {
	if service == integration.Airflow {
		// Airflow authentication is performed via the Go client
		// instead of the Python client, so we don't launch a job for it.
		return validateAirflowConfig(ctx, config)
	}

	if service == integration.Kubernetes {
		// Kuerbnetes authentication is performed via initializing a k8s client
		// instead of the Python client, so we don't launch a job for it.
		return validateKubernetesConfig(ctx, config)
	}

	if service == integration.Lambda {
		// Lambda authentication is performed by creating Lambda jobs
		// instead of the Python client, so we don't launch a job for it.
		return validateLambdaConfig(ctx, config)
	}

	jobName := fmt.Sprintf("authenticate-operator-%s", uuid.New().String())
	if service == integration.Conda {
		return validateConda(ctx, jobManager, jobName)
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
func checkIntegrationSetStorage(svc integration.Service, conf auth.Config) (bool, error) {
	if svc != integration.S3 && svc != integration.GCS {
		// Only S3 and GCS can be used for storage
		return false, nil
	}

	data, err := conf.Marshal()
	if err != nil {
		return false, err
	}

	switch svc {
	case integration.S3:
		var c integration.S3Config
		if err := json.Unmarshal(data, &c); err != nil {
			return false, err
		}
		return bool(c.UseAsStorage), nil
	case integration.GCS:
		var c integration.GCSConfig
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
	svc integration.Service,
	conf auth.Config,
	orgID string,
	dagReader workflow_dag.Reader,
	dagWriter workflow_dag.Writer,
	artifactReader artifact.Reader,
	artifactResultReader artifact_result.Reader,
	operatorReader operator.Reader,
	integrationReader integration.Reader,
	db database.Database,
) error {
	data, err := conf.Marshal()
	if err != nil {
		return err
	}

	var storageConfig *shared.StorageConfig

	switch svc {
	case integration.S3:
		var c integration.S3Config
		if err := json.Unmarshal(data, &c); err != nil {
			return err
		}

		storageConfig, err = convertS3IntegrationtoStorageConfig(&c)
		if err != nil {
			return err
		}
	case integration.GCS:
		var c integration.GCSConfig
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
		dagReader,
		dagWriter,
		artifactReader,
		artifactResultReader,
		operatorReader,
		integrationReader,
		db,
	); err != nil {
		return err
	}

	// Change global storage config
	return config.UpdateStorage(storageConfig)
}

func convertS3IntegrationtoStorageConfig(c *integration.S3Config) (*shared.StorageConfig, error) {
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
	case integration.AccessKeyS3ConfigType:
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
	case integration.ConfigFileContentS3ConfigType:
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
	case integration.ConfigFilePathS3ConfigType:
		// The credentials are already in the form of a filepath and profile, so no changes
		// need to be made
		storageConfig.S3Config.CredentialsPath = c.ConfigFilePath
		storageConfig.S3Config.CredentialsProfile = c.ConfigFileProfile
	default:
		return nil, errors.Newf("Unknown S3ConfigType: %v", c.Type)
	}

	return storageConfig, nil
}

func convertGCSIntegrationtoStorageConfig(c *integration.GCSConfig) *shared.StorageConfig {
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

func validateLambdaConfig(
	ctx context.Context,
	config auth.Config,
) (int, error) {
	if err := engine.AuthenticateLambdaConfig(ctx, config); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

func validateConda(
	ctx context.Context,
	jobManager job.JobManager,
	jobName string,
) (int, error) {
	errMsg := "Unable to validate conda installation. Do you have conda installed?"
	if err := exec.Command("conda", "--version").Run(); err != nil {
		return http.StatusBadRequest, errors.Wrap(err, errMsg)
	}

	return http.StatusOK, nil
}
