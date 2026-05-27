# Manual Smoke Test

End-to-end checklist run before each release. Targets the core
Excel→Git pipeline on Windows, macOS, and Linux. Roughly 10 minutes.

If any step fails, capture the output and open an issue tagged
`reliability` before shipping.

## Prep

1. Build the binary for the target OS:
   ```
   make build
   ```
2. Create a scratch directory:
   ```
   mkdir -p /tmp/gitcells-smoke && cd /tmp/gitcells-smoke
   ```
3. Place a real-world `.xlsx` you trust as `sample.xlsx` (any
   workbook with formulas, merged cells, comments, and at least one
   non-Latin sheet name).

## Steps

### 1. Init

```
gitcells init
```

Expect:
- `.gitcells.yaml`, `.gitignore`, `.gitattributes` created.
- `.gitattributes` contains `.gitcells/data/** text eol=lf`.
- Initial commit lands with all three files.

### 2. First convert

```
gitcells convert sample.xlsx
```

Expect:
- `.gitcells/data/sample_chunks/` exists.
- Inside: `workbook.json`, one `sheet_*.json` per sheet,
  `.gitcells_chunks.json`.
- No `.tmp.*` or `.old.*` siblings under `.gitcells/data/`.
- Any unsupported features logged as per-sheet warnings.

### 3. Round-trip back to Excel

```
gitcells convert .gitcells/data/sample_chunks/ -o roundtrip.xlsx
```

Open `roundtrip.xlsx` and verify:
- All sheets present, in correct order.
- Cell values, formulas, merged cells match the source.
- Custom number formats render as expected (currency symbols, dates).
- Hyperlinks click through to the right targets.

### 4. Watch + atomic save

```
gitcells watch . &
```

In Excel: edit a cell in `sample.xlsx` and save (Cmd-S / Ctrl-S).

Within ~5 seconds expect:
- One commit on the current branch, message like
  `GitCells: modify sample.xlsx`.
- `git log --stat` shows only the changed sheet's chunk file in the
  diff, plus the workbook metadata.

### 5. Graceful shutdown drain

While `watch` is running, modify `sample.xlsx` in Excel and save.
**Before the 2-second debounce fires**, send SIGTERM to the watcher:

```
kill -TERM %1
```

Expect:
- Log shows `Shutting down (draining in-flight conversions)...`.
- The pending conversion completes; the commit lands before the
  process exits.

### 6. Locked-file retry (Windows-specific)

On Windows, with Excel still holding `sample.xlsx` open (no save
needed), run:

```
gitcells convert sample.xlsx
```

Expect:
- Either the conversion succeeds within ~10s (lock released
  transiently), or
- The error includes `sharing violation` / `being used by another
  process` and is logged with that exact phrase.

### 7. Failure log

Cause a deliberate failure: while `watch` is running, `chmod 000` the
`.gitcells/data` directory, then save the Excel file.

Expect:
- The watcher logs the error.
- `.gitcells/failures.log` has a line with the timestamp, event type,
  path, and underlying error.

Restore permissions: `chmod 755 .gitcells/data`.

### 8. Unicode + space sheet names

Use a workbook whose sheet names include a non-ASCII character and a
space (e.g. `データ table`).

Expect:
- A chunk file like `sheet_<sanitized>_<6hex>.json` exists.
- Round-trip preserves the original sheet name (inside the chunk
  payload, not the filename).

### 9. Cross-platform metadata read

Copy the `.gitcells/data/` tree from one OS to another (e.g. Linux
→ macOS). Run:

```
gitcells convert .gitcells/data/sample_chunks/ -o cross-platform.xlsx
```

Expect:
- The conversion succeeds.
- Open `.gitcells_chunks.json` and confirm all paths use forward
  slashes regardless of which OS wrote them.

### 10. Clean teardown

```
git status
```

Expect:
- Working tree clean.
- No stray staging or backup directories under `.gitcells/data/`.

---

## Pass criteria

All 10 steps pass with no warnings beyond the documented
**Not preserved** set in `docs/FIDELITY.md`. Any other warning is a
bug.
