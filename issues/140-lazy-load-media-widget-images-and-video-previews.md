# 140 — Lazy load media widget images and video previews

**Status**: In Progress
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [112](112-add-media-controls-play-pause-zoom-and-panning-for-the-image-media-widget.md)

---

## 1. Problem & Motivation

Images and video previews in the media widget can delay rendering while they load. Make media loading lazy by default: if an image or video preview has not loaded within 50 ms, continue loading it in the background and show a dim “loading” indication in the widget while waiting.

---

## 2. Technical Specification / Findings

The existing media widget controls are tracked in [112](112-add-media-controls-play-pause-zoom-and-panning-for-the-image-media-widget.md). Preserve normal rendering once the media finishes loading.

---

## 3. Milestones & Plan

- **/goal**: Implement default background loading for media that takes longer than 50 ms, or stop and report when blocked on a user decision or denied permission.

- **M1 (Lazy Media Loading & Dim Indicator)**:
  - Update `media/widget.go` to lazy-load media with a 50 ms threshold: if loading finishes within 50 ms, render immediately; otherwise, continue in the background and render a dim "loading" indication in the widget.
  - When background load finishes, seamlessly update widget state for subsequent draws (or trigger redraw).
  - Add automated unit tests in `media/widget_test.go` covering immediate load, delayed background load with "loading" indicator, and eventual render completion.
  - Deliverable commit: `feat(media): lazy load media widget images and video previews (issue 140 M1)`

- **M2 (Verification & Installation)**:
  - Run full test suite (`go test ./...`), `make test-q1`, and `make install`.
