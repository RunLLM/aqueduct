#-----------------Kubernetes Add-ons----------------------

module "cluster_autoscaler" {
  source = "./cluster-autoscaler"

  count = var.enable_cluster_autoscaler ? 1 : 0

  eks_cluster_version = local.eks_cluster_version
  helm_config         = var.cluster_autoscaler_helm_config
  addon_context       = local.addon_context
}

module "metrics_server" {
  count             = var.enable_metrics_server ? 1 : 0
  source            = "./metrics-server"
  helm_config       = var.metrics_server_helm_config
  addon_context     = local.addon_context
}

module "nvidia_device_plugin" {
  source = "./nvidia-device-plugin"

  count = var.enable_nvidia_device_plugin ? 1 : 0

  helm_config       = var.nvidia_device_plugin_helm_config
  addon_context     = local.addon_context
}
