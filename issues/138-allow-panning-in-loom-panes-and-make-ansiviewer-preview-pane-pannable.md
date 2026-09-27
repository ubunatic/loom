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
  - Delivered in `a4bc213`: Added `OffsetX`/`OffsetY`, `Pan()`, `SetOffset()`, `Offset()`, ANSI-aware horizontal line clipping with style preservation, wide-glyph clipping, and unit tests in `view_test.go`.

- **M2 (Ansiviewer Preview Panning)**:
  - **Scope**:
    - Wire up horizontal and vertical panning in `examples/ansiviewer/ansiviewer/viewer.go` for the preview pane.
    - Support horizontal offset in `browser` (`offsetX` / `b.previewView.OffsetX`) and ensure `writeANSI` or line rendering applies horizontal offset clipping when drawing ANSI and text content.
    - Support panning key bindings (`left`/`right`/`up`/`down`, `h`/`l`/`j`/`k`, `pgup`/`pgdn`, `home`/`end`) in `browser.HandleKey`.
    - Update `framedBrowser` status line to include panning instructions (`hjkl/arrows pan`).
    - Add tests in `examples/ansiviewer/ansiviewer/viewer_test.go` verifying horizontal & vertical panning on ANSI content.
  - Deliverable commit: `feat(ansiviewer): enable 2D panning in preview pane (issue 138 M2)`

- **M3 (Verification & Regression Gate)**:
  - Run full test suite (`make test`), verify `make install`.
