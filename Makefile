MAKEFLAGS += -x

# Define variables
BINARY_NAME=cloudclimbers-slack-bot-runner
MAIN_PACKAGE=./cloudclimbers-slack-bot
CREATE_PLUGIN_DIR=./cloudclimbers-slack-bot/plugins/create
GET_PLUGIN_DIR=./cloudclimbers-slack-bot/plugins/get
DELETE_PLUGIN_DIR=./cloudclimbers-slack-bot/plugins/delete
HELM_CHART_DIR=./helm

# Docker image repositories and tags
PROJECT_ID=slack-id
LOCATION=europe-central2
GCR_REPO=gcr.io/$(PROJECT_ID)
MAIN_IMAGE_REPO=$(GCR_REPO)/cloudclimbers-slack-bot
CREATE_IMAGE_REPO=$(GCR_REPO)/cloudclimbers-slack-bot-create-plugin
GET_IMAGE_REPO=$(GCR_REPO)/cloudclimbers-slack-bot-get-plugin
DELETE_IMAGE_REPO=$(GCR_REPO)/cloudclimbers-slack-bot-delete-plugin
IMAGE_TAG=latest

# Helm release name and namespace
HELM_RELEASE_NAME=cloudclimbers-slack-bot
HELM_NAMESPACE=cloudclimbers

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

# Build Docker images
docker-build: build
	echo "==> Building Docker images..."
	docker buildx build --platform $(OS)/$(ARCH) -t $(MAIN_IMAGE_REPO):$(IMAGE_TAG) .
	docker buildx build --platform $(OS)/$(ARCH) -t $(CREATE_IMAGE_REPO):$(IMAGE_TAG) $(CREATE_PLUGIN_DIR)
	docker buildx build --platform $(OS)/$(ARCH) -t $(GET_IMAGE_REPO):$(IMAGE_TAG) $(GET_PLUGIN_DIR)
	docker buildx build --platform $(OS)/$(ARCH) -t $(DELETE_IMAGE_REPO):$(IMAGE_TAG) $(DELETE_PLUGIN_DIR)
	echo "==> Docker build completed"

#Build Docker images for Github repo
docker-build-ghcr: build
	echo "==> Building Docker images..."
	docker buildx build --platform $(OS)/$(ARCH) -t ghcr.io/diamonce/cloudclimbers:$(IMAGE_TAG) .
	docker buildx build --platform $(OS)/$(ARCH) -t ghcr.io/diamonce/cloudclimbers-create-plugin:$(IMAGE_TAG) $(CREATE_PLUGIN_DIR)
	docker buildx build --platform $(OS)/$(ARCH) -t ghcr.io/diamonce/cloudclimbers-get-plugin:$(IMAGE_TAG) $(GET_PLUGIN_DIR)
	docker buildx build --platform $(OS)/$(ARCH) -t ghcr.io/diamonce/cloudclimbers-delete-plugin:$(IMAGE_TAG) $(DELETE_PLUGIN_DIR)
	echo "==> Docker build completed"

# Push Docker images to GCR
docker-push: gcr-init docker-build
	echo "==> Pushing Docker images to GCR..."
	docker push $(MAIN_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(CREATE_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(GET_IMAGE_REPO):$(IMAGE_TAG)
	docker push $(DELETE_IMAGE_REPO):$(IMAGE_TAG)
	echo "==> Docker images pushed to GCR"

docker-push-ghcr: docker-build
	echo "==> Pushing Docker images to GHCR..."
    docker push ghcr.io/diamonce/cloudclimbers:$(IMAGE_TAG)
	docker push ghcr.io/diamonce/cloudclimbers-create-plugin:$(IMAGE_TAG)
	docker push ghcr.io/diamonce/cloudclimbers-get-plugin:$(IMAGE_TAG)
	docker push ghcr.io/diamonce/cloudclimbers-delete-plugin:$(IMAGE_TAG)
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
#	helm repo update

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

.PHONY: go-init deps build run test clean docker-build docker-run docker-compose-build docker-compose-up docker-compose-down gcr-init gcr-push helm-deps helm-install helm-upgrade helm-uninstall helm-repo-add
