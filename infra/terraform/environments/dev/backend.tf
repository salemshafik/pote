terraform {
  backend "gcs" {
    bucket = "pote"
    prefix = "pote/dev/terraform/state"
  }
}
