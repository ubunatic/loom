# 227 — Fix media zoom reset and pan state instability

**Status**: Closed — fixed media zoom reset and pan state instability
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Bug
**Related**: `media/widget.go`, `gallery/gallery.go`, `issues/226-media-widget-scaling-zoom-cropping-and-centering-in-bounded-panels.md`

## Goal

`/goal`: Ensure `media.Widget` maintains its zoom level and viewport pan position reliably without unexpectedly resetting to 1:1 during mouse drag, keyboard panning, or tab switching, or stop and report when blocked on a user decision or denied permission.

## 1. Problem & Motivation

When zooming in on media in `loom widgets --show` (Gallery) or `examples/media`, starting to pan or interact with the widget sometimes resets the view back to 1:1 native resolution or clears the pan offsets.

## 2. Technical Findings & Root Cause Analysis

1. **Pan-axis reset in `w.pan()`**:
   In `media/widget.go`, `pan(dx, dy)` independently evaluates `targetCols > cols` and `targetRows > rows`. If one axis does not exceed bounds while the other does, the non-overflowing axis forcefully resets its offset (`w.panX = 0` or `w.panY = 0`), which can cause jarring jumps.
2. **Hard pan reset in `w.setZoom()`**:
   Every zoom adjustment (`setZoom`) forcefully sets `w.panX = 0, w.panY = 0`, discarding previous user pan position rather than scaling or clamping the pan offset smoothly around the zoom anchor.
3. **Mouse Drag & Hitbox Overlap**:
   In `ConsumeMouse`, clicking or dragging near the control buttons or widget boundary can trigger button actions or unhandled events if child-local coordinates and drag state transitions race with draw cycle rect updates.
4. **Gallery Container Tab Re-selection**:
   If unhandled key events (e.g., arrow keys or digits) bubble to `loom.Tabs` in `loom widgets --show`, tab state transitions or child re-initializations may reset widget instances.

## 3. Acceptance Criteria

- [x] `media.Widget` retains its active zoom and pan state during mouse dragging and keyboard panning without resetting to 1:1.
- [x] `setZoom` preserves relative pan offsets (clamping to the new maximum overflow bounds instead of unconditionally resetting to (0,0)).
- [x] Mouse drag transitions smoothly track `(dx, dy)` across the entire image viewport without losing drag capture or triggering unintended button clicks.
- [x] Tab and container key routing correctly handles pan arrow keys (`left`, `right`, `up`, `down`) without leaking to parent navigation.
- [x] Unit and PTY tests in `media/widget_test.go` and `gallery/gallery_test.go` cover continuous drag panning, zooming while panned, and event capture invariants.
- [x] `make test` and interactive testing in `loom widgets --show` pass cleanly (terminal interactions verified by automated real-PTY tests).

## 4. Implementation & Verification

- Shared rounded pixel geometry now drives scaling, zoom anchoring, pan clamping and drag eligibility. Zoom preserves the source center per axis; drag distance maps terminal cells to image pixels over half the overflow.
- Tabs captures handled left presses and routes drags/releases to the pressed child in child-local coordinates outside the panel. Release clears capture; active media drags ignore repeated presses over controls.
- Before the fix, regression tests reproduced pan resetting from `(.5,-.4)` to `(0,0)`, a control-crossing drag changing zoom from `2` to `1.6`, false overflow for a rounded 14-pixel image in seven quadblock cells, and lost tab drag capture.
- Final `make test-q1` passed schema validation, geometry replay, vet, unit tests and real Gallery PTY interactions. `make install` passed. Independent review confirmed geometry and capture behavior and verified completed-frame PTY snapshots and explicit release assertions.
