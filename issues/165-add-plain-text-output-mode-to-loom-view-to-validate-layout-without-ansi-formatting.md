# 165 — Add plain text dump mode to loom view to validate layout without ANSI formatting

**Status**: Closed — loom view --plain/-p flag implemented and verified
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: #127, #129, #162, #163

---

## 1. Problem & Motivation

`loom view <file.ansi>` provides an interactive TUI pane to inspect ANSI files using `loom.AnsiBuffer`.
However, when validating layout, table alignment, box borders, and text columns, terminal colors and raw escape sequences can mask visual alignment issues, or make it difficult to inspect the pure rendered character grid in tests, pipes, or diffs (`diff -u`).

Having a non-interactive plain text mode (e.g. `loom view --raw`, `loom view --plain`, or `loom view --text`) that echoes back the parsed 2D grid as plain text (with colors and control sequences stripped) allows developers and scripts to:
1. Validate that box layout, unicode width calculations, and spacing effects are preserved.
2. Diff rendered ANSI outputs against expected plain-text fixtures without ANSI noise.
3. Inspect files non-interactively in standard pipelines.

## 2. Technical Specification / Findings

- **CLI Flag in `cmd/loom`**:
  - Add a flag to `viewCommand()` in `cmd/loom/main.go`, such as `--plain` / `-p` (or `--text` / `--raw`):
    - When flag is specified, bypass `loom.New(24)` interactive pane runner.
    - Load/parse the file into `*loom.AnsiBuffer` via `loom.ParseAnsiBuffer(string(data), 1, 1)`.
    - Render and print the buffer grid as plain text (e.g., via a helper `buf.PlainText()` or `buf.String()` that extracts characters without SGR/escape codes).
    - Trailing spaces on each row can be trimmed, matching `AnsiBuffer.Serialize()` conventions.
- **AnsiBuffer Plain Text Export**:
  - Check if `AnsiBuffer` currently exposes a plain-text export method.
  - If not, provide an exported method on `AnsiBuffer` (e.g. `PlainText() string` or `buf.PlainText()`) that iterates over rows and emits runes without escape sequences.

## 3. Milestones & Delivery

### M1: Add Plain Text Export to AnsiBuffer and --plain Flag to loom view
- **Delivered**: `a29def2` (*"feat(loom): add plain text dump mode to loom view (issue 165 M1)"*).
- Added `(*AnsiBuffer).PlainText() string` in `ansibuffer.go` to emit 2D grid contents without styling or escape sequences, trimming trailing blank cells per row.
- Added `--plain` / `-p` flag to `loom view` in `cmd/loom/main.go`, writing plain text output directly to stdout without initializing the interactive TUI pane.
- Added unit tests in `ansibuffer_test.go` and `cmd/loom/main_test.go`.
- Tested `loom view --plain docs/data/mc-julia256.ansi` successfully producing clean Unicode plain text.
- Verified with `make test-q1` and `make install`.
