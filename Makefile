# HarborArk Makefile
# Cross-platform build support targeting Linux

# Project variables
BINARY_NAME=pnas
BINARY_LINUX=$(BINARY_NAME)
BUILD_DIR=build
VERSION ?= $(shell git describe --tags --always --dirty="-dev" 2>/dev/null || echo "v0.0.0")
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Go build flags
LDFLAGS=-ldflags "-X 'pnas/internal/config.AppConfig.Server.Version=$(VERSION)' -X 'pnas/internal/config.AppConfig.Server.Commit=$(COMMIT)' -X 'pnas/internal/config.AppConfig.Server.BuildTime=$(DATE)'"

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Default target - build for current platform
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build: $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for Linux (cross-platform)
.PHONY: build-linux
build-linux: $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_LINUX)-linux-amd64 .
	@echo "Linux AMD64 binary built: $(BUILD_DIR)/$(BINARY_LINUX)-linux-amd64"

# Alternative name for Linux build
.PHONY: linux
linux: build-linux

# Build for multiple Linux architectures
.PHONY: build-linux-all
build-linux-all: build-linux-amd64 build-linux-arm64 build-linux-arm build-linux-386

.PHONY: build-linux-amd64
build-linux-amd64: $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_LINUX)-linux-amd64 .
	@echo "Linux AMD64 binary built: $(BUILD_DIR)/$(BINARY_LINUX)-linux-amd64"

.PHONY: build-linux-arm64
build-linux-arm64: $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_LINUX)-linux-arm64 .
	@echo "Linux ARM64 binary built: $(BUILD_DIR)/$(BINARY_LINUX)-linux-arm64"

.PHONY: build-linux-arm
build-linux-arm: $(BUILD_DIR)
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_LINUX)-linux-arm .
	@echo "Linux ARM binary built: $(BUILD_DIR)/$(BINARY_LINUX)-linux-arm"

.PHONY: build-linux-386
build-linux-386: $(BUILD_DIR)
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_LINUX)-linux-386 .
	@echo "Linux 386 binary built: $(BUILD_DIR)/$(BINARY_LINUX)-linux-386"

# Build for all supported platforms (Linux focused)
.PHONY: build-all
build-all: build-linux

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	@echo "Build directory cleaned"

# Dist clean (clean + remove all binaries)
.PHONY: distclean
distclean: clean
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*

# Install binary to system (for current platform)
.PHONY: install
install: build
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Run the application (for development)
.PHONY: run
run:
	go run . serve

# Run with debug mode (for development)
.PHONY: run-dev
run-dev:
	go run . serve --debug

# Test the application
.PHONY: test
test:
	go test ./...

# Test with coverage
.PHONY: test-cover
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Build with race detection
.PHONY: build-race
build-race: $(BUILD_DIR)
	go build -race $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-race .

# Generate documentation (if needed)
.PHONY: docs
docs:
	go run github.com/swaggo/swag/cmd/swag@latest init --dir=./internal --generalInfo=server.go

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Vet code
.PHONY: vet
vet:
	go vet ./...

# Tidy go modules
.PHONY: tidy
tidy:
	go mod tidy

# Check for outdated dependencies
.PHONY: outdated
outdated:
	go list -u -m all

# Build and package for distribution
.PHONY: package
package: build-linux
	tar -czf $(BUILD_DIR)/$(BINARY_LINUX)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_LINUX)-linux-amd64
	@echo "Linux package created: $(BUILD_DIR)/$(BINARY_LINUX)-$(VERSION)-linux-amd64.tar.gz"

# Help target
.PHONY: help
help:
	@echo "HarborArk Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build            - Build for current platform"
	@echo "  make linux            - Build for Linux (alias for build-linux)"
	@echo "  make build-linux      - Build for Linux AMD64"
	@echo "  make build-linux-all  - Build for all Linux architectures"
	@echo "  make build-all        - Build for all supported platforms"
	@echo "  make package          - Build Linux binary and create distribution package"
	@echo "  make run              - Run the server (for development)"
	@echo "  make run-dev          - Run with debug mode"
	@echo "  make test             - Run tests"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make distclean        - Remove build artifacts and binaries"
	@echo "  make fmt              - Format code"
	@echo "  make vet              - Vet code"
	@echo "  make tidy             - Tidy go modules"
	@echo ""