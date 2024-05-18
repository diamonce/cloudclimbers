variable "project_id" {
  description = "The project ID to deploy to."
  type        = string
}

variable "region" {
  description = "The region to deploy to."
  type        = string
  default     = "us-central1"
}

variable "slack_token" {
  description = "Slack bot token"
  type        = string
  sensitive   = true
}

variable "cluster_name" {
  description = "The name of the Kubernetes cluster."
  type        = string
  default     = "cloudclimbers-slack-bot-cluster"
}
