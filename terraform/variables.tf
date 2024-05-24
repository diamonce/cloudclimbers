variable "project_id" {
  description = "The ID of the GCP project."
  type        = string
  default     = "slack-id"
}

variable "region" {
  description = "The region where the GKE cluster will be deployed."
  type        = string
  default     = "europe-central2-c"
}

variable "credentials_file" {
  description = "Path to the GCP service account key file."
  type        = string
  default     = "/Users/dmytrochernenko/.config/gcloud/slack-id-921f7b160345-tf.json"
}

variable "cluster_name" {
  description = "The name of the GKE cluster."
  type        = string
  default     = "cloudclimbers-slack-bot-cluster"
}
