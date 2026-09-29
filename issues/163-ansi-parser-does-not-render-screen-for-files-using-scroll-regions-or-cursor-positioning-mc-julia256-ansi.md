# 163 — ANSI parser exits prematurely or drops screen when files use cursor positioning (mc-julia256.ansi)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Bug
**Related**: #127, #162

---

## 1. Problem & Motivation

Viewing `docs/data/mc-julia256.ansi` with `loom view docs/data/mc-julia256.ansi` or `ansiviewer docs/data/mc-julia256.ansi` renders an empty screen.

`mc-julia256.ansi` is a real-world terminal capture from Midnight Commander. It uses VT100/ANSI sequences including scroll margin configurations (`ESC [ 1 ; 16 r`), cursor positioning (`ESC [ 16 ; 1 H`), and absolute row/column coordinates (`ESC [ row ; col H`).

In `loom.ParseAnsiBuffer` (`ansibuffer.go`):
1. Buffer dimensions are initially estimated solely from newline count:
   `lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")`
   `maxRows := max(minRows, len(lines))`
   For `mc-julia256.ansi`, `len(lines)` is only 13 because much of the full-screen drawing is accomplished via cursor positioning (`CSI H`) rather than sequential newlines.
2. The main parsing loop condition in `ansibuffer.go`:
   `for i := 0; i < n && y < maxRows;`
   At the very start of the file (byte 7), sequence `\x1b[16;1H` positions the cursor at row 16 (`y = 15`). Because `maxRows == 13`, `y < maxRows` evaluates to `15 < 13` (**false**), immediately terminating the entire parser loop after reading only 14 characters out of 2114+!
3. Furthermore, when cursor positioning targets rows beyond `maxRows` or columns beyond `maxCols`, or when scroll region sequences (`CSI r`) dictate the screen height, `ParseAnsiBuffer` should either dynamically grow the buffer or pre-scan CSI coordinates/scroll regions to size the grid accurately instead of aborting the parse loop.

## 2. Technical Specification / Findings

- **Root Cause**:
  - `ansibuffer.go:692`: `for i := 0; i < n && y < maxRows;` aborts the parser if `y >= maxRows`.
  - Initial `maxRows` and `maxCols` in `ParseAnsiBuffer` only look at `\n` splits and `stripANSI(line)` lengths without scanning for `\x1b[<row>;<col>H`, `\x1b[<row>d`, `\x1b[<col>G`, or `\x1b[<top>;<bottom>r` (scroll margin).
  - Even if `y` goes temporarily out of bounds or points to the bottom status line, the parser must not stop consuming the stream (`i < n`).
- **Required Behavior**:
  - `ParseAnsiBuffer` must parse the entire input stream (`for i < n`).
  - Either pre-scan CSI row/col positions (`H`, `f`, `d`, `G`, `r`) during dimension sizing, or dynamically expand `buf.cells` / clamp coordinates within buffer bounds so content is not lost.
  - Correctly render `docs/data/mc-julia256.ansi` in `loom.ParseAnsiBuffer`.

## 3. Implementation & Verification Plan

- **/goal**: Fix `loom.ParseAnsiBuffer` to parse full ANSI streams containing cursor positioning and scroll regions without early termination, verifying with `docs/data/mc-julia256.ansi` and automated tests, or stop and report when blocked on user input or denied permission.
- Add regression tests in `ansibuffer_test.go` loading `docs/data/mc-julia256.ansi` and verifying that non-empty cells and midnight commander panel text ("projects/loom", "Name", "Size") are populated.
- Verify `loom view` and `ansiviewer` display the content.
- Ensure all tests pass with `make test-q1`.
