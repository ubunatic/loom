# 238 — Gallery All grid: borders and content-fitted row heights

**Status**: In Progress
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 232 (more widgets in the All tab)

---

/goal The gallery's "All" tab shows each widget in a bordered cell, and each grid row is as tall as
the tallest widget in that row, with tests; or stop and report when blocked on a user decision
(e.g. Grid API shape).

## 1. Problem & Motivation
User request: add borders to the "All" grid and use a grid that fits its content. Today every cell
gets the same height (`r.H / rows`), so a single-line NumberInput sits in a ~5-row cell. The biggest
widget in each row should decide that row's height.

## 2. Technical Specification / Findings
- `Grid.Draw` (`grid.go`) splits the area into equal cells. The "All" tab is `newAllDemo()` in
  `gallery/gallery.go` (3 columns, 24 widgets).
- loom has `Measurer` (`Measure(width) measure.Size`, `widget.go`) for preferred sizes. Widgets
  without it need a sensible fallback height.
- Fix in the library (AGENTS.md: fix the library, not the caller): an opt-in content-fitted row mode
  and optional cell borders on `Grid`, not a gallery-only layout.
- Mouse/focus routing must follow the new cell rects (event routing: start on `codex:sol:med`).
- Open: when rows exceed the screen, scroll or clip? Record the choice.

## 3. Implementation & Verification Plan
- Tests: row height = max measured height in the row; borders drawn; mouse hits map to the right child.
- `make test-q1`, `make install`, PTY check of `loom widgets --show All`.

**Handoff (2026-10-06):** `99b5e70` adds opt-in `Grid.FitRows`, used by the gallery All tab. dev238 fixed a vet error in `grid_fit_test.go` but never reran the suite. The host test-q1 run (log: scratchpad `q1-238-225.log`) was still running at wrap. Next: check `make test-q1`, then `make install`, then close.
**Result (2026-10-06):** host `make test-q1` FAILED, with 2 gallery tests, probably because FitRows changed the All layout: `TestAllDialogAndTableNavigationPTY` and `TestAllTabAddedWidgetsStayInCellsAndRouteInput/Chart`. Log: `/tmp/loom-238-225-test-q1.log`. All other packages passed, including the 225 moves. Fix these two, then rerun test-q1 and close 238 and 225.
