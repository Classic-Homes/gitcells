.PHONY: all build build-all test test-short test-coverage clean install install-dev lint fmt check deps docker docker-build docker-push release help

# Build configuration
VERSION ?= 1.0.0
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w -X 'github.com/Classic-Homes/gitcells/internal/constants.Version=$(VERSION)' -X 'github.com/Classic-Homes/gitcells/internal/constants.BuildTime=$(BUILD_TIME)' -X 'github.com/Classic-Homes/gitcells/internal/constants.CommitHash=$(COMMIT_HASH)'
BINARY := gitcells
DOCKER_REGISTRY ?= classichomes
DOCKER_IMAGE := $(DOCKER_REGISTRY)/gitcells

# Go configuration
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Installation paths
INSTALL_PATH := /usr/local/bin
CONFIG_PATH := ~/.config/gitcells

# Default target
all: test build

## Build targets

# Build for current platform
build:
	@echo "🔨 Building GitCells $(VERSION) for $(shell go env GOOS)/$(shell go env GOARCH)..."
	@mkdir -p dist
	CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o dist/$(BINARY) ./cmd/gitcells
	@echo "✅ Binary built: dist/$(BINARY)"

# Build for all supported platforms
build-all:
	@echo "🔨 Building GitCells $(VERSION) for all platforms..."
	@mkdir -p dist
	@echo "Building for Darwin amd64..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-amd64 ./cmd/gitcells
	@echo "Building for Darwin arm64..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64 ./cmd/gitcells
	@echo "Building for Linux amd64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64 ./cmd/gitcells
	@echo "Building for Linux arm64..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64 ./cmd/gitcells
	@echo "Building for Windows amd64..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-amd64.exe ./cmd/gitcells
	@echo "✅ All binaries built in dist/"

# Create user-friendly executables and archives
release: build-all
	@echo "📦 Creating user-friendly executables and archives..."
	@mkdir -p dist/releases
	# Windows
	@cp dist/$(BINARY)-windows-amd64.exe dist/releases/$(BINARY)-windows.exe
	@if command -v zip >/dev/null 2>&1; then \
		cd dist/releases && zip -q $(BINARY)-windows.zip $(BINARY)-windows.exe; \
	else \
		cd dist/releases && tar -czf $(BINARY)-windows.tar.gz $(BINARY)-windows.exe; \
	fi
	# Linux AMD64
	@cp dist/$(BINARY)-linux-amd64 dist/releases/$(BINARY)-linux
	@cd dist/releases && tar -czf $(BINARY)-linux.tar.gz $(BINARY)-linux
	# Linux ARM64
	@cp dist/$(BINARY)-linux-arm64 dist/releases/$(BINARY)-linux-arm64
	@cd dist/releases && tar -czf $(BINARY)-linux-arm64.tar.gz $(BINARY)-linux-arm64
	# macOS Intel
	@cp dist/$(BINARY)-darwin-amd64 dist/releases/$(BINARY)-macos-intel
	@cd dist/releases && tar -czf $(BINARY)-macos-intel.tar.gz $(BINARY)-macos-intel
	# macOS Apple Silicon
	@cp dist/$(BINARY)-darwin-arm64 dist/releases/$(BINARY)-macos-apple-silicon
	@cd dist/releases && tar -czf $(BINARY)-macos-apple-silicon.tar.gz $(BINARY)-macos-apple-silicon
	@echo "✅ Platform-appropriate executables and archives created in dist/releases/"
	@echo "📝 Direct executables: ready to use immediately!"
	@echo "📦 Archives: traditional distribution format"

# Test release binaries and archives
test-releases:
	@echo "🧪 Testing release binaries and archives..."
	@./scripts/test-releases.sh

# Validate GitHub Actions workflows
validate-workflows:
	@echo "🔍 Validating GitHub Actions workflows..."
	@./scripts/validate-workflows.sh

## Test targets

# Run all tests
test:
	@echo "🧪 Running all tests..."
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run tests without integration tests
test-short:
	@echo "🧪 Running unit tests..."
	$(GOTEST) -short -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Generate test coverage report
test-coverage: test
	@echo "📊 Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

## Quality targets

# Run linter
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "⚠️  golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Format code
fmt:
	@echo "🎨 Formatting code..."
	$(GOCMD) fmt ./...
	@echo "✅ Code formatted"

# Run all checks
check: fmt lint test
	@echo "✅ All checks passed"

## Dependency management

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

# Update dependencies
deps-update:
	@echo "🔄 Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Clean module cache
deps-clean:
	@echo "🧹 Cleaning module cache..."
	$(GOCMD) clean -modcache

## Installation targets

# Install for development (in GOPATH/bin or current directory)
install-dev: build
	@echo "🔧 Installing GitCells for development..."
	@if [ -n "$(GOPATH)" ]; then \
		cp dist/$(BINARY) $(GOPATH)/bin/$(BINARY); \
		echo "✅ Installed to $(GOPATH)/bin/$(BINARY)"; \
	else \
		echo "✅ Binary available at dist/$(BINARY)"; \
	fi

# Install system-wide (requires sudo)
install: build
	@echo "🔧 Installing GitCells system-wide..."
	@if [ "$(shell id -u)" != "0" ]; then \
		echo "⚠️  System installation requires sudo. Run: sudo make install"; \
		exit 1; \
	fi
	@mkdir -p $(INSTALL_PATH)
	cp dist/$(BINARY) $(INSTALL_PATH)/$(BINARY)
	chmod 755 $(INSTALL_PATH)/$(BINARY)
	@echo "✅ Installed to $(INSTALL_PATH)/$(BINARY)"

# Uninstall system-wide
uninstall:
	@echo "🗑️  Uninstalling GitCells..."
	@if [ "$(shell id -u)" != "0" ]; then \
		echo "⚠️  System uninstall requires sudo. Run: sudo make uninstall"; \
		exit 1; \
	fi
	rm -f $(INSTALL_PATH)/$(BINARY)
	@echo "✅ Uninstalled from $(INSTALL_PATH)/$(BINARY)"

## Docker targets

# Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(VERSION)"

# Run Docker container
docker-run: docker-build
	@echo "🐳 Running Docker container..."
	docker run --rm -it $(DOCKER_IMAGE):$(VERSION)

# Push Docker image
docker-push: docker-build
	@echo "🐳 Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest
	@echo "✅ Docker image pushed"

## Utility targets

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf dist/ coverage.txt coverage.html
	$(GOCLEAN)
	@echo "✅ Clean complete"

## Documentation targets

.PHONY: docs docs-install docs-build docs-serve docs-deploy docs-open

# Build and serve documentation
docs: docs-build docs-serve

# Install documentation dependencies
docs-install:
	@echo "📚 Setting up documentation environment..."
	@if [ ! -d venv ]; then \
		python3 -m venv venv; \
	fi
	@. venv/bin/activate && pip install -r requirements.txt
	@echo "✅ Documentation dependencies installed in venv/"
	@echo "💡 To activate: source venv/bin/activate"

# Build documentation site
docs-build:
	@echo "📚 Building documentation with MkDocs..."
	@if [ -d venv ]; then \
		. venv/bin/activate && mkdocs build; \
	else \
		echo "⚠️  Virtual environment not found. Run 'make docs-install' first"; \
		exit 1; \
	fi
	@echo "✅ Documentation built in site/ directory"

# Serve documentation locally
docs-serve:
	@echo "🌐 Starting MkDocs development server..."
	@if [ -d venv ]; then \
		. venv/bin/activate && mkdocs serve; \
	else \
		echo "⚠️  Virtual environment not found. Run 'make docs-install' first"; \
		exit 1; \
	fi

# Deploy documentation
docs-deploy:
	@echo "🚀 Deploying documentation..."
	@if [ -d venv ]; then \
		. venv/bin/activate && mkdocs gh-deploy --clean; \
	else \
		echo "⚠️  Virtual environment not found. Run 'make docs-install' first"; \
		exit 1; \
	fi

# Open documentation
docs-open:
	@if [ -d site ]; then \
		python3 -c "import webbrowser; webbrowser.open('file://$(shell pwd)/site/index.html')"; \
		echo "📄 Opened documentation in browser"; \
	else \
		echo "⚠️  Documentation not built. Run 'make docs-build' first"; \
		exit 1; \
	fi

# Run the binary (for testing)
run: build
	@echo "🚀 Running GitCells..."
	./dist/$(BINARY) --help

# Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Commit Hash: $(COMMIT_HASH)"

# Development setup
dev-setup:
	@echo "🛠️  Setting up development environment..."
	$(GOMOD) download
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "✅ Development environment ready"

# Show help
help:
	@echo "GitCells Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build Targets:"
	@echo "  build              Build binary for current platform"
	@echo "  build-all          Build binaries for all platforms"
	@echo "  release            Create user-friendly executables (ready to use)"
	@echo "  test-releases      Test release binaries and archives"
	@echo "  validate-workflows Validate GitHub Actions workflows"
	@echo ""
	@echo "Test Targets:"
	@echo "  test           Run all tests with coverage"
	@echo "  test-short     Run unit tests only"
	@echo "  test-coverage  Generate HTML coverage report"
	@echo ""
	@echo "Quality Targets:"
	@echo "  lint   Run golangci-lint"
	@echo "  fmt    Format code with go fmt"
	@echo "  check  Run fmt, lint, and test"
	@echo ""
	@echo "Dependency Targets:"
	@echo "  deps         Download dependencies"
	@echo "  deps-update  Update all dependencies"
	@echo "  deps-clean   Clean module cache"
	@echo ""
	@echo "Installation Targets:"
	@echo "  install-dev  Install for development"
	@echo "  install      Install system-wide (requires sudo)"
	@echo "  uninstall    Uninstall system-wide (requires sudo)"
	@echo ""
	@echo "Docker Targets:"
	@echo "  docker-build  Build Docker image"
	@echo "  docker-run    Run Docker container"
	@echo "  docker-push   Push Docker image to registry"
	@echo ""
	@echo "Documentation Targets:"
	@echo "  docs         Build and serve documentation with MkDocs"
	@echo "  docs-install Install MkDocs and dependencies"
	@echo "  docs-build   Build static documentation site"
	@echo "  docs-serve   Serve documentation locally with live reload"
	@echo "  docs-deploy  Deploy documentation to GitHub Pages"
	@echo "  docs-open    Open built documentation in browser"
	@echo ""
	@echo "Utility Targets:"
	@echo "  clean      Clean build artifacts"
	@echo "  run        Run the binary"
	@echo "  version    Show version information"
	@echo "  dev-setup  Set up development environment"
	@echo "  help       Show this help message"

# Default goal
.DEFAULT_GOAL := all