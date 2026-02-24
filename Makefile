.PHONY: build install clean test test-race run snapshot release

# Build variables
BINARY_NAME=freshtime
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_DIR=./build
GO_FILES=$(shell find . -type f -name '*.go')

# ldflags for version injection
LDFLAGS=-s -w \
	-X github.com/hev/freshtime/internal/config.Version=$(VERSION) \
	-X github.com/hev/freshtime/internal/config.Commit=$(COMMIT) \
	-X github.com/hev/freshtime/internal/config.Date=$(DATE)

# Build the binary
build:
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/freshtime

# Install to GOPATH/bin
install:
	go install -ldflags="$(LDFLAGS)" ./cmd/freshtime

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -rf dist/
	go clean

# Run tests
test:
	go test -v ./...

# Run tests with race detector
test-race:
	go test -v -race ./...

# Run freshtime with default settings
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/freshtime
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/freshtime
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/freshtime
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/freshtime
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/freshtime

# Create a snapshot release (for testing)
snapshot:
	goreleaser release --snapshot --clean

# Create a release (requires GITHUB_TOKEN)
release:
	goreleaser release --clean

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the freshtime binary"
	@echo "  install    - Install freshtime to GOPATH/bin"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  test-race  - Run tests with race detector"
	@echo "  run        - Build and run freshtime"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  snapshot   - Create a snapshot release (testing)"
	@echo "  release    - Create a release with goreleaser"
