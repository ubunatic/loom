# 143 — Make the media example full-width and play video

**Status**: Closed — Implemented in 307c426 (M1): made examples/media full width and added video playback and PTY test coverage
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [111](111-add-an-image-media-widget-rendered-with-cati.md), [140](140-lazy-load-media-widget-images-and-video-previews.md)

---

## 1. Problem & Motivation

The `examples/media` app uses only a small fixed rectangle for the image, leaving much of the terminal unused and rendering media far smaller than its available area. It also only opens still images, so it does not demonstrate the library's video playback support.

## 2. Technical Specification / Findings

Use the available terminal width and height for the media area, fitting images to that area while preserving aspect ratio. Support video input in the example and start playback immediately through `media.NewVideo` and Loom's ticker/cati frame stream.

## 3. Implementation & Verification Plan

- **/goal**: Make the media example fill the current terminal and demonstrate immediately playing video with the media widget, or stop and report when blocked on a user decision or denied permission.
