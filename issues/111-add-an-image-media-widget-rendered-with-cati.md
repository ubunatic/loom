# 111 — Add an Image/Media widget rendered with cati

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [112](112-add-media-controls-play-pause-zoom-and-panning-for-the-image-media-widget.md) (controls), [090](090-support-image-backed-app-backgrounds-and-background-theme-switching.md),
`background.go` (own `image.Decode` path), `../cati` (`ubunatic.com/cati/v1/...`)

## Goal

A hostable `loom.Widget` that shows an image (PNG/JPEG/SVG) or a video/frame sequence inside its
rect, rendered through cati's public `v1` packages instead of loom's own renderer.

## Done when

- The widget draws the media scaled to fit its rect (aspect preserved) on the loom `Canvas`,
  with a selectable cati render mode (e.g. halfblock/quadblock/sextant).
- Multi-frame media advances on loom's Ticker; still images don't tick.
- A small example app shows it, plus unit tests and a PTY colour-cell check (107 method) that
  the image area is painted and nothing spills outside the rect.
- Recorded here: whether cati's output (ANSI lines) goes through `Canvas.WriteANSI`/ParseANSI or a
  direct cell API; whether `background.go` (090) should switch to the same path.

## Notes

- Canary first (`docs/Canary.md`): probe cati's v1 API and a go.mod dependency (local
  `replace` vs published module) before building on it. Check live cati and loom state first.
