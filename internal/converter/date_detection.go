package converter

import (
	"github.com/xuri/excelize/v2"
)

// Excel stores dates as floating-point serial numbers (days since the 1900
// epoch, with the 1900 leap-year bug baked in). Whether a cell *displays*
// as a date is purely a function of its number-format style. The
// underlying value is just a float, indistinguishable from any other
// number until the style is consulted.
//
// Pre-fix, the converter read every cell as its formatted display string
// via GetRows, so a date cell came back as "1/1/2020" — which then failed
// to parse as a number and got stored as CellTypeString. After a
// JSON→Excel round-trip, the cell was a plain text string and any formula
// like =MONTH(A1) returned #VALUE!. The fix is two-part:
//
//   1. Detect date-formatted cells via their style number format. The
//      built-in date format IDs are 14–22 and 45–47; custom formats are
//      identified by the presence of date/time tokens (y/m/d/h/s) outside
//      literal quoted sections.
//   2. For detected date cells, swap the formatted display string for the
//      raw serial value (via excelize.Options{RawCellValue: true}) and
//      tag the cell as CellTypeDate. On write, SetCellValue with the
//      float serial plus the round-tripped style format renders identical
//      output to the source workbook.

// isDateFormat reports whether the given Excel number-format descriptor
// indicates a date- or time-formatted cell.
func isDateFormat(numFmtID int, customFmt string) bool {
	if (numFmtID >= 14 && numFmtID <= 22) || (numFmtID >= 45 && numFmtID <= 47) {
		return true
	}
	if customFmt == "" {
		return false
	}
	return containsDateToken(customFmt)
}

// containsDateToken scans a custom Excel number-format string for the
// y/m/d/h/s tokens that mark it as date- or time-related. Tokens inside
// quoted literals (`"..."`), inside square-bracket sections (`[...]`, used
// for color/locale/condition prefixes), or preceded by a backslash escape
// don't count. The token character is case-insensitive.
//
// Note: `m` is intentionally treated as a date token. In Excel format
// syntax, `m` outside quoted literals is always month or minute — there
// is no general-purpose use of `m` in a numeric format. A custom format
// like `0"meters"` keeps the `m` in quotes and is not flagged.
func containsDateToken(fmt string) bool {
	inQuote := false
	inBracket := false
	for i := 0; i < len(fmt); i++ {
		ch := fmt[i]
		switch ch {
		case '"':
			inQuote = !inQuote
			continue
		case '[':
			inBracket = true
			continue
		case ']':
			inBracket = false
			continue
		case '\\':
			i++ // skip the escaped next byte
			continue
		}
		if inQuote || inBracket {
			continue
		}
		lower := ch | 0x20
		if lower == 'y' || lower == 'm' || lower == 'd' || lower == 'h' || lower == 's' {
			return true
		}
	}
	return false
}

// dateFormatCache memoizes per-styleID date-format checks so a dense
// sheet doesn't repeat the GetStyle lookup for every cell. Most
// workbooks have far fewer unique styles than cells, so the cache hit
// rate is high.
type dateFormatCache map[int]bool

// isDateCell looks up the style for the given cell ref and reports
// whether its number format renders as a date or time. Errors from the
// excelize lookups are treated as "not a date" — date detection is a
// best-effort upgrade, not a hard requirement, so a style read failure
// degrades to the prior string-storage behavior rather than failing the
// whole conversion.
func isDateCell(f *excelize.File, sheetName, cellRef string, cache dateFormatCache) bool {
	styleID, err := f.GetCellStyle(sheetName, cellRef)
	if err != nil || styleID == 0 {
		return false
	}
	if v, ok := cache[styleID]; ok {
		return v
	}
	styl, err := f.GetStyle(styleID)
	if err != nil || styl == nil {
		cache[styleID] = false
		return false
	}
	customFmt := ""
	if styl.CustomNumFmt != nil {
		customFmt = *styl.CustomNumFmt
	}
	v := isDateFormat(styl.NumFmt, customFmt)
	cache[styleID] = v
	return v
}
