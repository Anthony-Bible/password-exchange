variable "project_id" {
  description = "GCP project ID"
  type        = string
  default     = "gen-lang-client-0479899208"
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "instance_name" {
  description = "Existing Cloud SQL instance to create the databases on. Owned by sre-bible's Terraform; read-only here."
  type        = string
  default     = "sre-bible"
}

variable "k8s_namespace" {
  description = "Kubernetes namespace of the password-exchange workloads"
  type        = string
  default     = "default"
}

variable "k8s_service_account" {
  description = "Kubernetes service account the workloads run as"
  type        = string
  default     = "password-exchange"
}
