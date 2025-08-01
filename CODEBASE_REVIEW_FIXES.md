# GitCells Codebase Review - Required Fixes

## Critical Priority Fixes

### 1. Remove Duplicate TUI Models
**Issue**: Duplicate basic and enhanced versions of TUI models
**Location**: `internal/tui/models/`

**Fix**: Consolidate to single implementation
- [ ] Merge `dashboard.go` and `dashboard_enhanced.go` into single `dashboard.go`
- [ ] Merge `error_log.go` and `error_log_enhanced.go` into single `error_log.go`
- [ ] Merge `setup.go` and `setup_enhanced.go` into single `setup.go`
- [ ] Update all imports and references throughout the codebase
- [ ] Remove the unused model files

### 2. Implement Missing CLI Commands
**Issue**: Status and sync commands are not implemented
**Location**: `cmd/gitcells/status.go`, `cmd/gitcells/sync.go`

**Fix**: Implement command functionality
- [ ] Implement `status` command to show:
  - List of tracked Excel files
  - Last sync time for each file
  - Pending changes
  - Git status integration
- [ ] Implement `sync` command to:
  - Convert all modified Excel files to JSON
  - Stage changes in Git
  - Create commit with configured message template
  - Optional push to remote

### 3. Complete Init Command
**Issue**: Init command has TODO for git repository initialization
**Location**: `cmd/gitcells/init.go`

**Fix**: Add git initialization logic
- [ ] Check if directory is already a git repo
- [ ] Initialize git repo if needed
- [ ] Create initial .gitignore with Excel temp files pattern
- [ ] Create initial commit

## Major Priority Fixes

### 4. Refactor Large Functions
**Issue**: Functions exceeding 100 lines with multiple responsibilities
**Location**: `internal/converter/types.go`

**Fix**: Break down `extractChartsFromExcelFile()`:
- [ ] Extract chart detection logic into `detectCharts()`
- [ ] Extract data pattern analysis into `analyzeDataPatterns()`
- [ ] Extract chart type inference into `inferChartType()`
- [ ] Extract series creation into `createChartSeries()`

### 5. Standardize Error Handling
**Issue**: Inconsistent use of custom errors vs standard errors
**Location**: Throughout codebase

**Fix**: Use consistent error pattern
- [ ] Always use `utils.WrapError()` for internal errors
- [ ] Define error types for each package
- [ ] Update all error returns to use consistent format
- [ ] Add error documentation

## Medium Priority Fixes

### 6. Centralize Constants
**Issue**: Repeated constants across files
**Location**: Multiple files

**Fix**: Create central constants package
- [x] Create `internal/constants/files.go` for file-related constants
- [x] Move all file permissions to constants
- [x] Move all file extensions to constants
- [x] Update all references

### 7. Add Missing Tests
**Issue**: Limited test coverage for some components
**Location**: `internal/tui/models/`, `cmd/gitcells/`

**Fix**: Add comprehensive tests
- [ ] Add tests for all TUI models
- [ ] Add tests for CLI command logic
- [ ] Add integration tests for git operations
- [ ] Achieve minimum 80% coverage

### 8. Improve Path Security
**Issue**: Limited path traversal validation
**Location**: File operations throughout

**Fix**: Add path validation utility
- [ ] Create `utils.ValidatePath()` function
- [ ] Check for path traversal attempts
- [ ] Validate paths are within expected directories
- [ ] Apply to all file operations

## Low Priority Fixes

### 9. Standardize Import Organization
**Issue**: Inconsistent import grouping
**Location**: All Go files

**Fix**: Apply consistent import pattern
- [ ] Standard library imports first
- [ ] Empty line
- [ ] External dependencies
- [ ] Empty line
- [ ] Internal packages

### 10. Consistent Error Messages
**Issue**: Different error message formats
**Location**: Throughout codebase

**Fix**: Standardize format
- [ ] Use format: "failed to [action]: %w"
- [ ] Update all error messages
- [ ] Document standard in CONTRIBUTING.md

## Implementation Order

1. **Week 1**: Critical fixes (1-3)
   - Remove duplicate TUI models
   - Implement status command
   - Implement sync command
   - Complete init command

2. **Week 2**: Major fixes (4-5)
   - Refactor large functions
   - Standardize error handling

3. **Week 3**: Medium fixes (6-8)
   - Centralize constants
   - Add missing tests
   - Improve path security

4. **Week 4**: Low priority fixes (9-10)
   - Standardize imports
   - Consistent error messages

## Testing Strategy

After each fix:
1. Run `make test` to ensure no regressions
2. Run `make lint` to check code quality
3. Test affected commands manually
4. Update documentation if needed

## Success Criteria

- [ ] All TODO comments replaced with implementation
- [ ] No duplicate model files
- [ ] Test coverage > 80%
- [ ] All linting checks pass
- [ ] Documentation updated