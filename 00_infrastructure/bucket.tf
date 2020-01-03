resource "google_storage_bucket" "khappygo" {
  name          = "khappygo"
  location      = "EU"
  force_destroy = true
}
resource "google_storage_bucket" "khappygo_event_source" {
  name          = "khappygo-source"
  location      = "EU"
  force_destroy = true
}
