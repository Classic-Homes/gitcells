# AGENTS.md - Quick Reference for Coding Agents

## Build/Test/Lint Commands
- `make build` - Build for current platform
- `make test` - Run all tests with coverage
- `make test-short` - Run unit tests only (skip integration tests)
- `go test ./internal/converter -v` - Test specific package
- `go test -run TestFunctionName ./path/to/package` - Run single test
- `make lint` - Run golangci-lint
- `make fmt` - Format code with go fmt
- `make check` - Run fmt, lint, and test together

## Code Style Guidelines
- **Imports**: Group stdlib, external deps, internal packages (use goimports)
- **Formatting**: Use `go fmt`, tabs for indentation
- **Types**: Define interfaces in consumer packages, implement in provider packages
- **Naming**: Use camelCase for private, PascalCase for public; short names for small scopes
- **Error Handling**: Wrap errors with context using custom `utils.WrapError()`, return nil for success
- **Comments**: Package comment on package line, document all exported types/funcs with GoDoc format
- **Testing**: Use testify/assert, table-driven tests, test files named `*_test.go`
- **Logging**: Use logrus with structured fields, not fmt.Println
- **Constants**: Group related constants with const blocks, use typed constants
- **File Paths**: Always use `filepath` package for cross-platform compatibility
