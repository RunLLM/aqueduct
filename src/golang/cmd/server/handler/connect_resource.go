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

// Route: /resource/connect
// Method: POST
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//		`resource-name`: the name for the resource
//		`resource-service`: the service type for the resource
//		`resource-config`: the json-serialized resource config
//
// Response: none
//
// If this route finishes successfully, then an resource entry is guaranteed to have been created
// in the database.
type ConnectResourceHandler struct {
	PostHandler

	Database   database.Database
	JobManager job.JobManager

	ArtifactRepo         repos.Artifact
	ArtifactResultRepo   repos.ArtifactResult
	DAGRepo              repos.DAG
	ResourceRepo         repos.Resource
	StorageMigrationRepo repos.StorageMigration
	OperatorRepo         repos.Operator

	PauseServer   func()
	RestartServer func()
}

func (*ConnectResourceHandler) Headers() []string {
	return []string{
		routes.ResourceNameHeader,
		routes.ResourceServiceHeader,
		routes.ResourceConfigHeader,
	}
}

type ConnectResourceArgs struct {
	*aq_context.AqContext
	Name         string         // User specified name for the resource
	Service      shared.Service // Name of the service to connect (e.g. Snowflake, Postgres)
	Config       auth.Config    // Resource config
	UserOnly     bool           // Whether the resource is only accessible by the user or the entire org
	SetAsStorage bool           // Whether the resource should be used as the storage layer
}

type ConnectResourceResponse struct{}

func (*ConnectResourceHandler) Name() string {
	return "ConnectResource"
}

func (h *ConnectResourceHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Unable to connect resource.")
	}

	service, userOnly, err := request.ParseResourceServiceFromRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to connect resource.")
	}

	name, configMap, err := request.ParseResourceConfigFromRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to connect resource.")
	}

	if name == "" {
		return nil, http.StatusBadRequest, errors.New("Resource name is not provided")
	}

	// On startup, we currently enforce that such a resource does not exist by forcibly deleting it.
	// Therefore, we don't want users to be able to create a resource with this name.
	if name == shared.DeprecatedDemoDBResourceName && service == shared.Sqlite {
		return nil, http.StatusBadRequest, errors.Newf("%s is a reserved name for SQLite resources.", shared.DeprecatedDemoDBResourceName)
	}

	if service == shared.Github || service == shared.GoogleSheets {
		return nil, http.StatusBadRequest, errors.Newf("%s resource type is currently not supported", service)
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

	// Check if this resource should be used as the new storage layer
	setStorage, err := checkIfUseResourceAsStorage(service, staticConfig)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to connect resource.")
	}

	return &ConnectResourceArgs{
		AqContext:    aqContext,
		Service:      service,
		Name:         name,
		Config:       staticConfig,
		UserOnly:     userOnly,
		SetAsStorage: setStorage,
	}, http.StatusOK, nil
}

func (h *ConnectResourceHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*ConnectResourceArgs)

	emptyResp := ConnectResourceResponse{}

	statusCode, err := ValidatePrerequisites(
		ctx,
		args.Service,
		args.Name,
		args.ID,
		args.OrgID,
		args.Config,
		h.ResourceRepo,
		h.Database,
	)
	if err != nil {
		return emptyResp, statusCode, err
	}

	// Validate resource config
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

	// Assumption: we are always ADDING a new resource, so `resourceObj` must be a freshly created resource entry.
	// Note that the config of this returned `resourceObj` may be outdated.
	resourceObj, statusCode, err := ConnectResource(ctx, h, args, h.ResourceRepo, h.Database)
	if err != nil {
		return emptyResp, statusCode, err
	}

	if args.SetAsStorage {
		confData, err := args.Config.Marshal()
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		newStorageConfig, err := storage.ConvertResourceConfigToStorageConfig(args.Service, confData)
		if err != nil {
			return emptyResp, http.StatusBadRequest, errors.Wrap(err, "Resource config is malformed.")
		}

		err = storage_migration.Perform(
			ctx,
			args.OrgID,
			resourceObj,
			newStorageConfig,
			h.PauseServer,
			h.RestartServer,
			h.ArtifactRepo,
			h.ArtifactResultRepo,
			h.DAGRepo,
			h.ResourceRepo,
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

// ConnectResource connects a new resource specified by `args`.
// It returns the resource object, the status code for the request and an error, if any.
// If an error is returns, the resource object is guaranteed to be nil. Conversely, the resource
// object is always well-formed on success.
func ConnectResource(
	ctx context.Context,
	h *ConnectResourceHandler, // This only needs to be non-nil if the resource can be AWS.
	args *ConnectResourceArgs,
	resourceRepo repos.Resource,
	DB database.Database,
) (_ *models.Resource, _ int, err error) {
	// Extract non-confidential config
	publicConfig := args.Config.PublicConfig()

	// Always create the resource entry with a running state to start.
	runningAt := time.Now()
	publicConfig[exec_env.ExecStateKey] = execution_state.SerializedRunning(&runningAt)

	// Must open a transaction to write the initial resource state, because the AWS resource
	// may need to perform multiple writes.
	txn, err := DB.BeginTx(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to connect resource.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	var resourceObject *models.Resource
	if args.UserOnly {
		// This is a user-specific resource
		resourceObject, err = resourceRepo.CreateForUser(
			ctx,
			args.OrgID,
			args.ID,
			args.Service,
			args.Name,
			(*shared.ResourceConfig)(&publicConfig),
			txn,
		)
	} else {
		resourceObject, err = resourceRepo.Create(
			ctx,
			args.OrgID,
			args.Service,
			args.Name,
			(*shared.ResourceConfig)(&publicConfig),
			txn,
		)
	}
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to connect resource.")
	}

	if args.Service == shared.AWS {
		if h == nil {
			return nil, http.StatusInternalServerError, errors.New("Internal error: No route handler present when registering an AWS resource.")
		}
		if statusCode, err := setupCloudResource(
			ctx,
			args,
			h,
			txn,
		); err != nil {
			return nil, statusCode, err
		}
	}
	if err := txn.Commit(ctx); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to connect resource.")
	}

	// The initial resource entry has been written. Any errors from this point on will need to update
	// the that entry to reflect the failure.
	defer func() {
		if err != nil {
			execution_state.UpdateOnFailure(
				ctx,
				"", // outputs
				err.Error(),
				string(args.Service),
				(*shared.ResourceConfig)(&publicConfig),
				&runningAt,
				resourceObject.ID,
				resourceRepo,
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
		resourceObject.ID,
		args.Config,
		vaultObject,
	); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to connect resource.")
	}

	// For those resources that require asynchronous setup, we spin those up here. When those goroutines are
	// complete, they write their results back to the config column of their resource entry.
	// Note that kicking off any asynchronous setup is the last thing this method does. This ensures that there
	// will never be any status update races between the goroutines and the main thread.
	// TODO(ENG-2523): move base conda env creation outside of ConnectResource.
	if args.Service == shared.Conda {
		go func() {
			// We must copy the Database inside the goroutine, because the underlying DB connection
			// will error if passed between the main thread and goroutine.
			condaDB, err := database.NewDatabase(DB.Config())
			if err != nil {
				log.Errorf("Error creating DB for Conda: %v", err)
				return
			}

			condaErr := setupCondaAsync(resourceRepo, resourceObject.ID, publicConfig, runningAt, condaDB)
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

			lambdaErr := setupLambdaAsync(resourceRepo, resourceObject.ID, publicConfig, runningAt, lambdaDB)
			if lambdaErr != nil {
				log.Errorf("Lambda setup failed: %v", lambdaErr)
			}
		}()
	} else {
		// No asynchronous setup is needed for these services, so we can simply mark the connection entries as successful.
		err = execution_state.UpdateOnSuccess(
			ctx,
			string(args.Service),
			(*shared.ResourceConfig)(&publicConfig),
			&runningAt,
			resourceObject.ID,
			resourceRepo,
			DB,
		)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
	}
	return resourceObject, http.StatusOK, nil
}

// Asynchronously setup the lambda resource.
func setupLambdaAsync(
	resourceRepo repos.Resource,
	resourceID uuid.UUID,
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
				(*shared.ResourceConfig)(&publicConfig),
				&runningAt,
				resourceID,
				resourceRepo,
				DB,
			)
		} else {
			_ = execution_state.UpdateOnSuccess(
				context.Background(),
				string(shared.Lambda),
				(*shared.ResourceConfig)(&publicConfig),
				&runningAt,
				resourceID,
				resourceRepo,
				DB,
			)
		}
	}()

	return lambda_utils.ConnectToLambda(
		context.Background(),
		publicConfig[lambda_utils.RoleArnKey],
	)
}

// Asynchronously setup the conda resource.
func setupCondaAsync(
	resourceRepo repos.Resource,
	resourceID uuid.UUID,
	publicConfig map[string]string,
	runningAt time.Time,
	DB database.Database,
) (err error) {
	var condaPath string
	var output string
	defer func() {
		// Update both the conda path and execution state of the resource's config.
		publicConfig[exec_env.CondaPathKey] = condaPath

		if err != nil {
			execution_state.UpdateOnFailure(
				context.Background(),
				output,
				err.Error(),
				string(shared.Conda),
				(*shared.ResourceConfig)(&publicConfig),
				&runningAt,
				resourceID,
				resourceRepo,
				DB,
			)
		} else {
			// Update the conda execution state to be successful.
			_ = execution_state.UpdateOnSuccess(
				context.Background(),
				string(shared.Conda),
				(*shared.ResourceConfig)(&publicConfig),
				&runningAt,
				resourceID,
				resourceRepo,
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

	if service == shared.GAR {
		return validateGARConfig(config)
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
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to connect resource.")
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
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to connect resource.")
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
		return http.StatusBadRequest, errors.Wrap(err, "Unable to authenticate Airflow credentials. Please check them.")
	}

	return http.StatusOK, nil
}

// checkIfUseResourceAsStorage returns whether this resource should be used as the storage layer.
func checkIfUseResourceAsStorage(svc shared.Service, conf auth.Config) (bool, error) {
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
		var c shared.S3ResourceConfig
		if err := json.Unmarshal(data, &c); err != nil {
			return false, err
		}
		return bool(c.UseAsStorage), nil
	case shared.GCS:
		var c shared.GCSResourceConfig
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
	if err := engine.AuthenticateAndUpdateK8sConfig(ctx, config); err != nil {
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

func validateGARConfig(
	config auth.Config,
) (int, error) {
	if err := container_registry.AuthenticateGARConfig(config); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

// ValidatePrerequisites validates if the resource for the given service can be connected at all.
// 1) Checks if an resource already exists for unique resources including conda, email, and slack.
// 2) Checks if the name has already been taken.
func ValidatePrerequisites(
	ctx context.Context,
	svc shared.Service,
	name string,
	userID uuid.UUID,
	orgID string,
	conf auth.Config,
	resourceRepo repos.Resource,
	DB database.Database,
) (int, error) {
	// We expect the new name to be unique.
	_, err := resourceRepo.GetByNameAndUser(ctx, name, userID, orgID, DB)
	if err == nil {
		return http.StatusBadRequest, errors.Newf("Cannot connect to an resource %s, since it already exists.", name)
	}
	if !errors.Is(err, database.ErrNoRows()) {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to query for existing resources.")
	}

	if svc == shared.Conda {
		condaResource, err := exec_env.GetCondaResource(
			ctx, userID, resourceRepo, DB,
		)
		if err != nil {
			return http.StatusInternalServerError, errors.Wrap(err, "Unable to verify if conda is connected.")
		}

		if condaResource != nil {
			return http.StatusBadRequest, errors.Newf(
				"You already have conda resource %s connected.",
				condaResource.Name,
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

	if svc != shared.Conda && shared.IsComputeResource(svc) {
		// For all non-conda compute resources, we require the metadata store to be cloud storage.
		if config.Storage().Type == shared.FileStorageType {
			return http.StatusBadRequest, errors.Newf("You need to setup cloud storage as metadata store before registering compute resource of type %s.", svc)
		}
	}

	// These resources should be unique.
	if svc == shared.Email || svc == shared.Slack {
		resources, err := resourceRepo.GetByServiceAndUser(ctx, svc, userID, DB)
		if err != nil {
			return http.StatusInternalServerError, errors.Wrap(err, "Unable to verify if email is connected.")
		}

		if len(resources) > 0 {
			return http.StatusBadRequest, errors.Newf(
				"You already have an %s resource %s connected.",
				svc,
				resources[0].Name,
			)
		}

		return http.StatusOK, nil
	}

	// For AWS resource, we require the user to have AWS CLI and Terraform installed.
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

	// For ECR resource, we require the user to have AWS CLI installed.
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

	// For on-demand GKE resource, we require the user to have Terraform, gcloud, and gcloud gke-gcloud-auth-plugin plugin installed.
	if svc == shared.Kubernetes {
		// Parse the config and see if cloud provider is GCP
		k8sConf, err := lib_utils.ParseK8sConfig(conf)
		if err != nil {
			return http.StatusInternalServerError, errors.Wrap(err, "Unable to parse Kubernetes configuration.")
		}

		if k8sConf.CloudProvider == shared.GCPProvider {
			if _, _, err := lib_utils.RunCmd("terraform", []string{"--version"}, "", false); err != nil {
				return http.StatusNotFound, errors.Wrap(err, "terraform executable not found. Please go to https://developer.hashicorp.com/terraform/downloads to install terraform")
			}

			_, _, err := lib_utils.RunCmd("gcloud", []string{"--version"}, "", false)
			if err != nil {
				return http.StatusNotFound, errors.Wrap(err, "gcloud executable not found. Please go to https://cloud.google.com/sdk/docs/install to install gcloud")
			}

			componentsOutput, _, err := lib_utils.RunCmd("gcloud", []string{"components", "list"}, "", false)
			if err != nil {
				return http.StatusUnprocessableEntity, errors.Wrap(err, "Error listing gcloud components")
			}

			if !strings.Contains(componentsOutput, "gke-gcloud-auth-plugin") {
				return http.StatusUnprocessableEntity, errors.New("gke-gcloud-auth-plugin is not installed. Please run `gcloud components install gke-gcloud-auth-plugin` to install it.")
			}
		}
	}

	// For GAR integration, we require the user to have gcloud installed.
	if svc == shared.GAR {
		_, _, err := lib_utils.RunCmd("gcloud", []string{"--version"}, "", false)
		if err != nil {
			return http.StatusNotFound, errors.Wrap(err, "gcloud executable not found. Please go to https://cloud.google.com/sdk/docs/install to install gcloud")
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
