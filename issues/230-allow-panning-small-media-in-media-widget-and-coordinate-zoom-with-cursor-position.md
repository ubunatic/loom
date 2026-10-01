# 230 — Allow panning small media in Media widget and coordinate zoom with cursor position

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `media/widget.go`, `media/widget_test.go`, 112, 202, 226, 227

---

## 1. Problem & Motivation

When small media is displayed inside a larger Media widget (e.g. 10x5 media on a 20x10 Media pane), the widget currently centers the media and forbids panning (`canPanLocked()` checks require overflow). Users need to be able to reposition the image anywhere in the media pane using keyboard arrow keys and mouse dragging. Additionally, zooming in/out should coordinate cleanly with the zoom center (the mouse cursor position when zooming) rather than solely the image center, providing a natural zoom and pan experience while preserving panning offsets.

## 2. Technical Specification / Findings

- Relax panning bounds in `media/widget.go` so that media smaller than the viewport can still be panned across the pane with keyboard and mouse drag.
- In zoom calculations, anchor zoom transformations around the current mouse cursor location (`MouseEvent.X`, `MouseEvent.Y`) during mouse wheel or zoom clicks.
- Maintain smooth coordination between cursor-centered zooming and free panning.

## 3. Implementation & Verification Plan

### Milestone 1 (Delivered: 239490b)
- Added `panDelta` / `viewSlack` helpers for unified small/large pan range [-1,1].
- Updated `canPanLocked()` to allow dragging small images.
- Added `setZoomAt(factor, cursorX, cursorY)` for cursor-anchored mouse wheel zoom.
- Added `anchorPan` function keeping the image point under cursor fixed through zoom.
- Drag sign unified across small and large images.
- Tests: keyboard pan, screen-position draw test, mouse drag, cursor zoom in/out, geometry regression.
- `gofmt`, `go vet`, `go test ./media` all clean.

### Milestone 2: Pre-Work / Required Refinements
- Add three missing test cases in `media/widget_test.go` (no code changes needed):
  1. Zoom step crossing the small→large and large→small boundary: verify `anchorPan` handles the delta sign change at the crossing point.
  2. Mouse wheel scroll aimed at the control bar row (cursor below image area): verify clamp handles out-of-image cursor gracefully.
  3. One axis smaller than viewport, other axis overflowing: verify `panDelta` and `viewSlack` are correct in each axis independently.
- Run `make test-q1`, commit with message ending in `(issue 230 M2)`. Run `make install`. Report the commit hash.

/goal Enable small media panning and cursor-centered zooming in Media widget, verify with tests, or stop and report when blocked on a user decision or denied permission.
