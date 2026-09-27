# 138 — Allow panning in loom panes and make ansiviewer preview pane pannable

**Status**: Closed
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
   - Provide 2D offset support (`OffsetX`/`OffsetY` or viewport panning methods: `Pan`, `SetOffset`, `Offset`) in `loom.View`.
   - Horizontal offset slicing correctly clips/offsets ANSI and styled lines according to visual column widths without breaking ANSI escape codes or dropping color styles.
   - Support boundary clamping so offset cannot go negative or past content dimensions.

2. **Ansiviewer Panning Integration**:
   - Make the `Preview` pane in `examples/ansiviewer` pannable when focused.
   - Allow panning horizontally and vertically (e.g. arrow keys, `h`/`j`/`k`/`l`, `pgup`/`pgdn`, `home`/`end`).
   - Update status bar hints in `ansiviewer` to document panning navigation.

---

## 3. Milestones & Delivery

- **/goal**: Implement 2D panning support in Loom pane/view containers and make `ansiviewer`'s ANSI preview pane pannable with automated tests, or stop and report when blocked on a user decision or denied permission.

- **M1 (View / Pane 2D Panning Primitives)**:
  - Delivered in `a4bc213`: Added `OffsetX`/`OffsetY`, `Pan()`, `SetOffset()`, `Offset()`, ANSI-aware horizontal line clipping with style preservation, wide-glyph clipping, and unit tests in `view_test.go`.

- **M2 (Ansiviewer Preview Panning)**:
  - Delivered in `eda797f` & `7dbd968`: Wired up 2D panning (`offsetX`/`offset`) in `examples/ansiviewer/ansiviewer/viewer.go`, ANSI rendering with horizontal offset clipping in `writeANSI`, full keyboard panning navigation (`arrows`, `hjkl`, `pgup`/`pgdn`, `home`/`end`), updated status bar hints, and added regression tests in `viewer_test.go`.

- **M3 (Verification & Regression Gate)**:
  - Verified across entire test suite (`go test ./...`) and executed `make install`.
