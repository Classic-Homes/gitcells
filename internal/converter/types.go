package converter

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/xuri/excelize/v2"
)

// detectCellType determines the cell type based on its value and formula
func (c *converter) detectCellType(value interface{}, formula string) models.CellType {
	if value == nil {
		return models.CellTypeString
	}

	// Check for formula first (takes precedence)
	if formula != "" {
		// Check if it's an array formula
		if c.isArrayFormula(formula) {
			return models.CellTypeArrayFormula
		}
		return models.CellTypeFormula
	}

	switch v := value.(type) {
	case string:
		// Check if it's a number stored as string
		if _, err := strconv.ParseFloat(v, 64); err == nil {
			return models.CellTypeNumber
		}
		// Check if it's a boolean
		if strings.EqualFold(v, "true") || strings.EqualFold(v, "false") {
			return models.CellTypeBoolean
		}
		// Check if it starts with = (formula) - backup check
		if strings.HasPrefix(v, "=") {
			return models.CellTypeFormula
		}
		return models.CellTypeString
	case float64, float32, int, int32, int64:
		return models.CellTypeNumber
	case bool:
		return models.CellTypeBoolean
	default:
		return models.CellTypeString
	}
}

// isArrayFormula checks if a formula is an array formula
func (c *converter) isArrayFormula(formula string) bool {
	// Array formulas are typically surrounded by curly braces in Excel
	// or contain specific array patterns that Excel recognizes
	formulaTrimmed := strings.TrimSpace(formula)

	// Check for curly braces (Excel's array formula indicator)
	if strings.HasPrefix(formulaTrimmed, "{") && strings.HasSuffix(formulaTrimmed, "}") {
		return true
	}

	// Check for specific array function patterns
	formulaUpper := strings.ToUpper(formula)

	// Functions that are commonly used in array contexts with multiple ranges
	arrayOnlyFunctions := []string{"TRANSPOSE(", "MMULT(", "SUMPRODUCT("}
	for _, fn := range arrayOnlyFunctions {
		if strings.Contains(formulaUpper, fn) {
			return true
		}
	}

	// More conservative detection for regular functions
	// Only consider it an array formula if it has multiple complex range references
	// or multiple ranges in the same function
	rangeCount := strings.Count(formula, ":")
	commaCount := strings.Count(formula, ",")

	// Array formulas typically have multiple ranges or complex patterns
	if rangeCount > 1 && commaCount > 2 {
		return true
	}

	return false
}

// extractEnhancedFormulaInfo extracts detailed formula information
func (c *converter) extractEnhancedFormulaInfo(f *excelize.File, sheetName, cellRef string) (string, string, *models.ArrayFormula, error) {
	// Get standard A1 formula
	formula, err := f.GetCellFormula(sheetName, cellRef)
	if err != nil {
		return "", "", nil, err
	}

	// R1C1 formula extraction not yet supported by excelize - See issue #4
	formulaR1C1 := ""

	// Check for array formula
	var arrayFormula *models.ArrayFormula
	if formula != "" && c.isArrayFormula(formula) {
		// Array formula range detection limited by excelize - See issue #5
		arrayFormula = &models.ArrayFormula{
			Formula: formula,
			Range:   cellRef, // Placeholder - accurate range detection needs excelize enhancement
			IsCSE:   true,    // Assume CSE for now
		}
	}

	return formula, formulaR1C1, arrayFormula, nil
}

// parseNumber converts a string to a number (float64) if possible
func parseNumber(value string) (float64, error) {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, utils.WrapError(err, utils.ErrorTypeConverter, "parseNumber",
			fmt.Sprintf("failed to parse number from value: %s", value))
	}
	return num, nil
}

// extractFullCellStyle is implemented in style_extraction.go

// formatCommitMessage formats a commit message with variable substitution
func formatCommitMessage(template string, replacements map[string]string) string {
	result := template
	for key, value := range replacements {
		result = strings.ReplaceAll(result, "{"+key+"}", value)
	}
	return result
}

// extractProperties converts excelize document properties to our model
func (c *converter) extractProperties(props interface{}) models.DocumentProperties {
	if props == nil {
		return models.DocumentProperties{}
	}

	// Type assert to excelize.DocProperties
	if docProps, ok := props.(*excelize.DocProperties); ok {
		return models.DocumentProperties{
			Title:       docProps.Title,
			Subject:     docProps.Subject,
			Author:      docProps.Creator,
			Company:     docProps.Category,
			Keywords:    docProps.Keywords,
			Description: docProps.Description,
		}
	}

	return models.DocumentProperties{}
}

// cellReference converts column and row indices to Excel cell reference (e.g., "A1")
func cellReference(col, row int) (string, error) {
	if col < 1 || row < 1 {
		return "", utils.NewError(utils.ErrorTypeValidation, "cellReference",
			fmt.Sprintf("invalid cell coordinates: col=%d, row=%d", col, row))
	}

	// Convert column number to letter(s)
	colName := ""
	for col > 0 {
		col--
		colName = string(rune('A'+col%26)) + colName
		col /= 26
	}

	return fmt.Sprintf("%s%d", colName, row), nil
}

// MemoryStats tracks memory usage during conversion
type MemoryStats struct {
	InitialMemory uint64
	PeakMemory    uint64
	CurrentMemory uint64
	GCCount       uint32
}

// GetMemoryStats returns current memory statistics
func GetMemoryStats() *MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemoryStats{
		CurrentMemory: m.Alloc,
		PeakMemory:    m.TotalAlloc,
		GCCount:       m.NumGC,
	}
}

// ForceGC forces garbage collection and returns memory stats
func ForceGC() *MemoryStats {
	runtime.GC()
	return GetMemoryStats()
}

// StreamingConfig defines configuration for streaming processing
type StreamingConfig struct {
	ChunkSize        int    // Number of rows to process at once
	MemoryThreshold  uint64 // Memory threshold in bytes to trigger GC
	ProgressCallback func(processed, total int)
}

// DefaultStreamingConfig returns sensible defaults for streaming
func DefaultStreamingConfig() *StreamingConfig {
	return &StreamingConfig{
		ChunkSize:       1000,      // Process 1000 rows at a time
		MemoryThreshold: 100 << 20, // 100MB threshold
	}
}

// ProgressBar represents a simple console progress bar
type ProgressBar struct {
	current int
	total   int
	width   int
	stage   string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		current: 0,
		total:   total,
		width:   50, // Default width
	}
}

// Update updates the progress bar display
func (pb *ProgressBar) Update(stage string, current, total int) {
	pb.current = current
	pb.total = total
	pb.stage = stage

	// Calculate percentage
	percentage := float64(current) / float64(total) * 100
	if total == 0 {
		percentage = 100
	}

	// Calculate filled portion of bar
	filled := int(float64(current) / float64(total) * float64(pb.width))
	if filled > pb.width {
		filled = pb.width
	}

	// Build progress bar string
	bar := "["
	for i := 0; i < pb.width; i++ {
		switch {
		case i < filled:
			bar += "="
		case i == filled && current < total:
			bar += ">"
		default:
			bar += " "
		}
	}
	bar += "]"

	// Print progress (use \r to overwrite previous line)
	fmt.Printf("\r%s %s %.1f%% (%d/%d)", stage, bar, percentage, current, total)

	// Print newline when complete
	if current >= total {
		fmt.Println()
	}
}

// DefaultProgressCallback returns a default progress callback that shows a progress bar
func DefaultProgressCallback() func(string, int, int) {
	const progressBarMax = 100
	pb := NewProgressBar(progressBarMax)
	return pb.Update
}

// extractCharts extracts chart information from Excel sheet
func (c *converter) extractCharts(f interface{}, sheetName string) ([]models.Chart, error) {
	charts := []models.Chart{}

	// Since excelize doesn't support reading existing charts, we need to access
	// the underlying Excel file structure manually via ZIP file inspection
	excelFile, ok := f.(*excelize.File)
	if !ok {
		c.logger.Debug("Chart extraction requires excelize.File object")
		return charts, nil
	}

	// Get the file path from the excelize File object's internal state
	// Unfortunately, excelize doesn't expose the original file path directly,
	// so we'll need to work with what's available in the excelize File object

	// For now, we can extract chart information from the internal XML data
	// that excelize has already parsed, though this is limited
	charts, err := c.extractChartsFromExcelFile(excelFile, sheetName)
	if err != nil {
		c.logger.Debugf("Chart extraction encountered error: %v", err)
		// Return empty charts rather than error to maintain compatibility
		return []models.Chart{}, nil
	}

	c.logger.Debugf("Successfully extracted %d charts from sheet %s", len(charts), sheetName)
	return charts, nil
}

// extractChartsFromExcelFile extracts chart information from excelize File object
func (c *converter) extractChartsFromExcelFile(f *excelize.File, sheetName string) ([]models.Chart, error) {
	c.logger.Debugf("Attempting chart extraction for sheet: %s", sheetName)

	// Verify sheet exists
	if !c.sheetExists(f, sheetName) {
		c.logger.Debugf("Sheet %s not found in workbook", sheetName)
		return []models.Chart{}, nil
	}

	// Detect charts through pictures and data patterns
	hasPictures := c.detectPictures(f, sheetName)
	chartData := c.analyzeChartDataPatterns(f, sheetName)

	// Build charts from detected patterns
	charts := c.buildChartsFromPatterns(sheetName, chartData)

	// Add fallback chart if pictures detected but no data patterns
	if hasPictures && len(chartData) == 0 {
		charts = append(charts, c.createFallbackChart(sheetName))
	}

	if len(charts) > 0 {
		c.logger.Debugf("Detected %d potential charts in sheet %s", len(charts), sheetName)
	}

	return charts, nil
}

// sheetExists checks if a sheet exists in the workbook
func (c *converter) sheetExists(f *excelize.File, sheetName string) bool {
	sheetMap := f.GetSheetMap()
	for _, name := range sheetMap {
		if name == sheetName {
			return true
		}
	}
	return false
}

// detectPictures checks if there are any pictures in the sheet (potential charts)
func (c *converter) detectPictures(f *excelize.File, sheetName string) bool {
	// Check if there are any pictures or drawings (charts might be stored as drawings)
	pics, _ := f.GetPictures(sheetName, "A1") // This will return empty for non-picture cells
	return len(pics) > 0
}

// buildChartsFromPatterns creates chart models from detected data patterns
func (c *converter) buildChartsFromPatterns(sheetName string, patterns []ChartDataPattern) []models.Chart {
	charts := []models.Chart{}

	for i, pattern := range patterns {
		chart := models.Chart{
			ID:    fmt.Sprintf("chart_%s_%d", sheetName, i+1),
			Type:  c.inferChartType(pattern),
			Title: fmt.Sprintf("Chart %d in %s", i+1, sheetName),
			Position: models.ChartPosition{
				X:      float64(i * 100),
				Y:      0,
				Width:  400,
				Height: 300,
			},
			Series: c.createChartSeries(pattern),
		}
		charts = append(charts, chart)
	}

	return charts
}

// createFallbackChart creates a generic chart when pictures are detected but no data patterns
func (c *converter) createFallbackChart(sheetName string) models.Chart {
	return models.Chart{
		ID:    fmt.Sprintf("chart_%s_1", sheetName),
		Type:  "unknown",
		Title: fmt.Sprintf("Chart detected in %s", sheetName),
		Position: models.ChartPosition{
			X:      0,
			Y:      0,
			Width:  400,
			Height: 300,
		},
		Series: []models.ChartSeries{},
	}
}

// isNumeric checks if a string represents a numeric value
func (c *converter) isNumeric(s string) bool {
	if s == "" {
		return false
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// ChartDataPattern represents a detected data pattern that might be used for charts
type ChartDataPattern struct {
	HeaderRange string   // Range containing headers (e.g., "A1:D1")
	DataRange   string   // Range containing data (e.g., "A2:D10")
	Headers     []string // Header values
	NumericCols int      // Number of numeric columns
	DataRows    int      // Number of data rows
}

// analyzeChartDataPatterns analyzes the sheet for data patterns that might indicate charts
func (c *converter) analyzeChartDataPatterns(f *excelize.File, sheetName string) []ChartDataPattern {
	patterns := []ChartDataPattern{}

	// Get sheet data
	rows, err := f.GetRows(sheetName)
	if err != nil || !c.hasMinimumDataForChart(rows) {
		return patterns
	}

	// Find maximum columns
	maxCols := c.findMaxColumns(rows)
	if maxCols < 2 {
		return patterns
	}

	// Analyze first row as potential headers
	headers := rows[0]
	if len(headers) >= 2 {
		pattern := c.analyzeTabularData(f, sheetName, rows, headers)
		if pattern != nil {
			patterns = append(patterns, *pattern)
		}
	}

	return patterns
}

// hasMinimumDataForChart checks if there's enough data for a chart
func (c *converter) hasMinimumDataForChart(rows [][]string) bool {
	return len(rows) >= 2
}

// findMaxColumns finds the maximum number of columns across all rows
func (c *converter) findMaxColumns(rows [][]string) int {
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	return maxCols
}

// analyzeTabularData analyzes tabular data for chart patterns
func (c *converter) analyzeTabularData(f *excelize.File, sheetName string, rows [][]string, headers []string) *ChartDataPattern {
	stats := c.calculateDataStats(f, sheetName, rows, headers)

	// Need at least 2 numeric columns and 2 data rows for a valid chart pattern
	if stats.numericCols < 2 || stats.dataRows < 2 {
		return nil
	}

	return c.createChartPattern(headers, stats)
}

// dataStats holds statistics about data columns
type dataStats struct {
	numericCols int
	dataRows    int
}

// calculateDataStats calculates statistics about numeric columns and data rows
func (c *converter) calculateDataStats(f *excelize.File, sheetName string, rows [][]string, headers []string) dataStats {
	stats := dataStats{}

	// Analyze up to 10 columns and 20 rows for performance
	maxCols := len(headers)
	if maxCols > 10 {
		maxCols = 10
	}

	maxRows := len(rows)
	if maxRows > 20 {
		maxRows = 20
	}

	for colIndex := 0; colIndex < maxCols; colIndex++ {
		if c.isNumericColumn(f, sheetName, rows, colIndex, maxRows, &stats.dataRows) {
			stats.numericCols++
		}
	}

	return stats
}

// isNumericColumn checks if a column contains numeric data
func (c *converter) isNumericColumn(f *excelize.File, sheetName string, rows [][]string, colIndex, maxRows int, dataRows *int) bool {
	hasNumeric := false

	for rowIndex := 1; rowIndex < maxRows; rowIndex++ {
		if colIndex < len(rows[rowIndex]) {
			cellName, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			cellValue, _ := f.GetCellValue(sheetName, cellName)
			if c.isNumeric(cellValue) && cellValue != "" {
				hasNumeric = true
				if rowIndex > *dataRows {
					*dataRows = rowIndex
				}
			}
		}
	}

	return hasNumeric
}

// createChartPattern creates a ChartDataPattern from headers and statistics
func (c *converter) createChartPattern(headers []string, stats dataStats) *ChartDataPattern {
	headerEndCol, _ := excelize.CoordinatesToCellName(len(headers), 1)
	dataEndCol, _ := excelize.CoordinatesToCellName(len(headers), stats.dataRows)

	return &ChartDataPattern{
		HeaderRange: fmt.Sprintf("A1:%s", headerEndCol),
		DataRange:   fmt.Sprintf("A2:%s", dataEndCol),
		Headers:     headers,
		NumericCols: stats.numericCols,
		DataRows:    stats.dataRows - 1, // Exclude header row
	}
}

// inferChartType tries to infer the chart type based on data patterns
func (c *converter) inferChartType(pattern ChartDataPattern) string {
	// Simple heuristics for chart type inference
	switch {
	case pattern.NumericCols == 1:
		return "pie" // Single data series might be pie chart
	case pattern.NumericCols >= 2 && pattern.DataRows > 10:
		return "line" // Time series data might be line chart
	case pattern.NumericCols >= 2:
		return "column" // Multiple series might be column chart
	default:
		return "column" // Default to column chart
	}
}

// createChartSeries creates chart series from detected data patterns
func (c *converter) createChartSeries(pattern ChartDataPattern) []models.ChartSeries {
	series := []models.ChartSeries{}

	// Create series based on headers and data ranges
	for i, header := range pattern.Headers {
		if i == 0 {
			continue // First column is usually categories/labels
		}

		// Create range references for the series
		startCol, _ := excelize.CoordinatesToCellName(i+1, 2) // Data starts at row 2
		endCol, _ := excelize.CoordinatesToCellName(i+1, pattern.DataRows+1)

		categoriesStart, _ := excelize.CoordinatesToCellName(1, 2)
		categoriesEnd, _ := excelize.CoordinatesToCellName(1, pattern.DataRows+1)

		chartSeries := models.ChartSeries{
			Name:       header,
			Categories: fmt.Sprintf("%s:%s", categoriesStart, categoriesEnd),
			Values:     fmt.Sprintf("%s:%s", startCol, endCol),
		}
		series = append(series, chartSeries)
	}

	return series
}

// extractPivotTables extracts pivot table information from Excel sheet
func (c *converter) extractPivotTables(f interface{}, sheetName string) ([]models.PivotTable, error) {
	excelFile, ok := f.(*excelize.File)
	if !ok {
		c.logger.Debug("Pivot table extraction requires excelize.File object")
		return []models.PivotTable{}, nil
	}

	pivotTables := []models.PivotTable{}

	// Use excelize's GetPivotTables method to extract pivot table information
	excelPivotTables, err := excelFile.GetPivotTables(sheetName)
	if err != nil {
		c.logger.Debugf("Failed to get pivot tables from sheet %s: %v", sheetName, err)
		return pivotTables, nil
	}

	c.logger.Debugf("Found %d pivot tables in sheet %s", len(excelPivotTables), sheetName)

	// Convert excelize pivot table format to our internal format
	for i, excelPivot := range excelPivotTables {
		pivotTable := models.PivotTable{
			ID:          fmt.Sprintf("pivot_%s_%d", sheetName, i+1),
			Name:        excelPivot.Name,
			SourceRange: excelPivot.DataRange,
			TargetRange: excelPivot.PivotTableRange,
		}

		// Convert row fields
		for _, field := range excelPivot.Rows {
			pivotField := models.PivotField{
				Name:     field.Data,
				Position: len(pivotTable.RowFields),
				Subtotal: field.DefaultSubtotal,
			}
			pivotTable.RowFields = append(pivotTable.RowFields, pivotField)
		}

		// Convert column fields
		for _, field := range excelPivot.Columns {
			pivotField := models.PivotField{
				Name:     field.Data,
				Position: len(pivotTable.ColumnFields),
				Subtotal: field.DefaultSubtotal,
			}
			pivotTable.ColumnFields = append(pivotTable.ColumnFields, pivotField)
		}

		// Convert data fields (value fields)
		for _, field := range excelPivot.Data {
			dataField := models.PivotDataField{
				Name:         field.Data,
				Function:     c.convertPivotFunction(field.Subtotal),
				DisplayName:  field.Name,
				NumberFormat: fmt.Sprintf("%d", field.NumFmt),
			}
			pivotTable.DataFields = append(pivotTable.DataFields, dataField)
		}

		// Convert filter fields (page fields)
		for _, field := range excelPivot.Filter {
			pivotField := models.PivotField{
				Name:     field.Data,
				Position: len(pivotTable.FilterFields),
				Subtotal: field.DefaultSubtotal,
			}
			pivotTable.FilterFields = append(pivotTable.FilterFields, pivotField)
		}

		// Set pivot table settings
		pivotTable.Settings = &models.PivotTableSettings{
			ShowGrandTotals:   excelPivot.RowGrandTotals || excelPivot.ColGrandTotals,
			ShowRowHeaders:    excelPivot.ShowRowHeaders,
			ShowColumnHeaders: excelPivot.ShowColHeaders,
			CompactForm:       !excelPivot.ClassicLayout,
			OutlineForm:       excelPivot.ClassicLayout,
			TabularForm:       false, // Default to false as excelize doesn't directly expose this
		}

		pivotTables = append(pivotTables, pivotTable)
	}

	c.logger.Debugf("Successfully extracted %d pivot tables from sheet %s", len(pivotTables), sheetName)
	return pivotTables, nil
}

// convertPivotFunction converts excelize pivot function to our standard format
func (c *converter) convertPivotFunction(subtotal string) string {
	switch subtotal {
	case "Sum":
		return "SUM"
	case "Count":
		return "COUNT"
	case "Average":
		return "AVERAGE"
	case "Max":
		return "MAX"
	case "Min":
		return "MIN"
	case "Product":
		return "PRODUCT"
	case "CountNums":
		return "COUNTA"
	case "StdDev":
		return "STDEV"
	case "StdDevp":
		return "STDEVP"
	case "Var":
		return "VAR"
	case "Varp":
		return "VARP"
	default:
		c.logger.Debugf("Unknown pivot function: %s, defaulting to SUM", subtotal)
		return "SUM"
	}
}
