resource "google_storage_bucket" "khappygo" {
  name          = "khappygo"
  location      = "EU"
  force_destroy = true
}
