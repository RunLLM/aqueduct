data "google_compute_zones" "available" {
  provider = google

  project = var.project_id
  region  = local.region
}

resource "random_shuffle" "available_zones" {
  input        = data.google_compute_zones.available.names
  result_count = 3
}

locals {
	location = var.regional ? var.region : var.zones[0]
	region = var.regional ? var.region : join("-", slice(split("-", var.zones[0]), 0, 2))

	// For regional cluster - use var.zones if provided, use available otherwise, for zonal cluster use var.zones with first element extracted
 	node_locations = var.regional ? coalescelist(compact(var.zones), sort(random_shuffle.available_zones.result)) : slice(var.zones, 1, length(var.zones))

	node_pool_names         = [for np in toset(var.node_pools) : np.name]
  	node_pools              = zipmap(local.node_pool_names, tolist(toset(var.node_pools)))
}