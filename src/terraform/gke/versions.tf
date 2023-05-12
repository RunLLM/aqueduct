# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.64.0"
    }

    kubernetes = {
      source = "hashicorp/kubernetes"
    }
  }

  required_version = ">= 0.14"
}