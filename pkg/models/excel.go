package models

import (
	"time"
)

// ExcelDocument represents the complete Excel file structure
type ExcelDocument struct {
	Version      string             `json:"version"`
	Metadata     DocumentMetadata   `json:"metadata"`
	Sheets       []Sheet            `json:"sheets"`
	DefinedNames map[string]string  `json:"defined_names,omitempty"`
	Properties   DocumentProperties `json:"properties,omitempty"`
}

type DocumentMetadata struct {
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
	AppVersion   string    `json:"app_version"`
	OriginalFile string    `json:"original_file"`
	FileSize     int64     `json:"file_size"`
	Checksum     string    `json:"checksum"` // SHA256 of original
}

type DocumentProperties struct {
	Title       string `json:"title,omitempty"`
	Subject     string `json:"subject,omitempty"`
	Author      string `json:"author,omitempty"`
	Company     string `json:"company,omitempty"`
	Keywords    string `json:"keywords,omitempty"`
	Description string `json:"description,omitempty"`
}

type Sheet struct {
	Name               string                 `json:"name"`
	Index              int                    `json:"index"`
	Cells              map[string]Cell        `json:"cells"`
	MergedCells        []MergedCell           `json:"merged_cells,omitempty"`
	RowHeights         map[int]float64        `json:"row_heights,omitempty"`
	ColumnWidths       map[string]float64     `json:"column_widths,omitempty"`
	Hidden             bool                   `json:"hidden"`
	Protection         *SheetProtection       `json:"protection,omitempty"`
	ConditionalFormats []ConditionalFormat    `json:"conditional_formats,omitempty"`
	Charts             []Chart                `json:"charts,omitempty"`
	PivotTables        []PivotTable           `json:"pivot_tables,omitempty"`
}

type Cell struct {
	Value          interface{}     `json:"value"`
	Formula        string          `json:"formula,omitempty"`
	FormulaR1C1    string          `json:"formula_r1c1,omitempty"`    // R1C1 reference style
	ArrayFormula   *ArrayFormula   `json:"array_formula,omitempty"`   // Array formula details
	Style          *CellStyle      `json:"style,omitempty"`
	Type           CellType        `json:"type"`
	Comment        *Comment        `json:"comment,omitempty"`
	Hyperlink      string          `json:"hyperlink,omitempty"`
	DataValidation *DataValidation `json:"data_validation,omitempty"`
}

type CellType string

const (
	CellTypeString      CellType = "string"
	CellTypeNumber      CellType = "number"
	CellTypeBoolean     CellType = "boolean"
	CellTypeDate        CellType = "date"
	CellTypeError       CellType = "error"
	CellTypeFormula     CellType = "formula"
	CellTypeArrayFormula CellType = "array_formula"
)

// ArrayFormula represents an Excel array formula
type ArrayFormula struct {
	Formula string `json:"formula"`
	Range   string `json:"range"` // The range the array formula covers
	IsCSE   bool   `json:"is_cse"` // Ctrl+Shift+Enter array formula
}

// MergedCell represents a compact representation for merged cells
type MergedCell struct {
	Range string `json:"range"` // e.g., "A1:C3"
}

type CellStyle struct {
	Font         *Font         `json:"font,omitempty"`
	Fill         *Fill         `json:"fill,omitempty"`
	Border       *Border       `json:"border,omitempty"`
	NumberFormat string        `json:"number_format,omitempty"`
	Alignment    *Alignment    `json:"alignment,omitempty"`
}

type Font struct {
	Name      string  `json:"name,omitempty"`
	Size      float64 `json:"size,omitempty"`
	Bold      bool    `json:"bold,omitempty"`
	Italic    bool    `json:"italic,omitempty"`
	Underline string  `json:"underline,omitempty"`
	Color     string  `json:"color,omitempty"`
}

type Fill struct {
	Type    string `json:"type,omitempty"`
	Pattern string `json:"pattern,omitempty"`
	Color   string `json:"color,omitempty"`
	BgColor string `json:"bg_color,omitempty"`
}

type Border struct {
	Left   *BorderLine `json:"left,omitempty"`
	Right  *BorderLine `json:"right,omitempty"`
	Top    *BorderLine `json:"top,omitempty"`
	Bottom *BorderLine `json:"bottom,omitempty"`
}

type BorderLine struct {
	Style string `json:"style,omitempty"`
	Color string `json:"color,omitempty"`
}

type Alignment struct {
	Horizontal   string `json:"horizontal,omitempty"`
	Vertical     string `json:"vertical,omitempty"`
	WrapText     bool   `json:"wrap_text,omitempty"`
	TextRotation int    `json:"text_rotation,omitempty"`
}

type Comment struct {
	Author string `json:"author,omitempty"`
	Text   string `json:"text"`
}

type DataValidation struct {
	Type             string      `json:"type"`
	Operator         string      `json:"operator,omitempty"`
	Formula1         string      `json:"formula1,omitempty"`
	Formula2         string      `json:"formula2,omitempty"`
	AllowBlank       bool        `json:"allow_blank,omitempty"`
	ShowInputMessage bool        `json:"show_input_message,omitempty"`
	ShowErrorMessage bool        `json:"show_error_message,omitempty"`
	ErrorTitle       string      `json:"error_title,omitempty"`
	Error            string      `json:"error,omitempty"`
	PromptTitle      string      `json:"prompt_title,omitempty"`
	Prompt           string      `json:"prompt,omitempty"`
}

type SheetProtection struct {
	Password               string `json:"password,omitempty"`
	EditObjects            bool   `json:"edit_objects,omitempty"`
	EditScenarios          bool   `json:"edit_scenarios,omitempty"`
	FormatCells            bool   `json:"format_cells,omitempty"`
	FormatColumns          bool   `json:"format_columns,omitempty"`
	FormatRows             bool   `json:"format_rows,omitempty"`
	InsertColumns          bool   `json:"insert_columns,omitempty"`
	InsertRows             bool   `json:"insert_rows,omitempty"`
	InsertHyperlinks       bool   `json:"insert_hyperlinks,omitempty"`
	DeleteColumns          bool   `json:"delete_columns,omitempty"`
	DeleteRows             bool   `json:"delete_rows,omitempty"`
	SelectLockedCells      bool   `json:"select_locked_cells,omitempty"`
	SelectUnlockedCells    bool   `json:"select_unlocked_cells,omitempty"`
	Sort                   bool   `json:"sort,omitempty"`
	AutoFilter             bool   `json:"auto_filter,omitempty"`
	PivotTables            bool   `json:"pivot_tables,omitempty"`
}

type ConditionalFormat struct {
	Range     string                   `json:"range"`
	Type      string                   `json:"type"`
	Criteria  string                   `json:"criteria,omitempty"`
	Value     interface{}              `json:"value,omitempty"`
	Minimum   interface{}              `json:"minimum,omitempty"`
	Maximum   interface{}              `json:"maximum,omitempty"`
	Format    *CellStyle               `json:"format,omitempty"`
}

// Chart represents an Excel chart object
type Chart struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"` // bar, line, pie, scatter, etc.
	Title      string          `json:"title,omitempty"`
	Position   ChartPosition   `json:"position"`
	Series     []ChartSeries   `json:"series"`
	Legend     *ChartLegend    `json:"legend,omitempty"`
	Axes       *ChartAxes      `json:"axes,omitempty"`
	Style      *ChartStyle     `json:"style,omitempty"`
}

type ChartPosition struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type ChartSeries struct {
	Name       string `json:"name,omitempty"`
	Categories string `json:"categories,omitempty"` // Range reference like "Sheet1!A1:A10"
	Values     string `json:"values,omitempty"`     // Range reference like "Sheet1!B1:B10"
	Color      string `json:"color,omitempty"`
}

type ChartLegend struct {
	Position string `json:"position"` // top, bottom, left, right, none
	Show     bool   `json:"show"`
}

type ChartAxes struct {
	XAxis *ChartAxis `json:"x_axis,omitempty"`
	YAxis *ChartAxis `json:"y_axis,omitempty"`
}

type ChartAxis struct {
	Title      string  `json:"title,omitempty"`
	Min        *float64 `json:"min,omitempty"`
	Max        *float64 `json:"max,omitempty"`
	MajorUnit  *float64 `json:"major_unit,omitempty"`
	MinorUnit  *float64 `json:"minor_unit,omitempty"`
}

type ChartStyle struct {
	ColorScheme string `json:"color_scheme,omitempty"`
	PlotArea    string `json:"plot_area,omitempty"`
}

// PivotTable represents an Excel pivot table
type PivotTable struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	SourceRange   string              `json:"source_range"`   // Range reference like "Sheet1!A1:D100"
	TargetRange   string              `json:"target_range"`   // Where pivot table is placed
	RowFields     []PivotField        `json:"row_fields,omitempty"`
	ColumnFields  []PivotField        `json:"column_fields,omitempty"`
	DataFields    []PivotDataField    `json:"data_fields,omitempty"`
	FilterFields  []PivotField        `json:"filter_fields,omitempty"`
	Settings      *PivotTableSettings `json:"settings,omitempty"`
}

type PivotField struct {
	Name     string `json:"name"`
	Position int    `json:"position"`
	Subtotal bool   `json:"subtotal,omitempty"`
}

type PivotDataField struct {
	Name        string `json:"name"`
	Function    string `json:"function"` // SUM, COUNT, AVERAGE, etc.
	DisplayName string `json:"display_name,omitempty"`
	NumberFormat string `json:"number_format,omitempty"`
}

type PivotTableSettings struct {
	ShowGrandTotals   bool `json:"show_grand_totals"`
	ShowRowHeaders    bool `json:"show_row_headers"`
	ShowColumnHeaders bool `json:"show_column_headers"`
	CompactForm       bool `json:"compact_form"`
	OutlineForm       bool `json:"outline_form"`
	TabularForm       bool `json:"tabular_form"`
}