package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
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
		Name:         fmt.Sprintf("%s:%s", args.Name, dynamic.K8sClusterNameSuffix),
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
// cloud integration's destination directory.
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
