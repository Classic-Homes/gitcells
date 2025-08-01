# Contributing to GitCells

Thank you for your interest in contributing to GitCells! This guide will help you get started with contributing to the project.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive criticism
- Respect differing viewpoints and experiences

## How to Contribute

### Reporting Issues

1. **Check existing issues** to avoid duplicates
2. **Use issue templates** when available
3. **Provide details**:
   - GitCells version (`gitcells version`)
   - Operating system and version
   - Steps to reproduce
   - Expected vs actual behavior
   - Error messages and logs

### Suggesting Features

1. **Open a discussion** first for major features
2. **Explain the use case** and benefits
3. **Consider implementation** complexity
4. **Be open to feedback** and alternatives

### Contributing Code

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/your-feature-name`
3. **Make your changes** following our guidelines
4. **Write tests** for new functionality
5. **Update documentation** as needed
6. **Submit a pull request**

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional but recommended)
- Docker (for testing)

### Setting Up Your Environment

1. **Fork and clone**:
```bash
git clone https://github.com/YOUR-USERNAME/gitcells.git
cd gitcells
```

2. **Add upstream remote**:
```bash
git remote add upstream https://github.com/Classic-Homes/gitcells.git
```

3. **Install dependencies**:
```bash
go mod download
```

4. **Install development tools**:
```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Install other tools
make install-tools
```

5. **Verify setup**:
```bash
make test
```

## Development Workflow

### 1. Create a Feature Branch

```bash
# Update your fork
git checkout main
git pull upstream main
git push origin main

# Create feature branch
git checkout -b feature/amazing-feature
```

### 2. Make Your Changes

Follow these guidelines:
- Write clear, concise code
- Follow Go conventions
- Add comments for complex logic
- Keep functions small and focused

### 3. Write Tests

All new code should have tests:

```go
func TestNewFeature(t *testing.T) {
    // Arrange
    input := "test data"
    expected := "expected result"
    
    // Act
    result := NewFeature(input)
    
    // Assert
    if result != expected {
        t.Errorf("Expected %s, got %s", expected, result)
    }
}
```

### 4. Run Tests Locally

```bash
# Run all tests
make test

# Run specific tests
go test ./internal/converter/...

# Run with coverage
make test-coverage

# Run benchmarks
make bench
```

### 5. Check Code Quality

```bash
# Run linter
make lint

# Format code
make fmt

# Run all checks
make check
```

### 6. Update Documentation

- Update relevant documentation in `/docs`
- Add/update code comments
- Update README if needed
- Add examples for new features

### 7. Commit Your Changes

Write clear commit messages:

```bash
# Good
git commit -m "feat: add support for Excel 2021 formulas"
git commit -m "fix: handle empty cells in pivot tables"
git commit -m "docs: update API reference for new converter options"

# Bad
git commit -m "fixed stuff"
git commit -m "updates"
```

Follow conventional commits:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes
- `refactor:` Code refactoring
- `test:` Test additions/changes
- `chore:` Build process/auxiliary changes

### 8. Push and Create Pull Request

```bash
git push origin feature/amazing-feature
```

Then create a pull request on GitHub with:
- Clear title describing the change
- Description of what and why
- Reference to related issues
- Screenshots if UI changes

## Coding Guidelines

### Go Style

Follow standard Go conventions:

```go
// Package comment describes the package
package converter

import (
    "fmt"
    "strings"
    
    "github.com/Classic-Homes/gitcells/pkg/models"
)

// ExportOptions configures the export behavior
type ExportOptions struct {
    Format      string // Output format
    Compress    bool   // Enable compression
}

// Export converts and exports the document
func Export(doc *models.Document, opts ExportOptions) error {
    // Validate input
    if doc == nil {
        return fmt.Errorf("document cannot be nil")
    }
    
    // Process based on format
    switch opts.Format {
    case "json":
        return exportJSON(doc, opts)
    case "csv":
        return exportCSV(doc, opts)
    default:
        return fmt.Errorf("unsupported format: %s", opts.Format)
    }
}
```

### Error Handling

Always handle errors appropriately:

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to parse cell %s: %w", cellRef, err)
}

// Define package errors
var (
    ErrInvalidFormat = errors.New("invalid file format")
    ErrFileTooLarge = errors.New("file exceeds size limit")
)

// Check specific errors
if errors.Is(err, ErrInvalidFormat) {
    // Handle invalid format
}
```

### Testing

Write comprehensive tests:

```go
func TestConverter_ExcelToJSON(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *models.Document
        wantErr bool
    }{
        {
            name:    "valid Excel file",
            input:   "testdata/valid.xlsx",
            want:    &models.Document{...},
            wantErr: false,
        },
        {
            name:    "corrupted file",
            input:   "testdata/corrupted.xlsx",
            want:    nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := NewConverter()
            got, err := c.ExcelToJSON(tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("ExcelToJSON() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ExcelToJSON() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Documentation

Document all exported types and functions:

```go
// Converter handles conversion between Excel and JSON formats.
// It supports various Excel features including formulas, styles,
// and charts while maintaining data integrity.
type Converter struct {
    logger *logrus.Logger
    cache  Cache
}

// NewConverter creates a new Converter instance with the given logger.
// If logger is nil, a default logger will be used.
//
// Example:
//   conv := converter.NewConverter(logger)
//   doc, err := conv.ExcelToJSON("data.xlsx")
func NewConverter(logger *logrus.Logger) *Converter {
    // Implementation
}
```

## Project Structure

When adding new features:

1. **Commands** go in `cmd/gitcells/`
2. **Core logic** goes in `internal/`
3. **Public APIs** go in `pkg/`
4. **Tests** go alongside the code
5. **Test data** goes in `testdata/`

## Pull Request Process

### Before Submitting

- [ ] Tests pass locally
- [ ] Code follows style guidelines
- [ ] Documentation is updated
- [ ] Commit messages are clear
- [ ] Branch is up to date with main

### Review Process

1. **Automated checks** run on all PRs
2. **Code review** by maintainers
3. **Address feedback** promptly
4. **Squash commits** if requested
5. **Merge** when approved

### After Merge

- Delete your feature branch
- Update your fork
- Celebrate your contribution! 🎉

## Testing Guidelines

### Unit Tests

Test individual functions:
```go
func TestCalculateChecksum(t *testing.T) {
    // Test implementation
}
```

### Integration Tests

Test component interactions:
```go
func TestWatcherWithConverter(t *testing.T) {
    // Test implementation
}
```

### Benchmarks

Measure performance:
```go
func BenchmarkLargeFileConversion(b *testing.B) {
    // Benchmark implementation
}
```

## Release Process

1. **Version bump** following semver
2. **Update changelog**
3. **Create release PR**
4. **Tag release** after merge
5. **Build binaries**
6. **Publish release**

## Getting Help

- **Discord**: Join our community server
- **Discussions**: Use GitHub Discussions
- **Issues**: For bugs and features
- **Email**: dev@gitcells.io

## Recognition

Contributors are recognized in:
- Release notes
- Contributors file
- Project website
- Annual contributor report

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (GNU GPL v3).

## Next Steps

- Set up your [development environment](building.md)
- Review the [architecture](architecture.md)
- Check the [testing guide](testing.md)
- Read existing code and tests
- Pick an issue labeled "good first issue"

Thank you for contributing to GitCells!