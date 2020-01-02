resource "google_project_service" "k8s_api" {
  service = "container.googleapis.com"

  disable_dependent_services = true
}

resource "google_project_service" "cloudresourcemanager_api" {
  service = "cloudresourcemanager.googleapis.com"

  disable_dependent_services = true
}

resource "google_project_service" "cloudbbuild_api" {
  service = "cloudbuild.googleapis.com"

  disable_dependent_services = true
}


