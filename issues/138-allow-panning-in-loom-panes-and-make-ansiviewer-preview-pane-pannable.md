# 138 — Allow panning in loom panes and make ansiviewer preview pane pannable

**Status**: In Progress
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: `pane.go`, `view.go`, `examples/ansiviewer`, `issues/112`, `issues/137`

---

## 1. Problem & Motivation

When inspecting large terminal assets, wide ANSI mockups (such as multi-column layouts created by `cati`), or detailed preview buffers in Loom, content that exceeds the viewport dimensions is clipped. While vertical scrolling exists in certain widgets, Loom panes and preview containers lack generalized 2D panning (horizontal and vertical scrolling/offset controls), preventing users from viewing offscreen columns and rows in `ansiviewer`.

---

## 2. Technical Scope & Requirements

1. **Loom Pane / View Panning Support**:
   - Provide 2D offset support (`OffsetX`/`OffsetY` or viewport panning methods: `Pan`, `SetOffset`, `Offset`) in `loom.View` (or relevant pane/view primitives).
   - Horizontal offset slicing should correctly clip/offset ANSI and styled lines according to visual column widths without breaking ANSI escape codes.
   - Support boundary clamping so offset cannot go negative or past content dimensions.

2. **Ansiviewer Panning Integration**:
   - Make the `Preview` pane in `examples/ansiviewer` pannable when focused.
   - Allow panning horizontally and vertically (e.g. arrow keys, `h`/`j`/`k`/`l`, or Shift+scroll/arrows when preview focused).
   - Update status bar hints in `ansiviewer` to document panning navigation.

---

## 3. Milestones & Plan

- **/goal**: Implement 2D panning support in Loom pane/view containers and make `ansiviewer`'s ANSI preview pane pannable with automated tests, or stop and report when blocked on a user decision or denied permission.

- **M1 (View / Pane 2D Panning Primitives)**:
  - Implement 2D offset / panning (`OffsetX`, `OffsetY`, `Pan(dx, dy int)`) in `loom.View`. Ensure horizontal offset clipping preserves styling.
  - Add comprehensive unit tests in `view_test.go`.
  - Deliverable commit: `feat(view): add 2D offset and panning support in View (issue 138 M1)`

- **M2 (Ansiviewer Preview Panning)**:
  - Wire up horizontal and vertical panning in `examples/ansiviewer` for the preview pane when focused.
  - Handle key events (arrows, `h`/`j`/`k`/`l` when focused) and mouse scroll/events.
  - Update status bar instructions and add automated tests in `examples/ansiviewer/ansiviewer/viewer_test.go`.
  - Deliverable commit: `feat(ansiviewer): enable 2D panning in preview pane (issue 138 M2)`

- **M3 (Verification & Regression Gate)**:
  - Run full test suite (`go test ./...`), verify `make test` and `make install`.
