variable "region" {
  description = "GCP region"
  type        = string
}

variable "zone" {
  description = "GCP zone"
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