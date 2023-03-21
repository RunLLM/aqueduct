variable "region" {
  description = "AWS region"
  type        = string
}

variable "access_key" {
  description = "AWS access key ID"
  type        = string
  sensitive   = true
}

variable "secret_key" {
  description = "AWS secret access key"
  type        = string
  sensitive   = true
}

variable "credentials_file" {
  description = "AWS credentials file"
  type        = string
}

variable "profile" {
  description = "AWS profile"
  type        = string
}

variable "cpu_node_type" {
  description = "The EC2 instance type of the CPU node group"
  type        = string
}

variable "gpu_node_type" {
  description = "The EC2 instance type of the GPU node group."
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

variable "cluster_name" {
  description = "The name of the EKS cluster"
  type        = string
}