terraform {
  required_version = ">= 0.14"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 3.53"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.1"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.3"
    }
  }

  # STORE TFSTATE AT GOOGLE
  backend "gcs" {
    bucket = "slack-id-tf-state"
    prefix = "terraform/state"
  }
}

provider "google" {
  credentials = file(var.credentials_file)
  project     = var.project_id
  region      = var.region
}

resource "google_container_cluster" "primary" {
  name               = "cloudclimbers-slack-bot-cluster"
  location           = var.region
  initial_node_count = 3

  node_config {
    machine_type = "e2-medium"
    disk_size_gb = 50   # Reduce the disk size if appropriate
    disk_type    = "pd-standard"  # Change to standard persistent disks
  }
}

resource "google_container_node_pool" "primary_nodes" {
  name       = "node-pool"
  location   = google_container_cluster.primary.location
  cluster    = google_container_cluster.primary.name
  node_count = 1

  node_config {
    machine_type = "e2-medium"
    disk_size_gb = 50   # Reduce the disk size if appropriate
    disk_type    = "pd-standard"  # Change to standard persistent disks
  }
}

# Add any other required resources here

output "kubernetes_cluster_name" {
  value = google_container_cluster.primary.name
}

output "kubernetes_cluster_endpoint" {
  value = google_container_cluster.primary.endpoint
}

output "kubernetes_cluster_ca_certificate" {
  value = google_container_cluster.primary.master_auth.0.cluster_ca_certificate
}
