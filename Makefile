# Zoink - Fast directory navigation

.PHONY: build install test clean setup deps

# Build variables
BINARY_NAME=zoink
MODULE=$(shell go list -m)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
LDFLAGS=-ldflags "-X '$(MODULE)/cmd.version=$(VERSION)' -X '$(MODULE)/cmd.buildTime=$(BUILD_TIME)'"

# Default target
all: build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build the binary
build: deps
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

# Install to GOPATH/bin or GOBIN
install: deps
	go install $(LDFLAGS) .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Development setup
setup: deps
	@echo "Setting up development environment..."
	@mkdir -p bin
	@echo "Development environment ready!"

# Run the application (for development)
run: build
	./bin/$(BINARY_NAME)

# Build for multiple platforms
build-all: deps
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 .

# Generate shell completions
completions: build
	@mkdir -p completions
	./bin/$(BINARY_NAME) completion bash > completions/$(BINARY_NAME).bash
	./bin/$(BINARY_NAME) completion zsh > completions/$(BINARY_NAME).zsh
	./bin/$(BINARY_NAME) completion fish > completions/$(BINARY_NAME).fish

# Development helpers
fmt:
	go fmt ./...

lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build          Build the binary"
	@echo "  install        Install to GOPATH/bin"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage"
	@echo "  clean          Clean build artifacts"
	@echo "  setup          Setup development environment"
	@echo "  dev-setup      Setup with local binary and shell integration"
	@echo "  run            Build and run the application"
	@echo "  build-all      Build for multiple platforms"
	@echo "  completions    Generate shell completions"
	@echo "  fmt            Format code"
	@echo "  lint           Run linter"
	@echo "  deps           Download and tidy dependencies"
	@echo ""
	@echo "Development workflow:"
	@echo "  1. make dev-setup    # Sets up shell integration with local binary"
	@echo "  2. source ~/.zshrc   # Activate 'x' alias (uses system binary initially)"
	@echo "  3. export PATH=\"\$$(pwd)/bin:\$$PATH\"  # Switch to local binary"

# Development setup with shell integration
dev-setup: build
	@echo "Setting up development environment with local binary..."
	@echo "Adding $(PWD)/bin to PATH and running setup..."
	PATH="$(PWD)/bin:$$PATH" ./bin/$(BINARY_NAME) setup --quiet
	@echo ""
	@echo "Development setup complete!"
	@echo "To activate the 'x' alias, run:"
	@echo "  source ~/.zshrc (or your shell's config file)"
	@echo ""
	@echo "Then to use the local zoink binary instead of system binary, run:"
	@echo "  export PATH=\"$(PWD)/bin:\$$PATH\""
