package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/dynamic"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/google/uuid"
)

// setupCloudIntegration sets up the cloud integration's Terraform directory, registers a k8s
// integration and run `terraform init` to initialize the Terraform workspace.
func setupCloudIntegration(
	ctx context.Context,
	args *ConnectIntegrationArgs,
	h *ConnectIntegrationHandler,
	db database.Database,
) (int, error) {
	cloudIntegration, err := h.IntegrationRepo.GetByNameAndUser(
		ctx,
		args.Name,
		uuid.Nil,
		args.OrgID,
		db,
	)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve cloud integration.")
	}

	terraformPath := filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "cloud_integration", args.Name, "eks")
	if err = setupTerraformDirectory(terraformPath); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to create Terraform directory.")
	}

	kubeconfigPath := filepath.Join(terraformPath, "kube_config")

	awsConfig, err := lib_utils.ParseAWSConfig(args.Config)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to parse AWS config.")
	}

	config := shared.DynamicK8sConfig{
		Keepalive:   shared.DefaultDynamicK8sConfig.Keepalive,
		CpuNodeType: shared.DefaultDynamicK8sConfig.CpuNodeType,
		GpuNodeType: shared.DefaultDynamicK8sConfig.GpuNodeType,
		MinCpuNode:  shared.DefaultDynamicK8sConfig.MinCpuNode,
		MaxCpuNode:  shared.DefaultDynamicK8sConfig.MaxCpuNode,
		MinGpuNode:  shared.DefaultDynamicK8sConfig.MinGpuNode,
		MaxGpuNode:  shared.DefaultDynamicK8sConfig.MaxGpuNode,
	}

	config.Update(awsConfig.K8s)

	clusterName, err := dynamic.GenerateClusterName()
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to generate k8s cluster name.")
	}

	dynamicK8sConfig := map[string]string{
		shared.K8sTerraformPathKey:      terraformPath,
		shared.K8sKubeconfigPathKey:     kubeconfigPath,
		shared.K8sClusterNameKey:        clusterName,
		shared.K8sDynamicKey:            strconv.FormatBool(true),
		shared.K8sCloudIntegrationIdKey: cloudIntegration.ID.String(),
		shared.K8sUseSameClusterKey:     strconv.FormatBool(false),
		shared.K8sStatusKey:             string(shared.K8sClusterTerminatedStatus),
		shared.K8sDesiredCpuNodeKey:     config.MinCpuNode,
		shared.K8sDesiredGpuNodeKey:     config.MinGpuNode,
	}

	for k, v := range config.ToMap() {
		dynamicK8sConfig[k] = v
	}

	if err := dynamic.CheckIfValidConfig(dynamic.K8sClusterCreateAction, dynamicK8sConfig); err != nil {
		return http.StatusBadRequest, err
	}

	// Register a dynamic k8s integration.
	connectIntegrationArgs := &ConnectIntegrationArgs{
		AqContext:    args.AqContext,
		Name:         fmt.Sprintf("%s:%s", args.Name, dynamic.K8sIntegrationNameSuffix),
		Service:      shared.Kubernetes,
		Config:       auth.NewStaticConfig(dynamicK8sConfig),
		UserOnly:     false,
		SetAsStorage: false,
	}

	_, _, err = (&ConnectIntegrationHandler{
		Database:   db,
		JobManager: h.JobManager,

		ArtifactRepo:       h.ArtifactRepo,
		ArtifactResultRepo: h.ArtifactResultRepo,
		DAGRepo:            h.DAGRepo,
		IntegrationRepo:    h.IntegrationRepo,
		OperatorRepo:       h.OperatorRepo,
	}).Perform(ctx, connectIntegrationArgs)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to register dynamic k8s integration.")
	}

	if _, _, err := lib_utils.RunCmd("terraform", []string{"init"}, terraformPath, true); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Error initializing Terraform")
	}

	return http.StatusOK, nil
}

// setupTerraformDirectory copies all files and folders in the Terraform template directory to the
// cloud integration's destination directory, which is ~/.aqueduct/server/cloud_integration/<name>/eks.
func setupTerraformDirectory(dst string) error {
	// Create the destination directory if it doesn't exist.
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}

	_, stdErr, err := lib_utils.RunCmd("cp", []string{"-rT", dynamic.TerraformTemplateDir, dst}, "", false)
	if err != nil {
		return errors.New(stdErr)
	}

	return nil
}

// deleteCloudIntegrationHelper does the following:
// 1. Verifies that there is no workflow using the dynamic k8s integration.
// 2. Deletes the EKS cluster if it's running.
// 3. Deletes the cloud integration directory.
// 4. Deletes the implicitly created dynamic k8s integration.
func deleteCloudIntegrationHelper(
	ctx context.Context,
	args *deleteIntegrationArgs,
	h *DeleteIntegrationHandler,
) (int, error) {
	k8sIntegration, err := h.IntegrationRepo.GetByNameAndUser(
		ctx,
		fmt.Sprintf("%s:%s", args.integrationObject.Name, dynamic.K8sClusterNameSuffix),
		uuid.Nil,
		args.OrgID,
		h.Database,
	)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve the implicitly created k8s integration.")
	}

	// If there are active workflows using the dynamic k8s integration, we need to reject the
	// deletion request immediately, without deleting the cluster and Terraform directories.
	if statusCode, err := validateNoActiveWorkflowOnIntegration(
		ctx,
		k8sIntegration.ID,
		h.OperatorRepo,
		h.DAGRepo,
		h.IntegrationRepo,
		h.Database,
	); err != nil {
		return statusCode, err
	}

	// Delete the EKS cluster
	editDynamicEngineArgs := &editDynamicEngineArgs{
		AqContext:     args.AqContext,
		action:        forceDeleteAction,
		integrationId: k8sIntegration.ID,
		configDelta:   &shared.DynamicK8sConfig{},
	}
	_, statusCode, err := (&EditDynamicEngineHandler{
		Database:        h.Database,
		IntegrationRepo: h.IntegrationRepo,
	}).Perform(ctx, editDynamicEngineArgs)
	if err != nil {
		return statusCode, errors.Wrap(err, "Failed to delete the dynamic k8s cluster.")
	}

	// Clean up the cloud integration directory
	_, stdErr, err := lib_utils.RunCmd("rm", []string{
		"-rf",
		path.Dir(k8sIntegration.Config[shared.K8sTerraformPathKey]),
	}, // get the parent dir of Terraform path
		"",
		false,
	)
	if err != nil {
		return http.StatusInternalServerError, errors.New(stdErr)
	}

	deleteK8sIntegrationArgs := &deleteIntegrationArgs{
		AqContext:         args.AqContext,
		integrationObject: k8sIntegration,
		// We already validated this above, so we skip the validation during the deletion of the
		// dynamic k8s integration. There may be race conditions where new workflows are deployed
		// while we clean up the EKS cluster, but in these rare cases we just let the user delete
		// the broken workflows themselves afterwards.
		skipActiveWorkflowValidation: true,
	}

	_, statusCode, err = h.Perform(ctx, deleteK8sIntegrationArgs)
	if err != nil {
		return statusCode, errors.Wrap(err, "Failed to delete the implicitly created k8s integration.")
	}

	return http.StatusOK, nil
}
