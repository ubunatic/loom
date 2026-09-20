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

---

## Milestones (lean sprint, dev agent: haiku)

The host reviews only diffs and test output. This ticket is the only channel. Root-package
tests that need `/dev/tty` fail before this work; ignore those. Commit each milestone
with `git add <your files>` then `git commit`; if commit fails, say so in your report.

Evidence: each milestone commits a headless frame to `docs/progress/034/M<N>-<what>.ansi`,
produced by a test or small program (never hand-written) that renders through `loom.RenderTo`
or the equivalent headless path. The frame must show real styled cells.

### M1 - ParseANSI
- `ParseANSI(s string) []Cell` in package loom: 16-color, 256-color, 24-bit, bold, dim,
  italic, underline, reset. Unknown or non-SGR sequences are dropped, never emitted as cells.
  Wide runes and combining marks must follow the canvas cell conventions.
- Table tests for every sequence class; benchmark on a typical colored line.
- Evidence: `M1-parse-samples.ansi` (a row per sequence class, rendered from parsed cells).

### M2 - Canvas.WriteANSI
- `func (c *Canvas) WriteANSI(x, y int, text string) int` returns cells written, clips to
  bounds like `Write`, preserves inline styles.
- Tests: clipping, wide runes at the edge, style reset mid-string.
- Evidence: `M2-writeansi.ansi`, plus a round trip (render, then parse the ANSI back, compare cells).

### M3 - Replace ansiviewer-local parser
- Pre-Work: none.
- Replace the private ANSI parser in `examples/ansiviewer/ansiviewer/` with `Canvas.WriteANSI`
  where it is a drop-in; keep behavior and all existing ansiviewer tests green
  (including the mc recordings). If part of the local parser has no equivalent, keep only that part and say why.
- Evidence: `M3-ansiviewer.ansi` from an ansiviewer render of a colored file.
