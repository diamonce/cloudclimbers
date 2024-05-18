# Define variables
BINARY_NAME=cloudclimbers-slack-bot
MAIN_PACKAGE=./cmd
CREATE_PLUGIN_DIR=./charts/plugins/create
GET_PLUGIN_DIR=./charts/plugins/get
DELETE_PLUGIN_DIR=./charts/plugins/delete

# Docker image repositories and tags
MAIN_IMAGE_REPO=myrepo/cloudclimbers-slack-bot
CREATE_IMAGE_REPO=myrepo/create-plugin
GET_IMAGE_REPO=myrepo/get-plugin
DELETE_IMAGE_REPO=myrepo/delete-plugin
IMAGE_TAG=latest

# Ensure Go module is initialized
go-init:
	@echo "==> Initializing Go module..."
	@if [ ! -f go.mod ]; then go mod init github.com/dchernenko/cloudclimbers; fi
	@go mod tidy
	@echo "==> Go module initialized"

# Build main binary
build: go-init
	@echo "==> Building the binary..."
	@go build -o $(BINARY_NAME) $(MAIN_PACKAGE)/main.go
	@echo "==> Build completed: $(BINARY_NAME)"

# Run main binary
run: build
	@echo "==> Running the binary..."
	@./$(BINARY_NAME)

# Run tests
test:
	@echo "==> Running tests..."
	@go test ./... -v
	@echo "==> Tests completed"

# Clean build artifacts
clean:
	@echo "==> Cleaning the binary and temporary files..."
	@rm -f $(BINARY_NAME)
	@echo "==> Clean completed"

# Build Docker images
docker-build:
	@echo "==> Building Docker images..."
	@docker build -t $(MAIN_IMAGE_REPO):$(IMAGE_TAG) .
	@docker build -t $(CREATE_IMAGE_REPO):$(IMAGE_TAG) $(CREATE_PLUGIN_DIR)
	@docker build -t $(GET_IMAGE_REPO):$(IMAGE_TAG) $(GET_PLUGIN_DIR)
	@docker build -t $(DELETE_IMAGE_REPO):$(IMAGE_TAG) $(DELETE_PLUGIN_DIR)
	@echo "==> Docker build completed"

# Run Docker container for main application
docker-run: docker-build
	@echo "==> Running Docker container..."
	@docker run --rm -p 8080:8080 $(MAIN_IMAGE_REPO):$(IMAGE_TAG)

.PHONY: go-init build run test clean docker-build docker-run
