# 161 — Align media example with redesign 005

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [143](143-make-the-media-example-full-width-and-play-video.md), [145](145-show-measured-screen-dimensions-and-debug-data-in-a-status-bar-below-the-media-demo.md), [160](160-media-widget-playback-controls-and-poster-frame-support.md), `docs/data/media-demo-redesign-005.ansi`

---

## 1. Problem & Motivation
The current `examples/media` layout does not match the selected design in `docs/data/media-demo-redesign-005.ansi`. Adopting that compact layout will make the media preview and its state and controls easier to scan without redundant chrome.

## 2. Technical Specification / Findings
The mockup places the preview on the left and the filename, playback state, render mode, and `p`/`r`/`q` controls in a simple column on the right. It has no source panel, box borders, poster hint, or repeated labels. The example already uses `media.Widget` and supports playback controls.

## 3. Implementation & Verification Plan
**/goal**: Implement the selected 005 layout in `examples/media` and verify it at the mockup's terminal size, or stop and report when blocked on a user decision or denied permission.

**Acceptance Criteria**
- [ ] The example follows the selected mockup while retaining the media widget and its keyboard controls.
- [ ] The layout adapts without clipping at smaller terminal sizes.
- [ ] Relevant checks pass.
