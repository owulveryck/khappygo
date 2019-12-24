provider "google-beta" {
  region  = "europe-west4-b"
  project = var.project_id
  version = 3.3
}

provider "google" {
  region  = "europe-west4-b"
  project = var.project_id
  version = 3.3
}
