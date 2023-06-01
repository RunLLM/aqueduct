data "google_container_cluster" "cluster" {
  depends_on = [ google_container_cluster.primary ]
  project = var.project_id
  name = var.cluster_name
  location = var.region
}

data "google_client_config" "default" {}

provider "kubernetes" {
  host                   = data.google_container_cluster.cluster.endpoint
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(data.google_container_cluster.cluster.master_auth[0].cluster_ca_certificate)
}

provider "kubectl" {
  host                   = data.google_container_cluster.cluster.endpoint
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(data.google_container_cluster.cluster.master_auth[0].cluster_ca_certificate)
  load_config_file       = false
}

data "http" "nvidia_driver_installer_manifest" {
     url = "https://raw.githubusercontent.com/GoogleCloudPlatform/container-engine-accelerators/master/nvidia-driver-installer/cos/daemonset-preloaded.yaml"
}

resource "kubectl_manifest" "nvidia_driver_installer" {
     yaml_body = data.http.nvidia_driver_installer_manifest.body
}