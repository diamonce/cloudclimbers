output "cluster_name" {
  description = "The name of the Kubernetes cluster"
  value       = google_container_cluster.primary.name
}

output "kubernetes_cluster_endpoint" {
  description = "The endpoint of the Kubernetes cluster"
  value       = google_container_cluster.primary.endpoint
}

output "kubernetes_cluster_name" {
  description = "The name of the Kubernetes cluster"
  value       = google_container_cluster.primary.name
}

output "kubernetes_cluster_ca_certificate" {
  description = "The CA certificate of the Kubernetes cluster"
  value       = base64decode(google_container_cluster.primary.master_auth[0].cluster_ca_certificate)
}

output "argocd_server_url" {
  description = "The URL of the ArgoCD server"
  value       = kubernetes_service.argocd.status[0].load_balancer[0].ingress[0].hostname
}
