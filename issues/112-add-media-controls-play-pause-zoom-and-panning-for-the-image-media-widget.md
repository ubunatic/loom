# 112 — Add media controls: play/pause, zoom +/-, and panning for the Image/Media widget

**Status**: Closed
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: [111](111-add-an-image-media-widget-rendered-with-cati.md) (depends on it), [089](089-add-mouse-cursor-position-hints-and-configurable-visual-effects.md)

## Goal

Interactive controls for the 111 Image/Media widget: a play|pause button for multi-frame media,
zoom +/- controls, and panning when the media is zoomed in.

## Done when

- Play|pause toggles by clicking a button and pressing a key; the button shows its state.
- Zoom +/- by clicking buttons and pressing keys, with limits; zoom-out stops at fit-to-rect.
- Panning while zoomed works with the arrow keys and mouse drag, and is clamped to the media bounds.
  Mouse events are 0-based and child-local (`docs/Widgets.md`).
- Unit tests for the state (zoom level, pan offset clamping, play state), and a PTY test that
  clicks the controls and checks the rendered change (`docs/HoverTesting.md` method).

## Sprint goal (roadmap 180)
/goal Add play/pause, zoom, and pan controls on top of 160's playback API with state and PTY tests, and an example for human review; stop and report when control layout needs a user decision.
