variable "region" {
  description = "GCP region"
  type        = string
}

variable "zones" {
  description = "The zones to host cluster in"
  type        = list(string)
  default = []
}

variable "zone" {
  description = "GCP Zone"
  type        = string
}

variable "regional" {
  description = "Whether is a regional cluster (Zonal cluster if set false)"
  type        = bool
}

variable "secret_key" {
  description = "GCP service account key"
  type        = string
  sensitive   = true
}

variable "cluster_name" {
  description = "The name of the GKE cluster"
  type        = string
}

variable "project_id" {
  description = "The ID of the GCP project"
  type        = string
}

variable "description" {
  description = "Description of the cluster"
  type        = string 
}

variable "cluster_resource_labels" {
  type        = map(string)
  description = "The GCE resource labels (a map of key/value pairs) to be applied to the cluster"
  default     = {}
}

variable "cpu_node_type" {
  description = "The instance type of the CPU node group"
  type        = string
}

variable "gpu_node_type" {
  description = "The instance type of the GPU node group."
  type        = string
}

variable "min_cpu_node" {
  description = "Minimum number of nodes in the CPU node group"
  type        = number
}

variable "max_cpu_node" {
  description = "Maximum number of nodes in the CPU node group"
  type        = number
}

variable "min_gpu_node" {
  description = "Minimum number of nodes in the GPU node group"
  type        = number
}

variable "max_gpu_node" {
  description = "Maximum number of nodes in the GPU node group"
  type        = number
}

variable "disk_type" {
  description = "Type of disk "
  type        = string
}

variable "disk_size_in_gb" {
  description = "Disk Capacity"
  type        = number
}

variable "max_pods_per_node" {
  description = "Maximum number of pods per node"
  type        = number
}

variable "initial_node_count" {
  description = "Initial number of nodes in this pool"
  type        = number
}

variable "node_count" {
  description = "Number of nodes in this pool"
  type        = number
}

variable "create_gpu_node_pool" {
  description = "Decide if this resource pool has to be created"
  type        = bool
}

variable "node_pools" {
  type        = list(map(any))
  description = "List of maps containing node pools"

  default = [
    {
      name = "default-node-pool"
    },
  ]
}


variable "compute_engine_service_account" {
  description = "Service account to associate to the nodes in the cluster"
}
