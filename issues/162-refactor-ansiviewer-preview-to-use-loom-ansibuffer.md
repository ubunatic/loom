# 162 — Refactor ansiviewer preview to use loom.AnsiBuffer

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Architecture
**Related**: #100, #127, #148

---

## 1. Problem & Motivation
`loom view` in `cmd/loom` and other examples (`ansiedit`, `ansicanvas_demo`) use the standardized 2D cell grid primitive `loom.AnsiBuffer`.
In contrast, `examples/ansiviewer` still maintains an ad-hoc, ~100-line streaming escape-code tokenizer/parser (`writeANSI`, `applySGR`, `csiPosition`, etc.) in `examples/ansiviewer/ansiviewer/viewer.go`. This causes behavioral drift in wrapping, clipping, and ANSI handling between `ansiviewer` and `loom view`.

## 2. Technical Specification
Refactor `browser.Draw` and file selection in `examples/ansiviewer/ansiviewer/viewer.go`:
- When selecting an ANSI file (`KindANSI`), parse/load the file into a `*loom.AnsiBuffer` via `loom.ParseAnsiBuffer` or `loom.LoadAnsiBuffer`.
- In `browser.Draw()`, render `KindANSI` by copying cells from `*loom.AnsiBuffer` onto the canvas rect, clipped by scroll offset and horizontal offset (matching the pattern in `cmd/loom/main.go`).
- Retain `loom.View` for plain text (`KindText`), directory listings, and metadata.
- Clean up unused local ANSI stream tokenizer functions (`writeANSI`, `applySGR`, etc.) unless still needed for non-ANSI line rendering.

## 3. Milestones

### M1: Integrate AnsiBuffer into Browser Preview
- Update `browser` struct to hold `ansiBuf *loom.AnsiBuffer`.
- On selecting `KindANSI`, populate `ansiBuf` using `loom.ParseAnsiBuffer(string(data), 1, 1)`.
- In `browser.Draw()`, draw `ansiBuf` to `contentRect` when `b.kind == KindANSI`.
- Verify existing ansiviewer unit & PTY tests (`go test ./examples/ansiviewer/...`).
- Commit: `refactor(ansiviewer): use loom.AnsiBuffer for ANSI preview (issue 162 M1)`.

### M2: Cleanup Dead Code & Regression Verification
- Remove redundant streaming ANSI parsing helpers in `viewer.go` that are no longer referenced.
- Verify full test suite and clean build (`make test-q1`, `make install`).
- Commit: `refactor(ansiviewer): remove legacy streaming ANSI parser (issue 162 M2)`.

