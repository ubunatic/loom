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

### M1-M3 Review (host)
Delivered: `ParseANSI` (3cb3ed5), `Canvas.WriteANSI` (4878ad4), M3 evidence (cca8252). Tests and vet green.
Findings from the diff: M3 did not replace the local parser (accepted only if justified, see M4);
`M3-ansiviewer.ansi` therefore shows the local parser, not `WriteANSI`, which the evidence
must not claim. Parser defects: a bare `ESC[m` is not a reset, `22`/`23`/`24` are not handled,
and values above 255 wrap through `uint8`.

### M4 - Pre-Work / Required Refinements
1. `ParseANSI`: an empty SGR parameter list (`ESC[m`) and empty parts (`ESC[;1m`) count as 0.
   Add SGR `22` (bold and dim off) and `24` (underline off); `23` stays a no-op.
   Out-of-range 256-color or RGB values (over 255, negative) make the whole 38/48 sequence
   ignored, never wrapped. Each gets a failing test first.
2. Add a fuzz or property test: `ParseANSI` never panics, and every returned cell has either
   text, or is a continuation directly after a wide cell.
3. M3 honesty: in `viewer_m3_test.go`, the evidence test must not rewrite files under
   docs/ on every plain `go test` run; write only when env `LOOM_EVIDENCE=1` is set.
   Do the same for the evidence tests of M1 and M2 if they write to docs/.
4. M3 replacement: identify exactly which ansiviewer-local parser function handles plain
   SGR only (no cursor moves, no charset). If one exists, route SGR handling through
   `ParseANSI` or `WriteANSI`; keep the cursor/charset replay local. If SGR is entangled
   with cursor replay so it cannot be shared, add one sentence to the ticket-independent
   code comment in that file stating why, and report it. Existing ansiviewer tests, including
   the mc recordings, stay green.
5. Regenerate the three evidence frames with `LOOM_EVIDENCE=1` and commit them
   with the code, message ending '(issue 034 M4)'. Root `/dev/tty` failures are pre-existing.

### M4 Review (host)
Accepted: parser fixes, tests, `LOOM_EVIDENCE` gating, and the documented decision to keep
`applySGR` (entangled with cursor replay). One defect remains.

### M5 - Pre-Work
`viewer_m3_test.go` writes evidence to `../../docs/progress/034`, which resolves relative to
its package dir to `examples/docs/...` (a stray untracked dir now exists). Fix the path so
the frame lands in the repo-root `docs/progress/034/M3-ansiviewer.ansi` (for example
locate the repo root by walking up to `go.mod`), delete `examples/docs/`, regenerate
with `LOOM_EVIDENCE=1`, confirm `git status` shows only the intended file, and commit
'(issue 034 M5)'. Check the M1 and M2 evidence tests for the same relative-path mistake.
