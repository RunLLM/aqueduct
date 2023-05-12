package engine

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	databricks_lib "github.com/aqueducthq/aqueduct/lib/databricks"
	"github.com/aqueducthq/aqueduct/lib/dynamic"
	"github.com/aqueducthq/aqueduct/lib/k8s"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/spark"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
)

// Authenticates kubernetes configuration by trying to connect a client.
// In case of on-demand k8s resource, updates the k8s config with the
// cluster config parameters.
func AuthenticateAndUpdateK8sConfig(ctx context.Context, authConf auth.Config) error {
	conf, err := lib_utils.ParseK8sConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}

	if conf.Dynamic {
		// The following code path is currently reserved for AWS. Need to refactor it to be consistent
		// with GCP.
		if conf.CloudIntegrationId == "" {
			return errors.New("Dynamic K8s integration must have a cloud integration ID attached.")
		} else {
			return nil
		}
	}

	if conf.CloudProvider == shared.GCPProvider {
		// This is an on-demand GKE resource.
		k8sConfig := shared.DynamicK8sConfig{
			Keepalive:   strconv.Itoa(shared.K8sDefaultKeepalive),
			CpuNodeType: shared.GKEDefaultCpuNodeType,
			GpuNodeType: shared.GKEDefaultGpuNodeType,
			MinCpuNode:  strconv.Itoa(shared.K8sDefaultMinCpuNode),
			MaxCpuNode:  strconv.Itoa(shared.K8sDefaultMaxCpuNode),
			MinGpuNode:  strconv.Itoa(shared.K8sDefaultMinGpuNode),
			MaxGpuNode:  strconv.Itoa(shared.K8sDefaultMaxGpuNode),
		}

		// Parse authconf to shared.DynamicK8sConfig
		data, err := authConf.Marshal()
		if err != nil {
			return err
		}

		var customConfig shared.DynamicK8sConfig
		if err := json.Unmarshal(data, &customConfig); err != nil {
			return err
		}

		k8sConfig.Update(&customConfig)

		clusterName, err := dynamic.GenerateClusterNameGKE()
		if err != nil {
			return errors.Wrap(err, "Unable to generate k8s cluster name.")
		}

		terraformPath := filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "ondemand_k8s", clusterName)
		if err = dynamic.SetupTerraformDirectory(dynamic.GKETerraformTemplateDir, terraformPath); err != nil {
			return errors.Wrap(err, "Unable to create Terraform directory.")
		}

		if _, _, err := lib_utils.RunCmd("terraform", []string{"init"}, terraformPath, true); err != nil {
			return errors.Wrap(err, "Error initializing Terraform")
		}

		kubeconfigPath := filepath.Join(terraformPath, "kube_config")

		dynamicK8sConfig := map[string]string{
			shared.K8sTerraformPathKey:  terraformPath,
			shared.K8sKubeconfigPathKey: kubeconfigPath,
			shared.K8sClusterNameKey:    clusterName,
			shared.K8sDynamicKey:        strconv.FormatBool(true),
			shared.K8sUseSameClusterKey: strconv.FormatBool(false),
			shared.K8sStatusKey:         string(shared.K8sClusterTerminatedStatus),
			shared.K8sDesiredCpuNodeKey: k8sConfig.MinCpuNode,
			shared.K8sDesiredGpuNodeKey: k8sConfig.MinGpuNode,
		}

		for k, v := range k8sConfig.ToMap() {
			dynamicK8sConfig[k] = v
		}

		if err := dynamic.CheckIfValidConfig(dynamic.K8sClusterCreateAction, dynamicK8sConfig); err != nil {
			return err
		}

		// Update the authConf with the dynamicK8sConfig
		castedConf := authConf.(*auth.StaticConfig)
		for k, v := range dynamicK8sConfig {
			castedConf.Set(k, v)
		}

		return nil
	}

	return k8s.ValidateCluster(ctx, conf.ClusterName, conf.KubeconfigPath, bool(conf.UseSameCluster))
}

func AuthenticateDatabricksConfig(ctx context.Context, authConf auth.Config) error {
	databricksConfig, err := lib_utils.ParseDatabricksConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}

	databricksClient, err := databricks_lib.NewWorkspaceClient(
		databricksConfig.WorkspaceURL,
		databricksConfig.AccessToken,
	)
	if err != nil {
		return errors.Wrap(err, "Unable to create Databricks Workspace Client.")
	}
	_, err = databricks_lib.ListJobs(
		ctx,
		databricksClient,
	)
	if err != nil {
		return errors.Wrap(err, "Unable to list Databricks Jobs.")
	}

	err = databricks_lib.AddEntrypointFilesToStorage(ctx)
	if err != nil {
		return errors.Wrap(err, "Unable to upload entrypoint files to storage.")
	}

	return nil
}

func AuthenticateSparkConfig(ctx context.Context, authConf auth.Config) error {
	sparkConfig, err := lib_utils.ParseSparkConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}

	livyClient := spark.NewLivyClient(sparkConfig.LivyServerURL)
	_, err = livyClient.GetSessions()
	if err != nil {
		return errors.Wrap(err, "Unable to list active Sessions on Livy Server.")
	}

	return nil
}

func AuthenticateAWSConfig(authConf auth.Config) error {
	conf, err := lib_utils.ParseAWSConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}

	if conf.AccessKeyId != "" && conf.SecretAccessKey != "" && conf.Region != "" {
		if conf.ConfigFilePath != "" || conf.ConfigFileProfile != "" {
			return errors.New("When authenticating via access keys, credential file path and profile must be empty.")
		}
	} else if conf.ConfigFilePath != "" && conf.ConfigFileProfile != "" {
		if conf.AccessKeyId != "" || conf.SecretAccessKey != "" || conf.Region != "" {
			return errors.New("When authenticating via credential file, access key fields must be empty.")
		}
	} else {
		return errors.New("Either 1) AWS access key ID, secret access key, region, or 2) credential file path, profile must be provided.")
	}

	return nil
}
