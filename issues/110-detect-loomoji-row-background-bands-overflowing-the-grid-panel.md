# 110 — Detect loomoji row background bands overflowing the grid panel

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Testing
**Related**: [109](109-detect-loomoji-emoji-width-drift-stray-right-edge-fragments-and-ragged-rows.md),
[107](107-pty-mouse-test-system-detect-loomoji-hover-vs-highlight-offset-via-rendered-fg-bg-cells.md),
`docs/HoverTesting.md`

## Goal

A black-box colour-cell test that **detects and reports** grid row backgrounds painted outside
the loomoji panel. **Do not fix the bug**; detection only.

## Observed (user screenshot, 2026-09-24)

- The row background bands have different widths. The first row's band runs past the
  panel's right edge (past the search bar's right edge), and the others stop short at different columns.
- The last row (`🅿️`) has no band.
- The bands do not line up with the search bar or with each other.

## Done when

- A PTY test using the 107 colour-cell grid (`Cells()`, `Style.Effective()`) finds the panel
  rect from the screen (e.g. search-bar extent) and reports, per row, the band's start and end
  columns and any cells painted outside the panel.
- It fails (or logs the overflow) on today's code. Note which findings disappear once 109's
  width drift is fixed, since they may share a root cause.
