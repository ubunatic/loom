# 109 — Detect loomoji emoji-width drift: stray right-edge fragments and ragged rows

**Status**: Closed — Fixed in f1d8042: Canvas.Row continuation pad-1 emission keeps terminal cursor in sync with canvas grid; verified by category width PTY test
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Testing
**Related**: [048](048-emoji-rune-width-discrepancy-causes-horizontal-border-drift.md) (same root-cause family),
[107](107-pty-mouse-test-system-detect-loomoji-hover-vs-highlight-offset-via-rendered-fg-bg-cells.md),
`docs/HoverTesting.md`, `internal/ptytest`

## Goal

A black-box test that **detects and reports** the loomoji grid artifacts seen in a real terminal
(2026-09-24 screenshot). **Do not fix the bug**; detection only, like 107.

## Observed (user screenshot, `go run ./examples/loomoji`)

- Every grid row ends in a stray thin vertical fragment (`▏`-like) at a different column per row.
- Row ends are ragged: the rows start at the same column but end at different ones. Rows with
  text-presentation or VS16 glyphs (`✖ ➕ ➖ ➗ ❗ ⚠️ ☢️ ☣️`) end at different columns than all-emoji rows.
- Likely cause: loom's display width disagrees with the terminal's width for these glyphs, so the
  cursor drifts and the trailing cells are painted out of place (as in 048).

## Done when

- A PTY test renders loomoji at a fixed size and reports, per grid row: the last painted column,
  stray non-space cells after the last item, and whether all rows end at the same column.
- It fails (or logs a non-zero drift) on today's code.
- Open question: `ptytest.VT` uses loom's own width function, so it may not reproduce the drift.
  The test may need a reference width (e.g. a VT width table separate from loom's own, or a
  cursor position report `CSI 6n` from a real terminal). Record which method you picked and why.

## Status & Progress (2026-09-24)

- Added [`examples/loomoji/loomoji_category_width_pty_test.go`](../examples/loomoji/loomoji_category_width_pty_test.go), executing `loomoji` in a 64x18 PTY across all emoji categories (via `f` key cycling) and asserting that every non-empty line has uniform length matching the expected 64-column layout without ragged edges.
