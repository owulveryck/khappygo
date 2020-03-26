resource "google_container_cluster" "my_cluster" {
  provider = "google-beta"

  name     = "knative-test"
  location = "europe-west4-b"

  # We can't create a cluster with no node pool defined, but we want to only use
  # separately managed node pools. So we create the smallest possible default
  # node pool and immediately delete it.
  remove_default_node_pool = true
  initial_node_count       = 1


  master_auth {
    username = ""
    password = ""

    client_certificate_config {
      issue_client_certificate = false
    }
  }
  # min_master_version = "1.15.7-gke.23"

  addons_config {
    cloudrun_config {
      disabled = true
    }
  }
}

resource "google_container_node_pool" "primary_preemptible_nodes" {
  provider   = "google-beta"
  name       = "my-node-pool"
  location   = "europe-west4-b"
  cluster    = google_container_cluster.my_cluster.name
  node_count = 3

  node_config {
    preemptible  = true
    machine_type = "n1-standard-1"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
      "https://www.googleapis.com/auth/compute",
      "https://www.googleapis.com/auth/devstorage.read_write",
      "https://www.googleapis.com/auth/datastore",
    ]
  }
}

