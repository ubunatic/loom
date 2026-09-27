# 138 — Allow panning in loom panes and make ansiviewer preview pane pannable

**Status**: Open
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
   - Provide 2D offset support (`OffsetX`/`OffsetY` or viewport panning) in `loom.View` or pane viewport rendering.
   - Support keyboard navigation (e.g. arrow keys, `h`/`j`/`k`/`l`, or Shift+scroll/arrows) and mouse drag/scroll events for panning viewports.
   - Integrate with scrollbar indicators if active to reflect horizontal and vertical offset positions.

2. **Ansiviewer Panning Integration**:
   - Make the `Preview` pane in `examples/ansiviewer` pannable when focused.
   - Allow panning horizontally across wide `.ansi` mockups (e.g. 100+ columns) and vertically across long documents.
   - Update status bar hints in `ansiviewer` to document panning shortcuts.

---

## 3. Implementation & Verification Plan

- **/goal**: Implement 2D panning support in Loom pane/view containers and make `ansiviewer`'s ANSI preview pane pannable with automated tests, or stop and report when blocked on a user decision or denied permission.
- Add unit and PTY/canvas tests in `view_test.go` and `examples/ansiviewer/ansiviewer/` validating viewport offset adjustments, boundary clamps, and keyboard/mouse panning actions.
- Verify `make test` and `make install`.
