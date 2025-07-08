# Zoink - Fast directory navigation

.PHONY: all build test test-coverage clean run build-all lint help

# Build variables
BINARY_NAME=zoink
MODULE=$(shell go list -m)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
LDFLAGS=-ldflags "-X '$(MODULE)/cmd.version=$(VERSION)' -X '$(MODULE)/cmd.buildTime=$(BUILD_TIME)'"

# Default target
all: build

# Build the binary
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

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

# Run the application (for development)
run: build
	./bin/$(BINARY_NAME)

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 .

lint:
	golangci-lint run

help:
	@echo "Available targets:"
	@echo "  all            Build the binary (default)"
	@echo "  build          Build the binary"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage"
	@echo "  clean          Clean build artifacts"
	@echo "  run            Build and run the application"
	@echo "  build-all      Build for multiple platforms"
	@echo "  lint           Run linter"
	@echo ""
	@echo "Development workflow:"
	@echo "  1. make build        # Build the binary"
	@echo "  2. export PATH=\"\$$(pwd)/bin:\$$PATH\"  # Use local binary"
	@echo "  3. zoink setup       # Setup shell integration"
	@echo "  4. source ~/.zshrc   # Activate 'x' alias"
