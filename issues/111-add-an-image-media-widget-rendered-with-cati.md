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

## Sprint Log

**Canary (dev-111, 2026-09-27):** cati v1 exposes `halfblock`, `quadblock` and `sextant`, each with `RenderToGrid` (a `core.Grid` of cells with glyph, fg and bg) and `Render(io.Writer)` (ANSI). Frames come from `halfblock.OpenVideoStream` (an `image.Image` channel plus a stop func). `ubunatic.com/cati v0.2.6` resolves with `go get`, so no local `replace` is needed.
**Record answers:** (1) map `RenderToGrid` cells directly into `Canvas`, with no round trip through ANSI; (2) `background.go` keeps its Braille path, and any shared path would be a separate change.

### M1 — media widget (stills)

**Pre-Work / constraints (host):**
1. Put the widget in its own subpackage (e.g. `media/`), so the core `loom` package does not import cati. Check `go mod graph` for what cati pulls in transitively, and report any cgo or heavy dependencies in the commit message.
2. Video tests must skip cleanly when the external tool that `OpenVideoStream` needs (e.g. ffmpeg) is missing.
3. The stream stop func and goroutines are released by `Close()`, following the pattern from 064/065.
