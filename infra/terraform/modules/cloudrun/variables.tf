variable "service_name" {
  description = "The name of the cloud run service"
  type        = string
}

variable "project_id" {
  description = "The project ID where the resources belong"
  type        = string
}

variable "region" {
  description = "The region where the cloud run resources will be deployed"
  type        = string
}

variable "max_instance_count" {
  description = "The maximum number of instances to allocate for the cloud run service"
  type        = number
}

variable "min_instance_count" {
  description = "The minimum number of instance to allocate for the cloud run service"
  type        = number
}

variable "image" {
  description = "The container image to deploy"
  type        = string
}

variable "cpu" {
  description = "The ammout of cpu to allocate for the service"
  type        = string
  default     = "0.5"
}

variable "memory" {
  description = "The amount of memory to allocate for the service"
  type        = string
  default     = "512Mi"
}

variable "startup_cpu_boost" {
  description = "Enable CPU boosting during startup"
  type        = bool
}

variable "cpu_idle" {
  description = "Determine whether CPU is allocated for service while being idle"
  type        = bool
  default     = false
}

variable "ports" {
  description = "A map of port names to port numbers for each service"
  type        = map(string)
  default = {
    "8080" = "http"
  }
}

variable "environment_variables" {
  description = "A map of environment variable names to values for each service"
  type        = map(string)
  default     = {}
}

variable "secrets" {
  description = "A map of secret names to secret values for each service in Secret Manager"
  type        = map(string)
  default     = {}
}

variable "startup_probes" {
  description = "The startup probes configuration for services"
  type = object({
    initial_delay_seconds = optional(number)
    failure_threshold     = optional(number)
    period_seconds        = optional(number)
    timeout_seconds       = optional(number)
    http_get = optional(object({
      path = string
      port = string
    }))
    tcp_socket = optional(object({
      port = number
    }))
  })
  default = null
}

variable "ingress" {
  description = "The ingress settings for the service. Valid settings are INGRESS_TRAFFIC_ALL, INGRESS_TRAFFIC_INTERNAL_ONLY and INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"
  type        = string
  default     = null
}

variable "egress" {
  description = "The egress settings for the service. Valid values are ALL_TRAFFIC and PRIVATE_RANGES_ONLY"
  type        = string
  default     = null
}

variable "network" {
  description = "The network to which the service will be connected"
  type        = string
  default     = null
}

variable "subnet" {
  description = "The subnet to which the service will be connected"
  type        = string
  default     = null
}

variable "allow_unauthenticated" {
  description = "whether to allow unathenticated access or not (defaults to false)"
  type        = bool
  default     = false
}