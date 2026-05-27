terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "7.33.0"
    }
  }
}

resource "google_cloud_run_v2_service" "default" {
  name     = var.service_name
  location = var.region
  project  = var.project_id

  scaling {
    max_instance_count = var.max_instance_count
    min_instance_count = var.min_instance_count
  }

  template {
    containers {
      image = var.image
      resources {
        limits = {
          cpu    = var.cpu
          memory = var.memory
        }
        startup_cpu_boost = var.startup_cpu_boost
        cpu_idle          = var.cpu_idle
      }
      dynamic "ports" {
        for_each = var.ports
        content {
          name           = ports.value
          container_port = tonumber(ports.key)
        }
      }
      dynamic "env" {
        for_each = var.environment_variables
        content {
          name  = env.key
          value = env.value
        }
      }
      dynamic "env" {
        for_each = var.secrets
        content {
          name = env.key
          value_source {
            secret_key_ref {
              secret  = env.value
              version = "latest"
            }
          }
        }
      }
      dynamic "startup_probe" {
        for_each = var.startup_probes != null ? [var.startup_probes] : []
        content {
          initial_delay_seconds = startup_probe.value.initial_delay_seconds
          failure_threshold     = startup_probe.value.failure_threshold
          period_seconds        = startup_probe.value.period_seconds
          timeout_seconds       = startup_probe.value.timeout_seconds
          dynamic "http_get" {
            for_each = startup_probe.value.http_get != null ? [startup_probe.value.http_get] : []
            content {
              path = http_get.value.path
              port = http_get.value.port
            }
          }
          dynamic "tcp_socket" {
            for_each = startup_probe.value.tcp_socket != null ? [startup_probe.value.tcp_socket] : []
            content {
              port = tcp_socket.value.port
            }
          }
        }
      }
    }

    dynamic "vpc_access" {
      for_each = var.network != null && var.subnet != null ? [1] : []
      content {
        network_interfaces {
          network    = var.network
          subnetwork = var.subnet
        }
      }
    }

    service_account = google_service_account.sa.email
  }
}

resource "google_service_account" "sa" {
  account_id = "${var.service_name}-sa"
  project    = var.project_id
}

resource "google_secret_manager_secret_iam_member" "secret_access" {
  project   = var.project_id
  for_each  = var.secrets
  secret_id = each.value
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.sa.email}"
  depends_on = [
    google_service_account.sa
  ]
}

resource "google_cloud_run_service_iam_member" "public_access" {
  count    = var.allow_unauthenticated ? 1 : 0
  project  = var.project_id
  location = var.region
  service  = google_cloud_run_v2_service.default.name
  role     = "roles/run.invoker"
  member   = "allUsers"
  depends_on = [
    google_cloud_run_v2_service.default
  ]
}
