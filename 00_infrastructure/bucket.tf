resource "google_storage_bucket" "aerobic_botany_270918" {
  name          = "aerobic-botany-270918"
  location      = "EU"
  force_destroy = true
}
resource "google_storage_bucket" "aerobic_botany_270918_event_source" {
  name          = "aerobic-botany-270918-source"
  location      = "EU"
  force_destroy = true
}
