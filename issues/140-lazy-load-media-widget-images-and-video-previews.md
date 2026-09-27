# 140 — Lazy load media widget images and video previews

**Status**: Closed
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

## 3. Milestones & Delivery

- **/goal**: Implement default background loading for media that takes longer than 50 ms, or stop and report when blocked on a user decision or denied permission.

- **M1 (Lazy Media Loading & Dim Indicator)**:
  - Delivered in `55c7dc9`: Updated `media/widget.go` with 50 ms threshold loading (`renderWithThreshold`). If media render completes within 50 ms, it renders immediately. If it takes longer, it renders asynchronously in the background while displaying a dim "loading" indication in the widget, storing the result in the cache for subsequent draws. Added unit tests in `media/widget_test.go` verifying immediate vs background delayed loading and completion.

- **M2 (Verification & Installation)**:
  - Verified with full test suite (`go test ./...`) and executed `make install`.
