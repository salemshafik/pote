variable "project_id" {
  description = "The ID of the project where the resources will be deployed"
  type        = string
}

variable "region" {
  description = "The region where the resources will be deployed"
  type        = string
}

variable "env_name" {
  description = "The name of the environment"
  type        = string
}

variable "dev_database_name" {
  description = "The name of the SQL database"
  type        = string
}

variable "dev_sql_instance_ip_adddress" {
  description = "The reserved private IP address of the SQL instance"
  type        = string
}

variable "dev_sql_instance_name" {
  description = "The name of the SQL instance to be deployed"
  type        = string
}

variable "zone_a" {
  description = "The zone where the resources will be deployed"
  type        = string
}
