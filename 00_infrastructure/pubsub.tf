data "google_iam_policy" "storage_to_pubsub_binding" {
  binding {
    role    = "roles/pubsub.publisher"
    members = ["serviceAccount:${data.google_storage_project_service_account.gcs_account.email_address}"]
  }
}

resource "google_project_iam_binding" "project" {
  role    = "roles/pubsub.publisher"
  members = ["serviceAccount:${data.google_storage_project_service_account.gcs_account.email_address}"]
}
