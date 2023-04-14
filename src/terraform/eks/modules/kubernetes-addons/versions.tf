terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.72"
    }
    time = {
      source  = "hashicorp/time"
      version = ">= 0.8"
    }
  }
}
