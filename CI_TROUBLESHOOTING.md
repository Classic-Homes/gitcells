# CI/CD Troubleshooting Guide

## Issue: Linting and Testing Pass Locally but Fail in CI

### Root Causes Identified:

1. **Go Version Mismatch**
   - Local: Go 1.24.5
   - CI: Go 1.23
   - Solution: Ensure local development uses Go 1.23 to match CI

2. **golangci-lint Version Differences**
   - Local: v1.64.8 (built with Go 1.24.1)
   - CI: Installs `@latest` at runtime
   - Solution: Pin golangci-lint to v1.61.0 in CI (compatible with Go 1.23)

3. **Missing Linter Configuration**
   - No `.golangci.yml` file to ensure consistent linting rules
   - Solution: Added `.golangci.yml` configuration file

4. **Unchecked Errors**
   - Several instances of unchecked error returns
   - Solution: Fixed error handling in affected files

## Changes Made:

### 1. Created `.golangci.yml` Configuration
```yaml
run:
  go: "1.23"
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - gosec
    - misspell
    - unconvert
    - prealloc
    - nakedret
    - exhaustive
    - gocritic

issues:
  max-issues-per-linter: 0
  max-same-issues: 0
```

### 2. Updated CI Workflow
- Pinned golangci-lint to v1.61.0 instead of using `@latest`
- This ensures consistent behavior between environments

### 3. Fixed Error Handling
- `internal/tui/validation/validators.go`: Fixed unchecked filepath.Walk error
- `internal/watcher/debouncer.go`: Fixed unchecked timer.Stop() error

## Remaining Linting Issues to Address:

1. **Security (gosec)**:
   - File permissions should be 0600 or less for sensitive files
   - Potential DoS via decompression bomb in updater.go

2. **Code Quality**:
   - Missing switch cases (exhaustive linter)
   - Single-case switches should be if statements (gocritic)
   - If-else chains should be switch statements (gocritic)
   - Pre-allocation opportunities for slices (prealloc)

## Recommendations:

1. **Local Development Setup**:
   ```bash
   # Install Go 1.23 to match CI
   brew install go@1.23
   
   # Install the same golangci-lint version as CI
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
   ```

2. **Pre-commit Checks**:
   ```bash
   # Add to your workflow
   make fmt      # Format code
   make lint     # Run linter
   make test     # Run tests
   ```

3. **CI Debugging**:
   - When CI fails, check the specific error messages
   - Run the same commands locally with the same versions
   - Use the `.golangci.yml` config to ensure consistency

## Testing the Fixes:

1. Run linting locally:
   ```bash
   golangci-lint run --config .golangci.yml
   ```

2. Run tests in short mode (as CI does):
   ```bash
   go test -short -v ./...
   ```

3. Simulate CI environment:
   ```bash
   docker run -it golang:1.23 bash
   # Inside container:
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
   # Clone repo and run make check
   ```

## Next Steps:

1. Fix the remaining linting issues identified above
2. Consider adding a pre-commit hook to catch issues early
3. Update developer documentation with setup instructions
4. Consider using a tool like `asdf` or `gvm` to manage Go versions