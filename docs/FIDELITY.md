# Round-Trip Fidelity Contract

GitCells aims to do less, perfectly. This page lists exactly which Excel
features survive a round trip (`.xlsx` → JSON chunks → `.xlsx`) and which
are dropped with a warning at convert time.

The principle: any feature in **Preserved** must be byte-equivalent (or
semantically identical for floating-point) after a round trip. Any
feature in **Not preserved** is documented and warned about at the time
of conversion — never silently lost.

## Preserved

### Sheet structure
- Sheet names (Unicode, spaces, most special characters)
- Sheet index (order)
- Sheet hidden flag
- Merged cell ranges (validated; malformed ranges are skipped with a warning)

### Cells
- Cell value (string, number, boolean, date as serialized, error literals)
- Formula text (A1 style)
- Cell type: `string`, `number`, `boolean`, `date`, `error`, `formula`
- Excel error values: `#REF!`, `#N/A`, `#VALUE!`, `#DIV/0!`, `#NAME?`,
  `#NUM!`, `#NULL!`, `#GETTING_DATA`
- Hyperlinks (external URLs and internal sheet references)
- Cell comments (author + text)
- Data validations with valid `sqref` ranges

### Styling
- Font: name, size, bold, italic, underline, color
- Fill: type, pattern (named or numeric), color, background color
- Border: per-side style and color
- Alignment: horizontal, vertical, wrap-text, text rotation
- Number formats: built-in IDs **and** custom format strings (passed
  through via `CustomNumFmt`)

### Layout
- Column widths
- Row heights

### Workbook
- Defined names
- Document properties: title, subject, author, keywords, description

## Not preserved (warned at convert time)

These features are detected and logged so you know they were skipped,
but they do not survive the round trip:

- **Images and embedded objects** — including charts that rely on images
- **Pivot tables**
- **Conditional formatting**
- **Tables** (ListObjects)
- **VBA macros** (`.xlsm` content beyond cell data)
- **Sheet protection** (password + permission flags)
- **Slicers and sparklines**
- **Rich-text-per-run** cell formatting (each cell's formatting is per-cell, not per-run)
- **Array-formula spill ranges** (the formula text is preserved; the
  spill range is recomputed by Excel on open)
- **Custom XML parts**
- **`Properties.Company`** — excelize cannot write this field back, so
  it is intentionally not extracted

## Path portability

Chunk metadata (`.gitcells_chunks.json`) stores chunk file paths as
**forward slashes regardless of writer OS**. The reader applies
`filepath.FromSlash` before joining, so a chunk tree authored on Linux
reads back correctly on Windows and vice versa. The writer strips
both `.xlsx` and `.json` from the source filename so the chunk
directory name (e.g. `sample_chunks`) is identical to what the reader
computes from the same source path.

Do not break this contract when editing `internal/converter/chunking.go`:

- Writer: normalize relative chunk paths via `filepath.ToSlash` before
  storing in metadata.
- Reader: apply `filepath.FromSlash` before passing each stored path
  to `filepath.Join`.

## Why "do less, perfectly"

A partial implementation of a feature is worse than no implementation:
users assume round-trip works, ship workbooks through git, and discover
the loss months later. We picked features above strictly on the basis
that we can fully round-trip them across Windows, macOS, and Linux.

If a feature you need is in the **Not preserved** list, you have two
options:

1. Convert the source workbook to a stripped form (no charts, no
   pivots) before adding it to git. GitCells will then preserve
   everything that remains.
2. Track the original `.xlsx` directly in git (with LFS) and use
   GitCells only for diff-readable JSON of the cell data.
