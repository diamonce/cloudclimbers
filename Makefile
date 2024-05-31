MAKEFLAGS += -x

# Define variables
BINARY_NAME=cloudclimbers-slack-bot-runner
MAIN_PACKAGE=./cloudclimbers-slack-bot
ARGO_CREATE_PLUGIN_DIR=./cloudclimbers-slack-bot/plugins/create_argo
FLUX_CREATE_PLUGIN_DIR=./cloudclimbers-slack-bot/plugins/create_flux
GET_PLUGIN_DIR=./cloudclimbers-slack-bot/plugins/get
DELETE_PLUGIN_DIR=./cloudclimbers-slack-bot/plugins/delete
HELM_CHART_DIR=./helm
FLUX_DIR=./flux

# Variables for the Cloudflare AI plugin
CLOUDFLARE_AI_PLUGIN_DIR=./cloudclimbers-slack-bot/plugins/expert_ai_model
CLOUDFLARE_AI_IMAGE_REPO=$(GCR_REPO)/cloudclimbers-slack-bot-expert-ai-plugin
GHCR_CLOUDFLARE_AI_IMAGE_REPO=$(GHCR_REPO)/cloudclimbers-slack-bot-expert-ai-plugin

# Docker image repositories and tags
PROJECT_ID=slack-id
LOCATION=europe-central2
GCR_REPO=gcr.io/$(PROJECT_ID)
GHCR_REPO=ghcr.io/diamonce/cloudclimbers
MAIN_IMAGE_REPO=$(GCR_REPO)/cloudclimbers-slack-bot
ARGO_CREATE_IMAGE_REPO=$(GCR_REPO)/cloudclimbers-slack-bot-create-argo-plugin
FLUX_CREATE_IMAGE_REPO=$(GCR_REPO)/cloudclimbers-slack-bot-create-flux-plugin
GET_IMAGE_REPO=$(GCR_REPO)/cloudclimbers-slack-bot-get-plugin
DELETE_IMAGE_REPO=$(GCR_REPO)/cloudclimbers-slack-bot-delete-plugin
IMAGE_TAG=latest
GHCR_MAIN_IMAGE_REPO=$(GHCR_REPO)/cloudclimbers-slack-bot
GHCR_ARGO_CREATE_IMAGE_REPO=$(GHCR_REPO)/cloudclimbers-slack-bot-create-argo-plugin
GHCR_FLUX_CREATE_IMAGE_REPO=$(GHCR_REPO)/cloudclimbers-slack-bot-create-flux-plugin
GHCR_GET_IMAGE_REPO=$(GHCR_REPO)/cloudclimbers-slack-bot-get-plugin
GHCR_DELETE_IMAGE_REPO=$(GHCR_REPO)/cloudclimbers-slack-bot-delete-plugin

# Helm release name and namespace
HELM_RELEASE_NAME=cloudclimbers-slack-bot
HELM_NAMESPACE=cloudclimbers

# Flux parameters
GITHUB_USER=diamonce
GITHUB_REPO=https://github.com/diamonce/cloudclimbers
GITHUB_TOKEN=
FLUX_CONTEXT=gke_slack-id_europe-central2-c_cloudclimbers-slack-bot-cluster
FLUX_PATH=$(FLUX_DIR)/clusters/$(FLUX_CONTEXT)

# Default architecture
ARCH=amd64
OS=linux

# Ensure Go module is initialized
go-init:
	echo "==> Initializing Go module in $(MAIN_PACKAGE)..."
	cd $(MAIN_PACKAGE) && if [ ! -f go.mod ]; then go mod init cloudclimbers-slack-bot; fi
	echo "==> Go module initialized"

# Download Go dependencies
deps: go-init
	echo "==> Downloading Go dependencies in $(MAIN_PACKAGE)..."
	cd $(MAIN_PACKAGE) && go mod tidy
	echo "==> Dependencies downloaded"

# Build create_flux plugin
build-create-flux:
	echo "==> Building the create_flux plugin..."
	- cp $(MAIN_PACKAGE)/go.mod $(FLUX_CREATE_PLUGIN_DIR)
	- cp $(MAIN_PACKAGE)/go.sum $(FLUX_CREATE_PLUGIN_DIR)
	cd $(FLUX_CREATE_PLUGIN_DIR) && GOTOOLCHAIN="" go mod tidy -e
	cd $(FLUX_CREATE_PLUGIN_DIR) && GOOS=$(OS) GOARCH=$(ARCH) go build -o create_flux ./create_plugin.go
	echo "==> Build completed: create_flux"

# Build main binary
build: deps
	echo "==> Building the binary..."
	cd $(MAIN_PACKAGE) && GOOS=$(OS) GOARCH=$(ARCH) go build -o "$(BINARY_NAME)" ./cmd/main.go
	echo "==> Build completed: $(BINARY_NAME)"

# Run main binary
run: build
	echo "==> Running the binary..."
	./$(BINARY_NAME)

# Run tests
test:
	echo "==> Running tests in $(MAIN_PACKAGE)..."
	cd $(MAIN_PACKAGE) && go test ./... -v
	echo "==> Tests completed"

# Clean build artifacts
clean:
	echo "==> Cleaning the binary and temporary files..."
	rm -f $(BINARY_NAME)
	echo "==> Clean completed"

# Create GCR repository if not exists
gcr-init:
	echo "==> Checking and creating GCR repository if it doesn't exist..."
	gcloud auth configure-docker
	gcloud services enable containerregistry.googleapis.com

# Build Docker images for GCR
docker-build: build build-create-flux
	echo "==> Building Docker images..."
	docker buildx build --platform $(OS)/$(ARCH) -t $(MAIN_IMAGE_REPO):$(IMAGE_TAG) .
	docker buildx build --platform $(OS)/$(ARCH) -t $(ARGO_CREATE_IMAGE_REPO):$(IMAGE_TAG) $(ARGO_CREATE_PLUGIN_DIR)
	docker buildx build --platform $(OS)/$(ARCH) -t $(FLUX_CREATE_IMAGE_REPO):$(IMAGE_TAG) $(FLUX_CREATE_PLUGIN_DIR)
	docker buildx build --platform $(OS)/$(ARCH) -t $(GET_IMAGE_REPO):$(IMAGE_TAG) $(GET_PLUGIN_DIR)
	docker buildx build --platform $(OS)/$(ARCH) -t $(DELETE_IMAGE_REPO):$(IMAGE_TAG) $(DELETE_PLUGIN_DIR)
	docker buildx build --platform $(OS)/$(ARCH) -t $(CLOUDFLARE_AI_IMAGE_REPO):$(IMAGE_TAG) $(CLOUDFLARE_AI_PLUGIN_DIR)
	echo "==> Docker build completed"

# Build Docker images for GHCR
docker-build-ghcr: build build-create-flux
	echo "==> Building Docker images..."
	docker buildx build --build-arg GITHUB_REF="$(GITHUB_REF)" --build-arg GITHUB_SHA="$(GITHUB_SHA)" --platform $(OS)/$(ARCH) -t $(GHCR_MAIN_IMAGE_REPO):$(IMAGE_TAG) .
	docker buildx build --build-arg GITHUB_REF="$(GITHUB_REF)" --build-arg GITHUB_SHA="$(GITHUB_SHA)" --platform $(OS)/$(ARCH) -t $(GHCR_ARGO_CREATE_IMAGE_REPO):$(IMAGE_TAG) $(ARGO_CREATE_PLUGIN_DIR)
	docker buildx build --build-arg GITHUB_REF="$(GITHUB_REF)" --build-arg GITHUB_SHA="$(GITHUB_SHA)" --platform $(OS)/$(ARCH) -t $(GHCR_FLUX_CREATE_IMAGE_REPO):$(IMAGE_TAG) $(FLUX_CREATE_PLUGIN_DIR)
	docker buildx build --build-arg GITHUB_REF="$(GITHUB_REF)" --build-arg GITHUB_SHA="$(GITHUB_SHA)" --platform $(OS)/$(ARCH) -t $(GHCR_GET_IMAGE_REPO):$(IMAGE_TAG) $(GET_PLUGIN_DIR)
	docker buildx build --build-arg GITHUB_REF="$(GITHUB_REF)" --build-arg GITHUB_SHA="$(GITHUB_SHA)" --platform $(OS)/$(ARCH) -t $(GHCR_DELETE_IMAGE_REPO):$(IMAGE_TAG) $(DELETE_PLUGIN_DIR)
	docker buildx build --build-arg GITHUB_REF="$(GITHUB_REF)" --build-arg GITHUB_SHA="$(GITHUB_SHA)" --platform $(OS)/$(ARCH) -t $(GHCR_CLOUDFLARE_AI_IMAGE_REPO):$(IMAGE_TAG) $(CLOUDFLARE_AI_PLUGIN_DIR)
	echo "==> Docker build completed!"

# Push Docker images to GCR
docker-push: gcr-init docker-build
	echo "==> Pushing Docker images to GCR..."
	docker push $(MAIN_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(ARGO_CREATE_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(FLUX_CREATE_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(GET_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(DELETE_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(CLOUDFLARE_AI_IMAGE_REPO):$(IMAGE_TAG)
	echo "==> Docker images pushed to GCR"


# Push Docker images to GHCR
docker-push-ghcr:  docker-build-ghcr
	echo "==> Pushing Docker images to GHCR..."
	docker push $(GHCR_MAIN_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(GHCR_ARGO_CREATE_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(GHCR_FLUX_CREATE_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(GHCR_GET_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(GHCR_DELETE_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(GHCR_CLOUDFLARE_AI_IMAGE_REPO):$(IMAGE_TAG)
	echo "==> Docker images pushed to GHCR"

# Run Docker container for main application
docker-run: docker-build
	echo "==> Running Docker container..."
	docker run --rm -p 8080:8080 $(MAIN_IMAGE_REPO):$(IMAGE_TAG)

# Build and run with Docker Compose
docker-compose-build:
	echo "==> Building Docker images with Docker Compose..."
	docker-compose build
	echo "==> Docker Compose build completed"

docker-compose-up:
	echo "==> Running Docker containers with Docker Compose..."
	docker-compose up --build
	echo "==> Docker Compose up completed"

docker-compose-down:
	echo "==> Stopping Docker containers with Docker Compose..."
	docker-compose down
	echo "==> Docker Compose down completed"

# Helm repository add
helm-repo-add:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add argo https://argoproj.github.io/argo-helm

# Helm dependency build
helm-deps: helm-repo-add
	helm dependency update $(HELM_CHART_DIR)
	helm dependency build $(HELM_CHART_DIR)

# Install Helm chart
helm-install: helm-deps
	- kubectl delete serviceaccount argocd --namespace cloudclimbers
	helm install $(HELM_RELEASE_NAME) $(HELM_CHART_DIR) --namespace $(HELM_NAMESPACE) --create-namespace

# Upgrade Helm chart
helm-upgrade: helm-deps
	helm upgrade $(HELM_RELEASE_NAME) $(HELM_CHART_DIR) --namespace $(HELM_NAMESPACE)

# Uninstall Helm chart
helm-uninstall:
	helm uninstall $(HELM_RELEASE_NAME) --namespace $(HELM_NAMESPACE)

# Flux installation
flux-install:
	brew install fluxcd/tap/flux || true
	flux check --pre
	echo "$(GITHUB_TOKEN)" | flux bootstrap github \
	    --context=$(FLUX_CONTEXT) \
	    --owner=$(GITHUB_USER) \
	    --repository=$(GITHUB_REPO) \
	    --branch=main \
	    --personal \
	    --path=$(FLUX_PATH) \
	    --token-auth
	# Authorisation for k8 native
	kubectl apply -f ./flux/clusterrole.yaml
	kubectl apply -f ./flux/clusterrolebinding.yaml

# Flux uninstallation
flux-uninstall:
	flux uninstall --silent --namespace=flux-system

.PHONY: go-init deps build run test clean docker-build docker-run docker-compose-build docker-compose-up docker-compose-down gcr-init gcr-push helm-deps helm-install helm-upgrade helm-uninstall helm-repo-add flux-install build-create-flux
