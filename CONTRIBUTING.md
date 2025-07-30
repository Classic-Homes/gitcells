# Contributing to SheetSync

Thank you for your interest in contributing to SheetSync! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Quality](#code-quality)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Make (for build automation)
- Docker (optional, for containerized development)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/sheetsync.git
   cd sheetsync
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/chrisloidolt/sheetsync.git
   ```

## Development Setup

### Install Dependencies

```bash
go mod download
go mod verify
```

### Build the Project

```bash
make build
```

Or manually:
```bash
go build -o dist/sheetsync cmd/sheetsync/main.go
```

### Run Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run integration tests
go test -v -tags=integration ./test/...
```

### Development Commands

```bash
# Build for all platforms
make build-all

# Run linter
golangci-lint run

# Clean build artifacts
make clean

# Run security scan
gosec ./...
```

## Making Changes

### Branching Strategy

- `main`: Stable production code
- `develop`: Integration branch for new features
- Feature branches: `feature/your-feature-name`
- Bug fixes: `bugfix/issue-description`
- Hotfixes: `hotfix/critical-fix`

### Creating a Feature Branch

```bash
git checkout develop
git pull upstream develop
git checkout -b feature/your-feature-name
```

### Code Standards

#### Go Style Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` and `goimports` for formatting
- Follow Go naming conventions
- Write clear, self-documenting code
- Use meaningful variable and function names

#### Project Structure

```
cmd/sheetsync/          # CLI entry points and commands
internal/               # Private application code
├── config/            # Configuration management
├── converter/         # Excel↔JSON conversion logic
├── git/              # Git operations
├── utils/            # Shared utilities
└── watcher/          # File system monitoring
pkg/models/            # Public data models
test/                  # Integration tests and test data
```

#### Import Organization

Group imports in this order:
1. Standard library
2. Third-party packages
3. Local project packages

```go
import (
    "context"
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/xuri/excelize/v2"

    "github.com/chrisloidolt/sheetsync/internal/config"
    "github.com/chrisloidolt/sheetsync/pkg/models"
)
```

#### Error Handling

- Use explicit error handling, avoid `panic()`
- Wrap errors with context using `fmt.Errorf()`
- Log errors at appropriate levels
- Return meaningful error messages to users

```go
if err != nil {
    return fmt.Errorf("failed to convert Excel file: %w", err)
}
```

#### Documentation

- Document all public functions and types
- Use clear, concise comments
- Include usage examples for complex functions
- Update README.md for user-facing changes

## Testing

### Test Categories

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete workflows

### Writing Tests

- Place test files alongside the code they test
- Use `_test.go` suffix for test files
- Follow the pattern: `TestFunctionName`
- Use table-driven tests for multiple scenarios
- Include both positive and negative test cases

```go
func TestExcelToJSON(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected *models.ExcelDocument
        wantErr  bool
    }{
        {
            name:     "simple xlsx file",
            input:    "testdata/simple.xlsx",
            expected: &models.ExcelDocument{...},
            wantErr:  false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ExcelToJSON(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ExcelToJSON() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            // Assert results...
        })
    }
}
```

### Test Data

- Place test files in `test/testdata/`
- Use small, focused test files
- Don't commit sensitive data
- Document test data format and purpose

## Code Quality

### Linting

We use `golangci-lint` with a comprehensive configuration:

```bash
golangci-lint run
```

### Security

- Run security scans: `gosec ./...`
- Avoid hardcoded secrets
- Use secure defaults
- Validate all inputs

### Performance

- Profile performance-critical code
- Avoid premature optimization
- Use benchmarks for performance tests
- Consider memory usage for large files

```go
func BenchmarkExcelToJSON(b *testing.B) {
    for i := 0; i < b.N; i++ {
        ExcelToJSON("testdata/large.xlsx")
    }
}
```

## Submitting Changes

### Before Submitting

1. **Update your branch**:
   ```bash
   git checkout develop
   git pull upstream develop
   git checkout your-feature-branch
   git rebase develop
   ```

2. **Run all checks**:
   ```bash
   go test ./...
   golangci-lint run
   gosec ./...
   ```

3. **Update documentation** if needed

### Commit Messages

Use conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Build/tooling changes

Examples:
- `feat(converter): add support for merged cells`
- `fix(git): handle file conflicts correctly`
- `docs: update installation instructions`

### Pull Request Process

1. **Create Pull Request**:
   - Use a descriptive title
   - Reference related issues
   - Provide clear description of changes
   - Include testing instructions

2. **PR Template**:
   ```markdown
   ## Description
   Brief description of changes

   ## Changes Made
   - List of specific changes
   - Another change

   ## Testing
   - [ ] Unit tests pass
   - [ ] Integration tests pass
   - [ ] Manual testing completed

   ## Checklist
   - [ ] Code follows project standards
   - [ ] Documentation updated
   - [ ] Tests added/updated
   - [ ] No breaking changes (or documented)
   ```

3. **Code Review**:
   - Address reviewer feedback
   - Keep discussions professional
   - Update code based on suggestions

4. **Merge Requirements**:
   - All CI checks pass
   - At least one approved review
   - Up-to-date with target branch
   - No merge conflicts

## Release Process

### Version Management

We use semantic versioning (SemVer):
- `MAJOR.MINOR.PATCH`
- Breaking changes: increment MAJOR
- New features: increment MINOR
- Bug fixes: increment PATCH

### Release Steps

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create release PR to `main`
4. Tag release after merge
5. GitHub Actions handles build and deployment

## Getting Help

- **Issues**: Report bugs and request features
- **Discussions**: Ask questions and share ideas
- **Email**: Contact maintainers for sensitive issues

## Recognition

Contributors are recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project documentation

Thank you for contributing to SheetSync! 🎉