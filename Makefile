# Makefile for deeplx-cli

# Variables
BINARY_NAME=deeplx-cli
VERSION?=dev
LDFLAGS=-ldflags="-s -w -X 'main.version=$(VERSION)'"
BUILD_DIR=build

# Default target
.PHONY: all
all: build

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)

# Build for current platform
.PHONY: build
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) main.go

# Run tests
.PHONY: test
test:
	go test -v ./...

# Run tests with race detector
.PHONY: test-race
test-race:
	go test -race -v ./...

# Run code analysis
.PHONY: vet
vet:
	go vet ./...

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Download dependencies
.PHONY: deps
deps:
	go mod download
	go mod verify

# Cross-compile for all platforms
.PHONY: build-all
build-all: clean
	mkdir -p $(BUILD_DIR)
	
	# Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 main.go
	
	# Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe main.go
	
	# macOS
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	
	# FreeBSD
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-freebsd-amd64 main.go

# Generate checksums
.PHONY: checksums
checksums: build-all
	cd $(BUILD_DIR) && sha256sum * > checksums.sha256

# Install binary to local system
.PHONY: install
install: build
	cp $(BINARY_NAME) /usr/local/bin/

# Uninstall binary from local system
.PHONY: uninstall
uninstall:
	rm -f /usr/local/bin/$(BINARY_NAME)

# Show version
.PHONY: version
version:
	@echo "Version: $(VERSION)"

# Display help information
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build      - Build binary for current platform"
	@echo "  build-all  - Cross-compile for all platforms"
	@echo "  test       - Run tests"
	@echo "  test-race  - Run tests with race detector"
	@echo "  vet        - Run go vet"
	@echo "  fmt        - Format code"
	@echo "  deps       - Download and verify dependencies"
	@echo "  checksums  - Generate checksums for all binaries"
	@echo "  install    - Install binary to /usr/local/bin"
	@echo "  uninstall  - Remove binary from /usr/local/bin"
	@echo "  clean      - Remove build artifacts"
	@echo "  version    - Show version"
	@echo "  help       - Show this help"
