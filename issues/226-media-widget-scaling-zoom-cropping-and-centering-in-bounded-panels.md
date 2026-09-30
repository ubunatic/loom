# 226 — Media widget scaling, zoom, cropping, and centering in bounded panels

**Status**: Closed — implemented media widget scaling, zoom, cropping, and centering in bounded panels
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Bug
**Related**: `media/widget.go`, `gallery/gallery.go`, `examples/media/`

## Goal

`/goal`: Support proper 1:1 pixel/cell mapping, fractional/multiple zoom scaling, viewport cropping, and centering for `media.Widget` within bounded parent containers (such as the `loom widgets --show` gallery panel), or stop and report when blocked on a user decision or denied permission.

## 1. Problem & Motivation

In bounded containers like the widget gallery (`loom widgets --show`), `media.Widget` does not correctly handle 1:1 unzoomed display, zoom scaling, and viewport cropping/centering relative to the panel dimensions. Currently, `cropForView` only operates when `zoom > 1` and relies on cati auto-fitting the image to the full viewport rect, distorting pixel/cell relationships or preventing smaller media from centering unscaled and larger media from cropping at 1:1.

## 2. Technical Specification & Examples

In halfblock mode, 1 terminal cell represents 1 pixel horizontally × 2 pixels vertically (cell aspect ~1:2).

### Example 1: Small Media (10x10 px -> 10x5 cells) in a 20x10 cell panel
- **1x zoom (1:1)**: Neither zoomed in nor out. The 10x5 cell image is rendered at 1:1 and centered within the 20x10 cell panel.
- **2x zoom**: Scales 2x to 20x10 cells, perfectly filling the 20x10 panel.
- **3x zoom**: Scales 3x to 30x15 cells, cropped to the 20x10 panel (centered crop fills the panel).
- **0.5x zoom**: Scales down to 5x2.5 cells (bottom row partially/half transparent if applicable) and centered in the 20x10 panel.

### Example 2: Large Media (40x40 px -> 40x20 cells) in a 20x10 cell panel
- **1x zoom (1:1)**: 40x20 cells exceeds the 20x10 panel; media is centered and cropped to 20x10.
- **2x zoom**: 80x40 cells, further cropped to 20x10.
- **0.5x zoom**: Scales down to 20x10 cells, fully fitting the 20x10 panel without cropping.
- **0.25x zoom**: Scales down to 10x5 cells, centered in the 20x10 panel with padding.

## 3. Acceptance Criteria

- [ ] `media.Widget` at 1x zoom displays media at native cell resolution (1:1 mapping based on mode), centering smaller media and cropping larger media within the target `imageRect`.
- [ ] Zooming in (> 1x) scales the media upward from center, cropping overflow at panel bounds.
- [ ] Zooming out (< 1x) scales the media downward from center, centering the smaller footprint within the panel.
- [ ] Interactive pan controls (arrows / drag) allow shifting the viewport over cropped/zoomed content.
- [ ] Halfblock, quadblock, and sextant modes calculate native cell dimensions correctly according to their sub-pixel division.
- [ ] Unit and visual/PTY tests verify centering, scaling, and cropping across small and large media fixtures.
- [ ] `make test` and `loom widgets --show` (Gallery) render correctly.
