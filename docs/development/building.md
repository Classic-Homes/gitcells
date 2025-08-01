# Building GitCells from Source

This guide explains how to build GitCells from source code for development or custom deployments.

## Prerequisites

### Required Tools

- **Go**: Version 1.21 or higher
- **Git**: For cloning the repository
- **Make**: For using the build scripts (optional but recommended)

### Optional Tools

- **Docker**: For building in containers
- **golangci-lint**: For code quality checks
- **goreleaser**: For release builds

## Getting the Source Code

### Clone the Repository

```bash
# Clone via HTTPS
git clone https://github.com/Classic-Homes/gitcells.git

# Or clone via SSH
git clone git@github.com:Classic-Homes/gitcells.git

# Enter the directory
cd gitcells
```

### Repository Structure

```
gitcells/
├── cmd/gitcells/       # Main application entry point
├── internal/           # Private application packages
├── pkg/               # Public packages
├── scripts/           # Build and utility scripts
├── Makefile          # Build automation
├── go.mod            # Go module definition
└── go.sum            # Go module checksums
```

## Building with Make

The easiest way to build GitCells is using the provided Makefile.

### Basic Build

```bash
# Build for current platform
make build

# Output binary will be in dist/
ls dist/
# gitcells-darwin-amd64 (on macOS)
# gitcells-linux-amd64 (on Linux)
# gitcells-windows-amd64.exe (on Windows)
```

### Build All Platforms

```bash
# Build for all supported platforms
make build-all

# Creates binaries for:
# - macOS (Intel and Apple Silicon)
# - Linux (AMD64 and ARM64)
# - Windows (AMD64)
```

### Development Build

```bash
# Quick build for development (current platform only)
make dev

# Build with race detector
make build-race

# Build with debug symbols
make build-debug
```

## Building Manually

If you prefer not to use Make, you can build directly with Go.

### Simple Build

```bash
# Build for current platform
go build -o gitcells ./cmd/gitcells

# Run the built binary
./gitcells version
```

### Production Build

```bash
# Build with version information
VERSION=$(git describe --tags --always --dirty)
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build -ldflags "-X main.version=$VERSION -X main.buildTime=$BUILD_TIME" \
    -o gitcells ./cmd/gitcells
```

### Cross-Platform Build

```bash
# Build for macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o gitcells-darwin-amd64 ./cmd/gitcells

# Build for macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o gitcells-darwin-arm64 ./cmd/gitcells

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o gitcells-linux-amd64 ./cmd/gitcells

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o gitcells-windows-amd64.exe ./cmd/gitcells
```

## Build Options

### Optimization Flags

```bash
# Optimize for size
go build -ldflags="-s -w" -o gitcells ./cmd/gitcells

# Enable compiler optimizations
go build -gcflags="-l=4" -o gitcells ./cmd/gitcells
```

### Static Linking

```bash
# Build statically linked binary (Linux)
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' \
    -o gitcells-static ./cmd/gitcells
```

### Build Tags

```bash
# Build with specific features
go build -tags "embed_files" -o gitcells ./cmd/gitcells

# Build without certain features
go build -tags "no_cgo" -o gitcells ./cmd/gitcells
```

## Docker Build

Build GitCells in a Docker container for consistent environments.

### Using Dockerfile

```bash
# Build Docker image
docker build -t gitcells:latest .

# Extract binary from container
docker create --name extract gitcells:latest
docker cp extract:/usr/local/bin/gitcells ./gitcells
docker rm extract
```

### Multi-Stage Build

```dockerfile
# Dockerfile for multi-stage build
FROM golang:1.21-alpine AS builder

WORKDIR /build
COPY . .
RUN go mod download
RUN go build -o gitcells ./cmd/gitcells

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /build/gitcells /usr/local/bin/
ENTRYPOINT ["gitcells"]
```

## Development Workflow

### Quick Development Cycle

```bash
# Install for development
make install

# Run tests
make test

# Run with hot reload (requires air)
air -c .air.toml
```

### Running from Source

```bash
# Run directly without building
go run ./cmd/gitcells version

# Run with arguments
go run ./cmd/gitcells watch --verbose .
```

### Debug Build

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o gitcells-debug ./cmd/gitcells

# Run with delve debugger
dlv exec ./gitcells-debug -- watch .
```

## Testing the Build

### Verify Build

```bash
# Check version
./gitcells version

# Run basic command
./gitcells init --help

# Test conversion
./gitcells convert test.xlsx
```

### Run Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/converter/...

# Run with coverage
make test-coverage

# Run benchmarks
make bench
```

## Build Automation

### GitHub Actions

The project uses GitHub Actions for CI/CD:

```yaml
# .github/workflows/build.yml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: make build-all
      - run: make test
```

### Local CI Testing

```bash
# Test CI locally with act
act -j build

# Run specific workflow
act -W .github/workflows/release.yml
```

## Release Builds

### Using GoReleaser

```bash
# Install goreleaser
go install github.com/goreleaser/goreleaser@latest

# Create snapshot release (no publishing)
goreleaser release --snapshot --clean

# Create and publish release
goreleaser release --clean
```

### Manual Release Build

```bash
# Set version
VERSION=v0.3.0

# Build all platforms with version
make release VERSION=$VERSION

# Creates:
# - dist/gitcells-darwin-amd64-v0.3.0
# - dist/gitcells-linux-amd64-v0.3.0
# - dist/gitcells-windows-amd64-v0.3.0.exe
```

## Build Troubleshooting

### Common Issues

1. **Module errors**:
```bash
go mod download
go mod tidy
```

2. **Build cache issues**:
```bash
go clean -cache
go clean -modcache
```

3. **Cross-compilation errors**:
```bash
# Install cross-compilation support
go env -w GO111MODULE=on
go env -w CGO_ENABLED=0
```

### Platform-Specific Issues

#### macOS
- Ensure Xcode Command Line Tools are installed
- Handle code signing for distribution

#### Windows
- Use Git Bash or WSL for Make commands
- Add `.exe` extension to output

#### Linux
- Install build-essential for CGO support
- Check GLIBC version compatibility

## Performance Optimization

### Build Optimization

```bash
# Profile-guided optimization
go build -pgo=cpu.pprof -o gitcells-pgo ./cmd/gitcells

# Link-time optimization
go build -ldflags="-s -w" -trimpath -o gitcells-opt ./cmd/gitcells
```

### Binary Size Reduction

```bash
# Strip debug information
strip gitcells

# Use UPX compression (optional)
upx --best gitcells
```

## Security Considerations

### Secure Builds

```bash
# Enable security features
go build -buildmode=pie -o gitcells ./cmd/gitcells

# Verify dependencies
go mod verify

# Check for vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

### Reproducible Builds

```bash
# Set consistent build environment
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export SOURCE_DATE_EPOCH=$(git log -1 --pretty=%ct)

# Build with trimpath
go build -trimpath -o gitcells ./cmd/gitcells
```

## Next Steps

- Review [Contributing Guide](contributing.md) for development workflow
- Check [Testing Guide](testing.md) for running tests
- See [Architecture](architecture.md) for system design
- View [GitHub Releases](https://github.com/Classic-Homes/gitcells/releases) for publishing