# 238 — Gallery All grid: borders and content-fitted row heights

**Status**: Open
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
