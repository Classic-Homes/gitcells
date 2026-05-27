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

// detectCellType determines the cell type based on its value and formula.
// Array formulas are detected by excelize's surrounding-brace convention
// but treated as ordinary CellTypeFormula — GitCells deliberately does
// not preserve the array-spill range (the previous reflection-based
// implementation was fragile against excelize upgrades).
// excelErrorValues is the set of Excel error markers — when a cell's
// displayed value is one of these strings (and it's not a formula
// result we want to recompute), it's an error literal that should
// round-trip as CellTypeError instead of being misclassified as text.
var excelErrorValues = map[string]struct{}{
	"#REF!":         {},
	"#N/A":          {},
	"#VALUE!":       {},
	"#DIV/0!":       {},
	"#NAME?":        {},
	"#NUM!":         {},
	"#NULL!":        {},
	"#GETTING_DATA": {},
}

func (c *converter) detectCellType(value interface{}, formula string) models.CellType {
	if value == nil && formula == "" {
		return models.CellTypeString
	}

	if formula != "" {
		return models.CellTypeFormula
	}

	switch v := value.(type) {
	case string:
		if _, isErr := excelErrorValues[v]; isErr {
			return models.CellTypeError
		}
		if _, err := strconv.ParseFloat(v, 64); err == nil {
			return models.CellTypeNumber
		}
		if strings.EqualFold(v, "true") || strings.EqualFold(v, "false") {
			return models.CellTypeBoolean
		}
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

func parseNumber(value string) (float64, error) {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, utils.WrapError(err, utils.ErrorTypeConverter, "parseNumber",
			fmt.Sprintf("failed to parse number from value: %s", value))
	}
	return num, nil
}

// extractProperties converts excelize document properties to our model.
func (c *converter) extractProperties(props interface{}) models.DocumentProperties {
	if props == nil {
		return models.DocumentProperties{}
	}
	if docProps, ok := props.(*excelize.DocProperties); ok {
		return models.DocumentProperties{
			Title:       docProps.Title,
			Subject:     docProps.Subject,
			Author:      docProps.Creator,
			Keywords:    docProps.Keywords,
			Description: docProps.Description,
		}
	}
	return models.DocumentProperties{}
}

// cellReference converts column and row indices to Excel cell reference (e.g., "A1").
func cellReference(col, row int) (string, error) {
	if col < 1 || row < 1 {
		return "", utils.NewError(utils.ErrorTypeValidation, "cellReference",
			fmt.Sprintf("invalid cell coordinates: col=%d, row=%d", col, row))
	}
	colName := ""
	for col > 0 {
		col--
		colName = string(rune('A'+col%26)) + colName
		col /= 26
	}
	return fmt.Sprintf("%s%d", colName, row), nil
}

// MemoryStats tracks memory usage during conversion.
type MemoryStats struct {
	InitialMemory uint64
	PeakMemory    uint64
	CurrentMemory uint64
	GCCount       uint32
}

func GetMemoryStats() *MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return &MemoryStats{
		CurrentMemory: m.Alloc,
		PeakMemory:    m.TotalAlloc,
		GCCount:       m.NumGC,
	}
}

func ForceGC() *MemoryStats {
	runtime.GC()
	return GetMemoryStats()
}

// StreamingConfig defines configuration for streaming processing.
type StreamingConfig struct {
	ChunkSize        int    // Number of rows to process at once
	MemoryThreshold  uint64 // Memory threshold in bytes to trigger GC
	ProgressCallback func(processed, total int)
}

func DefaultStreamingConfig() *StreamingConfig {
	return &StreamingConfig{
		ChunkSize:       1000,
		MemoryThreshold: 100 << 20, // 100MB
	}
}

// ProgressBar represents a simple console progress bar.
type ProgressBar struct {
	current int
	total   int
	width   int
	stage   string
}

func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{total: total, width: 50}
}

func (pb *ProgressBar) Update(stage string, current, total int) {
	pb.current = current
	pb.total = total
	pb.stage = stage

	percentage := float64(current) / float64(total) * 100
	if total == 0 {
		percentage = 100
	}

	filled := int(float64(current) / float64(total) * float64(pb.width))
	if filled > pb.width {
		filled = pb.width
	}

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

	fmt.Printf("\r%s %s %.1f%% (%d/%d)", stage, bar, percentage, current, total)
	if current >= total {
		fmt.Println()
	}
}

func DefaultProgressCallback() func(string, int, int) {
	const progressBarMax = 100
	pb := NewProgressBar(progressBarMax)
	return pb.Update
}
