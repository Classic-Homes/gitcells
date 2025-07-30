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
}

type Cell struct {
	Value          interface{}     `json:"value"`
	Formula        string          `json:"formula,omitempty"`
	Style          *CellStyle      `json:"style,omitempty"`
	Type           CellType        `json:"type"`
	Comment        *Comment        `json:"comment,omitempty"`
	Hyperlink      string          `json:"hyperlink,omitempty"`
	DataValidation *DataValidation `json:"data_validation,omitempty"`
}

type CellType string

const (
	CellTypeString  CellType = "string"
	CellTypeNumber  CellType = "number"
	CellTypeBoolean CellType = "boolean"
	CellTypeDate    CellType = "date"
	CellTypeError   CellType = "error"
	CellTypeFormula CellType = "formula"
)

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