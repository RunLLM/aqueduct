data "google_compute_zones" "available" {
  provider = google

  project = var.project_id
  region  = var.region
}