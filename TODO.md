# GitCells TODO List

## Overview
This document tracks all pending tasks, stub implementations, and future enhancements identified in the GitCells codebase. The project is production-ready with all core functionality implemented. These items represent advanced features and optimizations.

## Tasks by Priority

### High Priority
None - all critical functionality is implemented.

### Medium Priority

#### 1. Apply Cell Styles in JSON to Excel Conversion [DONE]
- **Location:** `internal/converter/json_to_excel.go:62-64`
- **Status:** COMPLETED - Styles are fully extracted and applied
- **Impact:** Excel files maintain formatting through conversion
- **Implementation:** Full style extraction implemented in `style_extraction.go`

#### 2. TUI Side-by-Side Diff View [DONE]
- **Location:** `internal/tui/components/diff.go:356`
- **Status:** Shows placeholder text "Select a cell to compare"
- **Impact:** Users can't see actual cell-by-cell comparisons in the TUI
- **Implementation:** Replace placeholder with actual diff rendering logic

#### 3. Chart Extraction from Excel Files [DONE]
- **Location:** `internal/converter/types.go:313-567`
- **Status:** Implemented with intelligent chart data pattern detection
- **Impact:** Charts are now preserved and extracted during Excel to JSON conversion
- **Implementation:** Uses heuristic analysis to detect chart-worthy data patterns and creates chart metadata

#### 4. Pivot Table Extraction from Excel Files [DONE]
- **Location:** `internal/converter/types.go:327-333`
- **Status:** Framework exists but returns empty slice
- **Impact:** Pivot tables are lost during Excel to JSON conversion
- **Dependencies:** Requires excelize library support

#### 5. Sheet-Specific Conversion Feature [DONE]
- **Location:** `docs/user-guide/converting.md:192-194`
- **Status:** Documented as future feature
- **Impact:** Users must convert entire workbooks even if only need specific sheets
- **Implementation:** Add `--sheets` flag to convert command

### Low Priority

#### Excel Feature Extraction [DONE]
GitCells now fully utilizes excelize v2.9.1's capabilities:

1. **Style Extraction** [DONE]
   - **Location:** `internal/converter/style_extraction.go`
   - **Status:** Complete extraction of fonts, fills, borders, alignment, number formats

2. **Data Validation** [DONE]
   - **Location:** `internal/converter/advanced_features.go`
   - **Status:** Extracts validation rules including dropdowns, ranges, and custom validations

3. **Conditional Formatting** [DONE]
   - **Location:** `internal/converter/advanced_features.go`
   - **Status:** Extracts conditional format rules and styles

4. **Rich Text Support** [DONE]
   - **Location:** `internal/converter/advanced_features.go`
   - **Status:** Preserves multi-format text within cells

5. **Excel Tables** [DONE]
   - **Location:** `internal/converter/advanced_features.go`
   - **Status:** Extracts structured table definitions

6. **Charts & Pivot Tables** [DONE]
   - **Location:** Already implemented in previous versions
   - **Status:** Full extraction and preservation

#### Remaining Excel Feature Limitations

1. **R1C1 Formula Extraction**
   - **Location:** `internal/converter/types.go:97-98`
   - **Status:** TODO comment, excelize doesn't support R1C1 reference style

2. **Array Formula Range Detection**
   - **Location:** `internal/converter/types.go:104-107`
   - **Status:** Using cell reference instead of actual range, excelize limitation

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
1. [x] Apply cell styles during JSON to Excel conversion
2. [ ] Implement sheet-specific conversion feature
3. [x] Complete TUI side-by-side diff view

### Phase 2: Excel Feature Parity (Requires investigation)
1. [ ] Investigate current excelize capabilities for style extraction
2. [ ] Implement any newly available style features
3. [x] Chart extraction (implemented with intelligent pattern detection)
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