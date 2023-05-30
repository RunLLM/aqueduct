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

// setupCloudResource sets up the cloud resource's Terraform directory, registers a k8s
// resource and run `terraform init` to initialize the Terraform workspace.
func setupCloudResource(
	ctx context.Context,
	args *ConnectResourceArgs,
	h *ConnectResourceHandler,
	db database.Database,
) (int, error) {
	cloudResource, err := h.ResourceRepo.GetByNameAndUser(
		ctx,
		args.Name,
		uuid.Nil,
		args.OrgID,
		db,
	)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve cloud resource.")
	}

	terraformPath := filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "cloud_resource", args.Name, "eks")
	if err = dynamic.SetupTerraformDirectory(dynamic.EKSTerraformTemplateDir, terraformPath); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to create Terraform directory.")
	}

	kubeconfigPath := filepath.Join(terraformPath, "kube_config")

	awsConfig, err := lib_utils.ParseAWSConfig(args.Config)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to parse AWS config.")
	}

	config := shared.DynamicK8sConfig{
		Keepalive:   strconv.Itoa(shared.K8sDefaultKeepalive),
		CpuNodeType: shared.EKSDefaultCpuNodeType,
		GpuNodeType: shared.EKSDefaultGpuNodeType,
		MinCpuNode:  strconv.Itoa(shared.K8sDefaultMinCpuNode),
		MaxCpuNode:  strconv.Itoa(shared.K8sDefaultMaxCpuNode),
		MinGpuNode:  strconv.Itoa(shared.K8sDefaultMinGpuNode),
		MaxGpuNode:  strconv.Itoa(shared.K8sDefaultMaxGpuNode),
	}

	config.Update(awsConfig.K8s)

	clusterName, err := dynamic.GenerateClusterName()
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to generate k8s cluster name.")
	}

	dynamicK8sConfig := map[string]string{
		shared.K8sTerraformPathKey:   terraformPath,
		shared.K8sKubeconfigPathKey:  kubeconfigPath,
		shared.K8sClusterNameKey:     clusterName,
		shared.K8sDynamicKey:         strconv.FormatBool(true),
		shared.K8sCloudResourceIdKey: cloudResource.ID.String(),
		shared.K8sUseSameClusterKey:  strconv.FormatBool(false),
		shared.K8sStatusKey:          string(shared.K8sClusterTerminatedStatus),
		shared.K8sDesiredCpuNodeKey:  config.MinCpuNode,
		shared.K8sDesiredGpuNodeKey:  config.MinGpuNode,
	}

	for k, v := range config.ToMap() {
		dynamicK8sConfig[k] = v
	}

	if err := dynamic.CheckIfValidConfig(dynamic.K8sClusterCreateAction, dynamicK8sConfig); err != nil {
		return http.StatusBadRequest, err
	}

	// Register a dynamic k8s resource.
	connectResourceArgs := &ConnectResourceArgs{
		AqContext:    args.AqContext,
		Name:         fmt.Sprintf("%s:%s", args.Name, dynamic.K8sResourceNameSuffix),
		Service:      shared.Kubernetes,
		Config:       auth.NewStaticConfig(dynamicK8sConfig),
		UserOnly:     false,
		SetAsStorage: false,
	}

	_, _, err = (&ConnectResourceHandler{
		Database:   db,
		JobManager: h.JobManager,

		ArtifactRepo:       h.ArtifactRepo,
		ArtifactResultRepo: h.ArtifactResultRepo,
		DAGRepo:            h.DAGRepo,
		ResourceRepo:       h.ResourceRepo,
		OperatorRepo:       h.OperatorRepo,
	}).Perform(ctx, connectResourceArgs)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to register dynamic k8s resource.")
	}

	if _, _, err := lib_utils.RunCmd("terraform", []string{"init"}, terraformPath, true); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Error initializing Terraform")
	}

	return http.StatusOK, nil
}

// deleteCloudIntegrationHelper does the following:
// 1. Verifies that there is no workflow using the dynamic k8s integration.
// 2. Deletes the EKS cluster if it's running.
// 3. Deletes the cloud resource directory.
// 4. Deletes the Aqueduct-generated dynamic k8s resource.
func deleteCloudResourceHelper(
	ctx context.Context,
	args *deleteResourceArgs,
	h *DeleteResourceHandler,
) (int, error) {
	k8sResource, err := h.ResourceRepo.GetByNameAndUser(
		ctx,
		fmt.Sprintf("%s:%s", args.resourceObject.Name, dynamic.K8sResourceNameSuffix),
		uuid.Nil,
		args.OrgID,
		h.Database,
	)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve the Aqueduct-generated k8s resource.")
	}

	// Delete the EKS cluster
	editDynamicEngineArgs := &editDynamicEngineArgs{
		AqContext:   args.AqContext,
		action:      forceDeleteAction,
		resourceID:  k8sResource.ID,
		configDelta: &shared.DynamicK8sConfig{},
	}
	_, statusCode, err := (&EditDynamicEngineHandler{
		Database:     h.Database,
		ResourceRepo: h.ResourceRepo,
	}).Perform(ctx, editDynamicEngineArgs)
	if err != nil {
		return statusCode, errors.Wrap(err, "Failed to delete the dynamic k8s cluster.")
	}

	// Clean up the cloud resource directory
	_, stdErr, err := lib_utils.RunCmd("rm", []string{
		"-rf",
		path.Dir(k8sResource.Config[shared.K8sTerraformPathKey]),
	}, // get the parent dir of Terraform path
		"",
		false,
	)
	if err != nil {
		return http.StatusInternalServerError, errors.New(stdErr)
	}

	deleteK8sResourceArgs := &deleteResourceArgs{
		AqContext:      args.AqContext,
		resourceObject: k8sResource,
		// We already validated this above, so we skip the validation during the deletion of the
		// dynamic k8s resource. There may be race conditions where new workflows are deployed
		// while we clean up the EKS cluster, but in these rare cases we just let the user delete
		// the broken workflows themselves afterwards.
		skipActiveWorkflowValidation: true,
	}

	_, statusCode, err = h.Perform(ctx, deleteK8sResourceArgs)
	if err != nil {
		return statusCode, errors.Wrap(err, "Failed to delete the Aqueduct-generated k8s resource.")
	}

	return http.StatusOK, nil
}
