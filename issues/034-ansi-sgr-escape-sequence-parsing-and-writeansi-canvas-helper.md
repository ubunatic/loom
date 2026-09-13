# 034 — ANSI SGR Escape Sequence Parsing and WriteANSI Canvas Helper

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [docs/HarnezUsageTarget.md](file:///home/uwe/projects/loom/docs/HarnezUsageTarget.md), [canvas.go](file:///home/uwe/projects/loom/canvas.go), [style.go](file:///home/uwe/projects/loom/style.go)

---

## 1. Problem & Motivation
When integrating `loom` with external text generators, syntax highlighters, or diagram engines (such as `termaid`), applications frequently receive strings pre-formatted with ANSI SGR escape sequences (e.g. `\x1b[1;38;5;39m`, `\x1b[32m`, `\x1b[0m`).

Currently:
- `Canvas.Write(x, y, text, style)` strips all ANSI escape codes via `plainTerminalText` and applies a single, uniform `loom.Style` across the entire string.
- `View.Draw` similarly calls `stripANSI(v.Lines[lineIdx])`, discarding all embedded ANSI color and style formatting.

As a result, consumers who wish to render colored, syntax-highlighted, or diagrammatic output into a `loom.Canvas` must implement a custom ANSI SGR escape code parser that translates escape codes into per-cell `loom.Cell` structures and `loom.Style` attributes.

## 2. Proposed Solution
1. **ANSI Parsing Helper**:
   - Provide `ParseANSI(s string) []Cell` or `ANSIStringToCells(s string) []Cell` in `loom` (or `loom/measure`) that decodes standard 3/4-bit, 8-bit (256-color), 24-bit (RGB), and text attributes (bold, dim, underline, reset) into a slice of `loom.Cell` structs.
2. **Canvas WriteANSI Method**:
   - Provide `c.WriteANSI(x, y int, text string) int` on `*Canvas` to write ANSI-formatted strings directly to the canvas, preserving inline colors and styles without discarding them.
3. **Optional Rich Line Support in `View`**:
   - Allow `View` to accept pre-styled lines or automatically render ANSI-styled lines without stripping attributes.

## 3. Verification & Acceptance
- Unit tests verifying 16-color, 256-color, 24-bit RGB, bold, dim, underline, and reset sequences correctly populate `Cell.Style`.
- Benchmark ensuring minimal overhead when parsing typical colored output lines.
