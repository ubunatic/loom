# 140 — Lazy load media widget images and video previews

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [112](112-add-media-controls-play-pause-zoom-and-panning-for-the-image-media-widget.md)

---

## 1. Problem & Motivation
Images and video previews in the media widget can delay rendering while they load. Make media loading lazy by default: if an image or video preview has not loaded within 50 ms, continue loading it in the background and show a dim “loading” indication in the widget while waiting.

## 2. Technical Specification / Findings
The existing media widget controls are tracked in [112](112-add-media-controls-play-pause-zoom-and-panning-for-the-image-media-widget.md). Preserve normal rendering once the media finishes loading.

## 3. Implementation & Verification Plan
**Goal**: Implement default background loading for media that takes longer than 50 ms, or stop and report when blocked on a user decision or denied permission.

**Done when**:
- Lazy loading is the default for images and video previews, with media still loading in the background after the 50 ms threshold.
- The widget shows a dim “loading” indication while waiting, then displays the loaded media.
- Relevant automated checks cover the loading state and completion behavior.
