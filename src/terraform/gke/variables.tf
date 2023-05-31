variable "region" {
  description = "GCP region"
  type        = string
}

variable "zone" {
  description = "GCP Zone"
  type        = string
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
  default = ""
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

variable "initial_node_count" {
  description = "Initial number of nodes in this pool"
  type        = number
  default = 1
}

variable "create_gpu_node_pool" {
  description = "Decide if this resource pool has to be created"
  type        = bool
  default = false
}

variable "disk_size_in_gb" {
  description = "Disk Capacity"
  type        = number
  default = 50
}