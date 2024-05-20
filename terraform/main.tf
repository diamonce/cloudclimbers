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
}

provider "google" {
  credentials = file("<path-to-your-service-account-key>.json")
  project     = var.project_id
  region      = var.region
}

resource "google_container_cluster" "primary" {
  name               = "cloudclimbers-slack-bot-cluster"
  location           = var.region
  initial_node_count = 3

  node_config {
    machine_type = "e2-medium"
  }
}

resource "google_container_node_pool" "primary_preemptible_nodes" {
  name       = "preemptible-node-pool"
  location   = google_container_cluster.primary.location
  cluster    = google_container_cluster.primary.name
  node_count = 1

  node_config {
    preemptible  = true
    machine_type = "e2-medium"
  }
}

resource "google_container_cluster" "cluster" {
  name               = var.cluster_name
  location           = var.region
  initial_node_count = 1
}

resource "kubernetes_namespace" "argocd" {
  metadata {
    name = "argocd"
  }
}

resource "helm_release" "argocd" {
  name       = "argocd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  namespace  = kubernetes_namespace.argocd.metadata[0].name
  values = [
    <<EOF
server:
  service:
    type: LoadBalancer
    port: 80
    targetPort: 8080
  config:
    url: https://argocd-server.argocd.svc.cluster.local
EOF
  ]
}

resource "kubernetes_namespace" "my_slack_bot" {
  metadata {
    name = "cloudclimbers-slack-bot"
  }
}

resource "kubernetes_secret" "slack_bot_secret" {
  metadata {
    name      = "cloudclimbers-slack-bot-secrets"
    namespace = kubernetes_namespace.my_slack_bot.metadata[0].name
  }

  data = {
    slackToken = base64encode(var.slack_token)
  }
}

resource "kubernetes_config_map" "slack_bot_config" {
  metadata {
    name      = "cloudclimbers-slack-bot-config"
    namespace = kubernetes_namespace.my_slack_bot.metadata[0].name
  }

  data = {
    "plugins.yaml" = file("${path.module}/plugins.yaml")
  }
}

resource "helm_release" "my_slack_bot" {
  name       = "cloudclimbers-slack-bot"
  repository = "file://../cloudclimbers-slack-bot"
  chart      = "cloudclimbers-slack-bot"
  namespace  = kubernetes_namespace.my_slack_bot.metadata[0].name
  values = [
    <<EOF
slackToken: "${kubernetes_secret.slack_bot_secret.data["slackToken"]}"
mongoURI: "mongodb://admin:password@mongodb:27017/mydatabase"
pluginsConfig: |
  plugins:
    create:
      url: "http://create:8081/create"
    get:
      url: "http://get:8082/get"
    delete:
      url: "http://delete:8083/delete"
  main_buttons:
    - text: "List Enabled Plugins"
      action_id: "list_enabled_plugins"
EOF
  ]
}

resource "kubernetes_deployment" "argocd" {
  metadata {
    name      = "argocd"
    namespace = kubernetes_namespace.argocd.metadata[0].name
  }
  spec {
    replicas = 1
    selector {
      match_labels = {
        app = "argocd"
      }
    }
    template {
      metadata {
        labels = {
          app = "argocd"
        }
      }
      spec {
        container {
          name  = "argocd"
          image = "argoproj/argocd:latest"
          ports {
            container_port = 8080
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "argocd" {
  metadata {
    name      = "argocd-service"
    namespace = kubernetes_namespace.argocd.metadata[0].name
  }
  spec {
    selector = {
      app = "argocd"
    }
    type = "LoadBalancer"
    ports {
      port        = 80
      target_port = 8080
    }
  }
}
