terraform {
  backend "gcs" {
    bucket = "eventsforcloudrun-next2020-infrastructure"
    prefix = "terraform/state"
  }
}
