terraform {
  required_version = ">= 1.9"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 8.4"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  # Same bucket as sre-bible's Terraform, separate prefix (separate state).
  backend "gcs" {
    bucket = "gen-lang-client-0479899208-tfstate"
    prefix = "password-exchange"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}
