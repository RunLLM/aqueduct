resource "google_container_node_pool" "primary_nodes" {
  name       = "${var.cluster_name}-cpu-node-pool"
  location = var.region
  cluster    = google_container_cluster.primary.name
  node_locations = [var.zone]
  node_count = var.min_cpu_node

  autoscaling {
    location_policy = "BALANCED"	
    min_node_count = var.min_cpu_node
    max_node_count = var.max_cpu_node
  }

  management {
    auto_repair = true
    auto_upgrade = true
  }
  node_config {
    oauth_scopes = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
      "https://www.googleapis.com/auth/devstorage.read_only"
    ]

    labels = {
      env = var.project_id
    }

    disk_size_gb = var.disk_size_in_gb
    preemptible  = true
    machine_type = var.cpu_node_type
    tags         = ["gke-node", "${var.project_id}-gke"]
    }
    upgrade_settings {
      max_surge = 1
      max_unavailable = 0
      strategy = "SURGE"
    }
}

resource "google_container_node_pool" "gpu_nodes" {
  count = var.create_gpu_node_pool ? 1 : 0
  name       = "${var.cluster_name}-gpu-node-pool"
  location   = var.region
  cluster    = google_container_cluster.primary.name
  node_locations = [var.zone]
  node_count = var.min_gpu_node

  autoscaling {
    location_policy = "BALANCED"
    min_node_count = var.min_gpu_node
    max_node_count = var.max_gpu_node
  }

  management {
    auto_repair = true
    auto_upgrade = true
  }

  node_config {
    oauth_scopes = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
      "https://www.googleapis.com/auth/devstorage.read_only"
    ]

    labels = {
      env = var.project_id
      accelerator = var.gpu_node_type
    }

    disk_size_gb = var.disk_size_in_gb
    preemptible  = true
    machine_type = var.cpu_node_type
    tags         = ["gke-node", "${var.project_id}-gke"]
    metadata = {
      disable-legacy-endpoints = "true"
      gke-node-accelerator     = var.gpu_node_type
    }

    reservation_affinity {
        consume_reservation_type = "ANY_RESERVATION"
        values = []
    }

    guest_accelerator {
        count = 1
        type  = var.gpu_node_type
      }
  }
  upgrade_settings {
    max_surge = 1
    max_unavailable = 0
  }
}
