# hostsctl Makefile

.PHONY: build test clean install lint fmt vet deps build-all

# Variables
BINARY_NAME := hostsctl
MAIN_PACKAGE := ./cmd/hostsctl
BUILD_DIR := build
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)

	# Linux AMD64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)

	# Linux ARM64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)

	# macOS AMD64
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)

	# macOS ARM64
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)

	# Windows AMD64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

# Run tests
test:
	@echo "Running tests..."
	go test -v -coverprofile=coverage.out ./...

# Test with coverage report
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install binary to system
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)

# Uninstall binary from system
uninstall:
	@echo "Removing $(BINARY_NAME) from /usr/local/bin..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Format code with gofmt
fmt:
	@echo "Formatting code..."
	gofmt -w .

# Vet code
vet:
	@echo "Running go vet..."
	go vet ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; exit 1)
	golangci-lint run

# Check all (format, vet, lint, test)
check: fmt vet lint test

# Run the binary
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

# Run with example hosts file
run-test:
	$(BUILD_DIR)/$(BINARY_NAME) --hosts-file ./test_hosts list

# Docker build
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .

# Nix targets
nix-build:
	@echo "Building with Nix..."
	nix build

nix-run:
	@echo "Running with Nix..."
	nix run . -- --help

nix-shell:
	@echo "Entering Nix development shell..."
	nix develop

nix-install:
	@echo "Installing with nix profile..."
	nix profile install .

nix-update:
	@echo "Updating flake lock..."
	nix flake update

# Create a test hosts file for development
test-hosts:
	@echo "Creating test hosts file..."
	@echo "# Test hosts file for development" > test_hosts
	@echo "127.0.0.1	localhost" >> test_hosts
	@echo "192.168.1.100	server.local web.local	# Test server" >> test_hosts
	@echo "# 192.168.1.200	disabled.local	# Disabled entry" >> test_hosts

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install binary to /usr/local/bin"
	@echo "  uninstall     - Remove binary from /usr/local/bin"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run linter (requires golangci-lint)"
	@echo "  check         - Run fmt, vet, lint, and test"
	@echo "  run           - Build and run the binary"
	@echo "  run-test      - Run with test hosts file"
	@echo "  docker-build  - Build Docker image"
	@echo "  nix-build     - Build with Nix"
	@echo "  nix-run       - Run with Nix"
	@echo "  nix-shell     - Enter Nix development shell"
	@echo "  nix-install   - Install with nix profile"
	@echo "  nix-update    - Update Nix flake lock"
	@echo "  test-hosts    - Create a test hosts file"
	@echo "  help          - Show this help"

# Default target
.DEFAULT_GOAL := help