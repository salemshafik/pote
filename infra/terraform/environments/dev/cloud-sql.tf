resource "google_sql_database" "database" {
  name     = var.dev_database_name
  instance = google_sql_database_instance.dev_instance.name
}

resource "google_compute_network" "private_network" {
  provider = google-beta

  name = "private-network"
}

resource "google_compute_global_address" "dev_sql_instance_private_ip_address" {
  provider = google-beta

  name          = var.dev_sql_instance_ip_adddress
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.private_network.id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  provider = google-beta

  network                 = google_compute_network.private_network.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.dev_sql_instance_private_ip_address.name]
}

resource "google_sql_database_instance" "dev_instance" {
  provider = google-beta

  name             = var.dev_sql_instance_name
  region           = var.region
  database_version = "POSTGRES_18"

  depends_on = [google_service_networking_connection.private_vpc_connection]

  settings {
    tier = "db-f1-micro"
    ip_configuration {
      ipv4_enabled                                  = false
      private_network                               = google_compute_network.private_network.self_link
      enable_private_path_for_google_cloud_services = true
    }
  }
}

provider "google-beta" {
  region = var.region
  zone   = var.zone_a
}