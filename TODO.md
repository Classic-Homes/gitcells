# GitCells TODO List

## Overview
This document tracks all pending tasks, stub implementations, and future enhancements identified in the GitCells codebase. The project is production-ready with all core functionality implemented. These items represent advanced features and optimizations.

## Tasks by Priority

### High Priority
None - all critical functionality is implemented.

### Medium Priority

#### 1. Apply Cell Styles in JSON to Excel Conversion [DONE]
- **Location:** `internal/converter/json_to_excel.go:62-64`
- **Status:** Cell styles are extracted but not applied during reconstruction
- **Impact:** Excel files lose formatting when converted from JSON back to Excel
- **Implementation:** Need to apply the style data that's already being extracted

#### 2. TUI Side-by-Side Diff View [DONE]
- **Location:** `internal/tui/components/diff.go:356`
- **Status:** Shows placeholder text "Select a cell to compare"
- **Impact:** Users can't see actual cell-by-cell comparisons in the TUI
- **Implementation:** Replace placeholder with actual diff rendering logic

#### 3. Chart Extraction from Excel Files
- **Location:** `internal/converter/types.go:313-320`
- **Status:** Framework exists but returns empty slice
- **Impact:** Charts are lost during Excel to JSON conversion
- **Dependencies:** Requires excelize library support

#### 4. Pivot Table Extraction from Excel Files
- **Location:** `internal/converter/types.go:327-333`
- **Status:** Framework exists but returns empty slice
- **Impact:** Pivot tables are lost during Excel to JSON conversion
- **Dependencies:** Requires excelize library support

#### 5. Sheet-Specific Conversion Feature
- **Location:** `docs/user-guide/converting.md:192-194`
- **Status:** Documented as future feature
- **Impact:** Users must convert entire workbooks even if only need specific sheets
- **Implementation:** Add `--sheets` flag to convert command

### Low Priority

#### Excel Style Extraction (Blocked by excelize library)
These features are waiting for upstream library support:

1. **R1C1 Formula Extraction**
   - **Location:** `internal/converter/types.go:97-98`
   - **Status:** TODO comment, excelize doesn't support R1C1

2. **Array Formula Range Detection**
   - **Location:** `internal/converter/types.go:104-107`
   - **Status:** Using cell reference instead of actual range

3. **Number Format Extraction**
   - **Location:** `internal/converter/types.go:135-136`
   - **Status:** excelize doesn't provide direct access

4. **Font Information Extraction**
   - **Location:** `internal/converter/types.go:139`
   - **Status:** Hardcoded to Calibri, size 11

5. **Fill/Background Color Extraction**
   - **Location:** `internal/converter/types.go:146`
   - **Status:** Not implemented

6. **Border Extraction**
   - **Location:** `internal/converter/types.go:149`
   - **Status:** Not implemented

7. **Alignment Extraction**
   - **Location:** `internal/converter/types.go:152`
   - **Status:** Not implemented

#### Future Enhancements

1. **Hybrid Chunking Strategy**
   - **Location:** `internal/converter/chunking.go:347-370`
   - **Status:** All methods return "not yet implemented" error
   - **Purpose:** Optimize large sheet splitting by ranges
   - **Impact:** Better performance for very large spreadsheets

2. **Plugin Architecture**
   - **Location:** `API.md:592`
   - **Status:** Documented but not implemented
   - **Purpose:** Support custom file formats and VCS systems
   - **Impact:** Extensibility for enterprise use cases

## Implementation Plan

### Phase 1: Quick Wins (Can be done now)
1. [ ] Apply cell styles during JSON to Excel conversion
2. [ ] Implement sheet-specific conversion feature
3. [ ] Complete TUI side-by-side diff view

### Phase 2: Excel Feature Parity (Requires investigation)
1. [ ] Investigate current excelize capabilities for style extraction
2. [ ] Implement any newly available style features
3. [ ] Chart extraction (if excelize supports it)
4. [ ] Pivot table extraction (if excelize supports it)

### Phase 3: Advanced Features (Future)
1. [ ] Design and implement plugin architecture
2. [ ] Implement hybrid chunking strategy if performance needs arise

## Notes

- All test skips found are legitimate integration test guards using `testing.Short()`
- No critical functionality is missing - the project is production-ready
- Most TODOs are related to Excel advanced features limited by the excelize library
- Consider contributing to excelize project to enable missing features

## Contributing

When working on any of these items:
1. Check if excelize has added support for the feature
2. Write comprehensive tests
3. Update documentation
4. Follow the existing code patterns and conventions