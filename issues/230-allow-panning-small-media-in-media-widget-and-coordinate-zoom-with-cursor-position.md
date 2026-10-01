# 230 — Allow panning small media in Media widget and coordinate zoom with cursor position

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `media/widget.go`, 112, 202, 226, 227

---

## 1. Problem & Motivation

When small media is displayed inside a larger Media widget (e.g. 10x5 media on a 20x10 Media pane), the widget currently centers the media and forbids panning (`canPanLocked()` checks require overflow). Users need to be able to reposition the image anywhere in the media pane using keyboard arrow keys and mouse dragging. Additionally, zooming in/out should coordinate cleanly with the zoom center (the mouse cursor position when zooming) rather than solely the image center, providing a natural zoom and pan experience while preserving panning offsets.

## 2. Technical Specification / Findings

- Relax panning bounds in `media/widget.go` so that media smaller than the viewport can still be panned across the pane with keyboard and mouse drag.
- In zoom calculations, anchor zoom transformations around the current mouse cursor location (`MouseEvent.X`, `MouseEvent.Y`) during mouse wheel or zoom clicks.
- Maintain smooth coordination between cursor-centered zooming and free panning.

## 3. Implementation & Verification Plan

- Update panning logic in `media/widget.go` to support small media panning across available viewport space.
- Adjust zoom calculations to factor in mouse cursor coordinates.
- Add unit tests in `media/widget_test.go` covering small-image panning and cursor-centered zooming.
- Verify with `make test-q1`.

/goal Enable small media panning and cursor-centered zooming in Media widget, verify with tests, or stop and report when blocked on a user decision or denied permission.
