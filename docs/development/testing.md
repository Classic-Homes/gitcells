# Testing Guide

This guide covers testing practices, strategies, and tools used in GitCells development.

## Testing Philosophy

GitCells follows these testing principles:
- **Test Pyramid**: More unit tests, fewer integration tests, minimal E2E tests
- **Test Coverage**: Aim for 80%+ coverage for critical paths
- **Test Readability**: Tests should document behavior
- **Fast Feedback**: Tests should run quickly
- **Deterministic**: Tests should not be flaky

## Test Organization

```
gitcells/
├── cmd/gitcells/
│   └── commands_test.go      # Command-line interface tests
├── internal/
│   ├── converter/
│   │   ├── converter_test.go # Unit tests
│   │   ├── testdata/        # Test fixtures
│   │   └── benchmarks_test.go # Performance tests
│   └── watcher/
│       └── watcher_test.go   # Component tests
├── test/
│   ├── integration_test.go  # Integration tests
│   └── e2e_test.go         # End-to-end tests
└── testdata/               # Shared test data
```

## Running Tests

### Basic Commands

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run specific package tests
go test ./internal/converter/...

# Run single test
go test -run TestExcelToJSON ./internal/converter

# Run tests with race detector
go test -race ./...
```

### Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Check coverage percentage
go test -cover ./...
```

### Benchmarks

```bash
# Run all benchmarks
make bench

# Run specific benchmark
go test -bench=BenchmarkLargeFile ./internal/converter

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./internal/converter

# Compare benchmark results
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

## Writing Tests

### Unit Tests

Unit tests focus on individual functions or methods.

```go
package converter

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCellTypeDetection(t *testing.T) {
    tests := []struct {
        name     string
        value    string
        formula  string
        expected CellType
    }{
        {
            name:     "string value",
            value:    "Hello",
            formula:  "",
            expected: CellTypeString,
        },
        {
            name:     "number value",
            value:    "123.45",
            formula:  "",
            expected: CellTypeNumber,
        },
        {
            name:     "formula cell",
            value:    "150",
            formula:  "=SUM(A1:A10)",
            expected: CellTypeFormula,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := detectCellType(tt.value, tt.formula)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Integration Tests

Integration tests verify component interactions.

```go
package test

import (
    "io/ioutil"
    "os"
    "path/filepath"
    "testing"
    
    "github.com/Classic-Homes/gitcells/internal/converter"
    "github.com/Classic-Homes/gitcells/internal/watcher"
    "github.com/stretchr/testify/require"
)

func TestWatcherWithConverter(t *testing.T) {
    // Create temp directory
    tmpDir, err := ioutil.TempDir("", "gitcells-test-*")
    require.NoError(t, err)
    defer os.RemoveAll(tmpDir)
    
    // Setup converter
    conv := converter.NewConverter(testLogger)
    
    // Setup watcher
    config := &watcher.Config{
        FileExtensions: []string{".xlsx"},
        DebounceDelay:  100 * time.Millisecond,
    }
    
    converted := make(chan string, 1)
    handler := func(event watcher.FileEvent) error {
        err := conv.ExcelToJSONFile(event.Path, event.Path+".json", converter.ConvertOptions{})
        if err == nil {
            converted <- event.Path
        }
        return err
    }
    
    w, err := watcher.NewFileWatcher(config, handler, testLogger)
    require.NoError(t, err)
    
    // Start watching
    err = w.AddDirectory(tmpDir)
    require.NoError(t, err)
    
    err = w.Start()
    require.NoError(t, err)
    defer w.Stop()
    
    // Create test file
    testFile := filepath.Join(tmpDir, "test.xlsx")
    createTestExcelFile(t, testFile)
    
    // Wait for conversion
    select {
    case path := <-converted:
        assert.Equal(t, testFile, path)
        assert.FileExists(t, testFile+".json")
    case <-time.After(5 * time.Second):
        t.Fatal("Timeout waiting for conversion")
    }
}
```

### Table-Driven Tests

Use table-driven tests for comprehensive coverage.

```go
func TestValidateConfig(t *testing.T) {
    tests := map[string]struct {
        config  Config
        wantErr bool
        errMsg  string
    }{
        "valid config": {
            config: Config{
                Version: "1.0",
                Watcher: WatcherConfig{
                    Directories: []string{"."},
                },
            },
            wantErr: false,
        },
        "missing version": {
            config: Config{
                Watcher: WatcherConfig{
                    Directories: []string{"."},
                },
            },
            wantErr: true,
            errMsg:  "version is required",
        },
        "invalid debounce": {
            config: Config{
                Version: "1.0",
                Watcher: WatcherConfig{
                    DebounceDelay: -1 * time.Second,
                },
            },
            wantErr: true,
            errMsg:  "debounce delay must be positive",
        },
    }
    
    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            err := ValidateConfig(tc.config)
            if tc.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tc.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Testing with Mocks

Use interfaces and mocks for dependency injection.

```go
// Define interface
type GitClient interface {
    Commit(files []string, message string) error
    Push() error
}

// Mock implementation
type mockGitClient struct {
    mock.Mock
}

func (m *mockGitClient) Commit(files []string, message string) error {
    args := m.Called(files, message)
    return args.Error(0)
}

// Test using mock
func TestAutoCommit(t *testing.T) {
    mockGit := new(mockGitClient)
    mockGit.On("Commit", []string{"file.json"}, "Updated file.xlsx").Return(nil)
    
    service := NewService(mockGit)
    err := service.ProcessFile("file.xlsx")
    
    assert.NoError(t, err)
    mockGit.AssertExpectations(t)
}
```

### Testing Error Cases

Always test error scenarios.

```go
func TestConverterErrors(t *testing.T) {
    conv := NewConverter(testLogger)
    
    t.Run("file not found", func(t *testing.T) {
        _, err := conv.ExcelToJSON("nonexistent.xlsx", ConvertOptions{})
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "no such file")
    })
    
    t.Run("invalid format", func(t *testing.T) {
        _, err := conv.ExcelToJSON("testdata/invalid.txt", ConvertOptions{})
        assert.Error(t, err)
        assert.ErrorIs(t, err, ErrUnsupportedFormat)
    })
    
    t.Run("corrupted file", func(t *testing.T) {
        _, err := conv.ExcelToJSON("testdata/corrupted.xlsx", ConvertOptions{})
        assert.Error(t, err)
    })
}
```

## Test Fixtures

### Creating Test Data

```go
// testdata/fixtures.go
package testdata

import (
    "github.com/xuri/excelize/v2"
)

func CreateSimpleExcel(path string) error {
    f := excelize.NewFile()
    
    // Add data
    f.SetCellValue("Sheet1", "A1", "Name")
    f.SetCellValue("Sheet1", "B1", "Value")
    f.SetCellValue("Sheet1", "A2", "Test")
    f.SetCellValue("Sheet1", "B2", 123.45)
    
    // Add formula
    f.SetCellFormula("Sheet1", "B3", "=SUM(B2:B2)")
    
    return f.SaveAs(path)
}

func CreateComplexExcel(path string) error {
    f := excelize.NewFile()
    
    // Add multiple sheets
    f.NewSheet("Data")
    f.NewSheet("Charts")
    
    // Add styles
    style, _ := f.NewStyle(&excelize.Style{
        Font: &excelize.Font{Bold: true},
        Fill: excelize.Fill{Type: "pattern", Color: []string{"#FF0000"}},
    })
    f.SetCellStyle("Sheet1", "A1", "B1", style)
    
    // Add chart
    f.AddChart("Charts", "A1", &excelize.Chart{
        Type: excelize.Col,
        Series: []excelize.ChartSeries{{
            Name:       "Sales",
            Categories: "Data!$A$2:$A$4",
            Values:     "Data!$B$2:$B$4",
        }},
    })
    
    return f.SaveAs(path)
}
```

### Testing Chart Detection

Chart detection works through pattern analysis of tabular data, not chart objects.

```go
func TestChartDetection(t *testing.T) {
    // Create test data that should trigger chart detection
    f := excelize.NewFile()
    
    // Headers
    f.SetCellValue("Sheet1", "A1", "Month")
    f.SetCellValue("Sheet1", "B1", "Sales")
    f.SetCellValue("Sheet1", "C1", "Profit")
    
    // Data rows with numeric values
    data := [][]interface{}{
        {"Jan", 1000, 200},
        {"Feb", 1200, 300},
        {"Mar", 1500, 400},
    }
    
    for i, row := range data {
        for j, value := range row {
            cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
            f.SetCellValue("Sheet1", cell, value)
        }
    }
    
    // Test conversion with chart detection enabled
    conv := converter.NewConverter(logger)
    options := converter.ConvertOptions{
        PreserveCharts: true,
    }
    
    doc, err := conv.ExcelToJSON("test.xlsx", options)
    require.NoError(t, err)
    
    // Verify chart was detected
    require.Len(t, doc.Sheets, 1)
    sheet := doc.Sheets[0]
    require.Len(t, sheet.Charts, 1)
    
    // Verify chart properties
    chart := sheet.Charts[0]
    assert.Equal(t, "column", chart.Type)
    assert.Contains(t, chart.ID, "chart_Sheet1_")
    assert.Len(t, chart.Series, 2) // Sales and Profit columns
    
    // Verify series data
    assert.Equal(t, "Sales", chart.Series[0].Name)
    assert.Equal(t, "Profit", chart.Series[1].Name)
    assert.Contains(t, chart.Series[0].Values, "B2:")
    assert.Contains(t, chart.Series[1].Values, "C2:")
}
```

#### Chart Detection Criteria

Charts are detected when:
- First row contains headers (text values)
- At least 2 numeric columns are present
- At least 2 data rows exist
- Numeric values are consistent in columns

#### Testing False Positives

```go
func TestNoChartDetection(t *testing.T) {
    f := excelize.NewFile()
    
    // Only one numeric column - should not detect chart
    f.SetCellValue("Sheet1", "A1", "Name")
    f.SetCellValue("Sheet1", "B1", "Age")
    f.SetCellValue("Sheet1", "A2", "John")
    f.SetCellValue("Sheet1", "B2", 25)
    
    // Convert and verify no charts detected
    doc, err := convertWithCharts(f)
    require.NoError(t, err)
    assert.Empty(t, doc.Sheets[0].Charts)
}
```

### Using Golden Files

Store expected outputs for comparison.

```go
func TestGoldenFiles(t *testing.T) {
    testCases := []string{
        "simple.xlsx",
        "formulas.xlsx",
        "styles.xlsx",
    }
    
    for _, tc := range testCases {
        t.Run(tc, func(t *testing.T) {
            // Convert Excel to JSON
            result, err := convertToJSON(filepath.Join("testdata", tc))
            require.NoError(t, err)
            
            // Compare with golden file
            goldenPath := filepath.Join("testdata", "golden", tc+".json")
            
            if *update {
                // Update golden files with -update flag
                err := ioutil.WriteFile(goldenPath, result, 0644)
                require.NoError(t, err)
            } else {
                // Compare with existing golden file
                expected, err := ioutil.ReadFile(goldenPath)
                require.NoError(t, err)
                assert.JSONEq(t, string(expected), string(result))
            }
        })
    }
}
```

## Performance Testing

### Benchmarks

```go
func BenchmarkExcelToJSON(b *testing.B) {
    sizes := []struct {
        name string
        rows int
        cols int
    }{
        {"small", 100, 10},
        {"medium", 1000, 50},
        {"large", 10000, 100},
    }
    
    for _, size := range sizes {
        b.Run(size.name, func(b *testing.B) {
            // Create test file
            file := createBenchmarkFile(b, size.rows, size.cols)
            defer os.Remove(file)
            
            conv := NewConverter(nil)
            opts := ConvertOptions{}
            
            b.ResetTimer()
            b.ReportAllocs()
            
            for i := 0; i < b.N; i++ {
                _, err := conv.ExcelToJSON(file, opts)
                if err != nil {
                    b.Fatal(err)
                }
            }
            
            b.ReportMetric(float64(size.rows*size.cols)/float64(b.Elapsed().Seconds()), "cells/sec")
        })
    }
}
```

### Load Testing

```go
func TestConcurrentConversion(t *testing.T) {
    const numFiles = 10
    const numWorkers = 4
    
    // Create test files
    files := make([]string, numFiles)
    for i := 0; i < numFiles; i++ {
        files[i] = createTestFile(t, fmt.Sprintf("test%d.xlsx", i))
        defer os.Remove(files[i])
    }
    
    // Run concurrent conversions
    conv := NewConverter(testLogger)
    errors := make(chan error, numFiles)
    
    var wg sync.WaitGroup
    sem := make(chan struct{}, numWorkers)
    
    start := time.Now()
    
    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()
            
            _, err := conv.ExcelToJSON(f, ConvertOptions{})
            if err != nil {
                errors <- err
            }
        }(file)
    }
    
    wg.Wait()
    close(errors)
    
    elapsed := time.Since(start)
    t.Logf("Converted %d files in %v", numFiles, elapsed)
    
    // Check for errors
    for err := range errors {
        t.Errorf("Conversion error: %v", err)
    }
}
```

## Testing Best Practices

### Test Naming

Use descriptive test names:
```go
// Good
func TestConverter_ExcelToJSON_PreservesFormulas(t *testing.T)
func TestWatcher_IgnoresTemporaryFiles(t *testing.T)

// Bad
func TestConvert(t *testing.T)
func Test1(t *testing.T)
```

### Test Independence

Tests should not depend on each other:
```go
func TestIndependent(t *testing.T) {
    // Setup - create all needed resources
    tmpDir := createTempDir(t)
    defer os.RemoveAll(tmpDir)
    
    // Test logic
    
    // No cleanup needed - defer handles it
}
```

### Assertions

Use appropriate assertions:
```go
// Use require for critical checks
require.NoError(t, err)
require.NotNil(t, result)

// Use assert for non-critical checks
assert.Equal(t, expected, actual)
assert.Contains(t, haystack, needle)

// Custom assertions
assert.Eventually(t, func() bool {
    return fileExists("output.json")
}, 5*time.Second, 100*time.Millisecond)
```

### Test Helpers

Create reusable test helpers:
```go
func createTestExcel(t *testing.T, name string) string {
    t.Helper()
    
    path := filepath.Join(t.TempDir(), name)
    err := CreateSimpleExcel(path)
    require.NoError(t, err)
    
    return path
}

func assertJSONEqual(t *testing.T, expected, actual string) {
    t.Helper()
    
    var exp, act interface{}
    require.NoError(t, json.Unmarshal([]byte(expected), &exp))
    require.NoError(t, json.Unmarshal([]byte(actual), &act))
    
    assert.Equal(t, exp, act)
}
```

## Continuous Integration

### GitHub Actions

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        go: ['1.20', '1.21']
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go }}
      
      - name: Run tests
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

### Pre-commit Hooks

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Run tests
go test ./... || exit 1

# Check formatting
gofmt -l . | grep -v vendor
if [ $? -eq 0 ]; then
    echo "Code needs formatting"
    exit 1
fi

# Run linter
golangci-lint run || exit 1
```

## Testing Tools

### Required Tools

```bash
# Test runner with colors
go install github.com/rakyll/gotest@latest

# Assertion library
go get github.com/stretchr/testify

# Mocking framework
go get github.com/stretchr/testify/mock

# Coverage visualization
go install github.com/axw/gocov/gocov@latest
go install github.com/AlekSi/gocov-xml@latest
```

### Useful Tools

```bash
# Test coverage in terminal
go install github.com/mfridman/tparse@latest
go test -json ./... | tparse

# Mutation testing
go install github.com/zimmski/go-mutesting/cmd/go-mutesting@latest

# Fuzz testing
go test -fuzz=FuzzParse
```

## Debugging Tests

### Verbose Output

```go
func TestWithLogging(t *testing.T) {
    if testing.Verbose() {
        t.Logf("Debug info: %v", someValue)
    }
}
```

### Skip Slow Tests

```go
func TestSlowOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping slow test in short mode")
    }
    // Slow test logic
}
```

### Isolate Failing Tests

```bash
# Run single test
go test -run TestSpecificFunction ./pkg/...

# Run with more verbosity
go test -v -run TestSpecificFunction ./pkg/...

# Debug with delve
dlv test ./pkg/... -- -test.run TestSpecificFunction
```

## Next Steps

- Review [Contributing Guide](contributing.md) for test requirements
- Check [Building Guide](building.md) for running tests locally
- See [Architecture](architecture.md) for testing strategies
- Read existing tests for examples