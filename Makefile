.PHONY: build-agent build-server build-all clean test run-agent run-server install

# Build configurations
AGENT_NAME=nerve-agent
SERVER_NAME=nerve-center
BUILD_DIR=build
VERSION ?= 1.0.0
GO_BUILD_FLAGS=-ldflags "-X main.Version=$(VERSION)"

# Build agent binary
build-agent:
	@echo "Building nerve-agent..."
	@mkdir -p $(BUILD_DIR)
	cd agent && go build $(GO_BUILD_FLAGS) -o ../$(BUILD_DIR)/$(AGENT_NAME) .

# Build server binary
build-server:
	@echo "Building nerve-center..."
	@mkdir -p $(BUILD_DIR)
	cd server && go build $(GO_BUILD_FLAGS) -o ../$(BUILD_DIR)/$(SERVER_NAME) .

# Build both binaries
build-all: build-agent build-server
	@echo "Build complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@find . -name "*.test" -delete

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Run agent (for testing)
run-agent:
	@go run agent/main.go --server=http://localhost:8080 --token=test-token --debug

# Run server (for testing)
run-server:
	@go run server/main.go --addr=:8080 --debug

# Install dependencies
install:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Cross-compile for Linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)/linux
	cd agent && GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o ../$(BUILD_DIR)/linux/$(AGENT_NAME) .
	cd server && GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o ../$(BUILD_DIR)/linux/$(SERVER_NAME) .

# Cross-compile for multiple platforms
build-cross: build-linux
	@echo "Cross-compilation complete!"

# Run development environment
dev:
	@echo "Starting development environment..."
	@make run-server &
	@sleep 2
	@make run-agent

# Create release
release: build-all
	@echo "Creating release package..."
	@mkdir -p $(BUILD_DIR)/release
	@cp $(BUILD_DIR)/$(AGENT_NAME) $(BUILD_DIR)/release/
	@cp $(BUILD_DIR)/$(SERVER_NAME) $(BUILD_DIR)/release/
	@cp deploy/install.sh $(BUILD_DIR)/release/
	@cp deploy/nerve-agent.service $(BUILD_DIR)/release/
	@cp -r config $(BUILD_DIR)/release/
	@tar -czf $(BUILD_DIR)/nerve-$(VERSION).tar.gz -C $(BUILD_DIR)/release .
	@echo "Release created: $(BUILD_DIR)/nerve-$(VERSION).tar.gz"

# Help
help:
	@echo "Available targets:"
	@echo "  build-agent     - Build nerve-agent binary"
	@echo "  build-server    - Build nerve-center binary"
	@echo "  build-all       - Build both binaries"
	@echo "  clean           - Clean build artifacts"
	@echo "  test            - Run tests"
	@echo "  run-agent       - Run agent for testing"
	@echo "  run-server      - Run server for testing"
	@echo "  install         - Install Go dependencies"
	@echo "  build-linux     - Build for Linux (amd64)"
	@echo "  build-cross     - Cross-compile for all platforms"
	@echo "  dev             - Run development environment"
	@echo "  release         - Create release package"
	@echo "  help            - Show this help message"

