terraform {
  backend "gcs" {
    bucket = "khappygo-infrastructure"
    prefix = "terraform/state"
  }
}
