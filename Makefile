MAKEFLAGS += -x

# Define variables
BINARY_NAME=cloudclimbers-slack-bot-runner
MAIN_PACKAGE=./cloudclimbers-slack-bot
CREATE_PLUGIN_DIR=./cloudclimbers-slack-bot/charts/plugins/create
GET_PLUGIN_DIR=./cloudclimbers-slack-bot/charts/plugins/get
DELETE_PLUGIN_DIR=./cloudclimbers-slack-bot/charts/plugins/delete

# Docker image repositories and tags
MAIN_IMAGE_REPO=dchernenko/cloudclimbers-slack-bot
CREATE_IMAGE_REPO=dchernenko/cloudclimbers-slack-bot-create-plugin
GET_IMAGE_REPO=dchernenko/cloudclimbers-slack-bot-get-plugin
DELETE_IMAGE_REPO=dchernenko/cloudclimbers-slack-bot-delete-plugin
IMAGE_TAG=latest

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
	cd $(MAIN_PACKAGE) && go build -o "$(BINARY_NAME)" ./cmd/main.go
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

# Build Docker images
docker-build: build
	echo "==> Building Docker images..."
	docker build -t $(MAIN_IMAGE_REPO):$(IMAGE_TAG) .
	docker build -t $(CREATE_IMAGE_REPO):$(IMAGE_TAG) $(CREATE_PLUGIN_DIR)
	docker build -t $(GET_IMAGE_REPO):$(IMAGE_TAG) $(GET_PLUGIN_DIR)
	docker build -t $(DELETE_IMAGE_REPO):$(IMAGE_TAG) $(DELETE_PLUGIN_DIR)
	echo "==> Docker build completed"

# Run Docker container for main application
docker-run: docker-build
	echo "==> Running Docker container..."
	docker run --rm -p 8080:8080 $(MAIN_IMAGE_REPO):$(IMAGE_TAG)

.PHONY: go-init deps build run test clean docker-build docker-run
