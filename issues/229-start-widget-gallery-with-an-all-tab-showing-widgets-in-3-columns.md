# 229 — Start widget gallery with an All tab showing widgets in 3 columns

**Status**: Closed — implemented 3-column All overview tab with unified focus handling and full test suite
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `gallery/gallery.go`, `grid.go`, 195, 198, 200, 207

---

## 1. Problem & Motivation

Currently, the `loom widgets` gallery starts on an individual widget tab. Starting the gallery with an "All" overview tab that presents a collection of widgets arranged in 3 columns (small ones first) will demonstrate a rich, dense TUI on a single screen and allow checking that widgets render and interact well together.

## 2. Technical Specification / Findings

- Add an "All" tab in `gallery/gallery.go` set as the initial default active tab.
- Arrange smaller widgets in a 3-column layout (e.g. Button, Toggle, Checkbox, NumberInput, Badge, PillCluster, ProgressBar, Sparkline, etc.) using Loom's layout/grid capabilities.
- Ensure keyboard focus, tab cycling, and mouse event routing function seamlessly across the multi-widget composite layout.

## 3. Implementation & Verification Plan

### Milestone 1 (Delivered: ed2df12d)
- Constructed 3-column "All" composite layout in `gallery/gallery.go`.
- Prepend "All" tab to `NewAll()` so it starts on the 3-column overview.
- Updated `grid.go` mouse handling to route clicks to child widgets.
- Added comprehensive layout and interaction tests in `gallery/gallery_test.go` and refreshed `docs/progress/gallery.ansi`.

### Milestone 2: Pre-Work & Refinements
- Run `gofmt -w` on `grid.go` and `gallery/gallery_test.go`.
- In `grid.go`, guard child index bounds when setting/clearing focus (`i < len(g.Children)`).
- Unify mouse and keyboard focus transitions using a shared helper `g.setFocus(i)` in `grid.go`.
- Add a direct unit test in `mouse_test.go` verifying that mouse press changes grid focus and scroll does not.
- Adjust any demo widget text widths (e.g. PillCluster or button/checkbox click widths) so 3-column display in 80x24 does not truncate.
- Strengthen layout test assertions to verify exact column cell boundaries (`r.X + k*cellW`).
- Run `make test-q1` and commit with message ending in `(issue 229 M2)`.

/goal Add the 3-column All overview tab as the default gallery starting tab and verify with tests, or stop and report when blocked on a user decision or denied permission.
