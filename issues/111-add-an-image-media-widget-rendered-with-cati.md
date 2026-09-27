# 111 — Add an Image/Media widget rendered with cati

**Status**: Closed — media.Widget via cati v0.2.6 (halfblock/quadblock/sextant, video on Ticker, aspect fit, cache, PTY rect check); visual check parked in 102 (8c06a7d, 6a65180, b3c6e71, 7ef3c86)
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

**M1 delivered (8c06a7d):** `media.Widget` supports halfblock, quadblock and sextant, maps `RenderToGrid` cells into the Canvas, and core loom has no cati import. **M2 (6a65180):** `NewVideo` advances on Ticker to the newest frame, stops at end of stream, and has an idempotent `Close`; its test skips when ffmpeg is missing. **M3 (b3c6e71):** `examples/media` with a PTY check that truecolor cells stay inside the rect; it is included in `make install`. Dependencies: only `ubunatic.com/cati v0.2.6` was added (no cgo). At runtime, video needs ffmpeg/ffprobe and SVG needs rsvg-convert.

### M4 — aspect, caching, robustness

**Pre-Work / Required Refinements (from M1-M3 review):**
1. **Aspect ratio is untested.** "Done when" requires fitting with the aspect ratio preserved. Add a test where a wide image drawn into a tall rect (and a tall image into a wide rect) keeps its proportions, within one cell, and is centered. Check whether cati's `RenderToGrid(img, cols, Options{Rows})` stretches when both are set, and compute the fitted cols and rows yourself if it does.
2. **Clear letterbox areas.** Cells in the rect that the image does not cover must be reset (blank, Reset style), so a new frame or size leaves no stale pixels.
3. **Do not rescale on every Draw.** `Draw` calls `RenderToGrid` on every repaint. Cache the grid, keyed by the image identity and the rect size, and re-render only when the frame or size changes.
4. **Errors are swallowed.** When rendering fails, draw a short error message inside the rect instead of leaving it blank.
5. The outside-rect assertion in `widget_test.go` only checks `Text != " "`. Also compare `Style` against Reset, so a background-only spill is caught.

**M4 delivered (7ef3c86):** centered aspect fit (cati already preserves the aspect ratio), letterbox clearing, a grid cache keyed by the image and rect size, render errors drawn inside the rect, and spill checks that also compare Style. The host re-ran `make test-q1`, and it passed, including `media` and `examples/media`.
**Upstream note:** `NewVideo` drains cati's frame channel before calling its stop func, because cati's `OpenVideoStream` reader and its cleanup both call `cmd.Wait`, and the M4 Close test hung. This workaround belongs in cati (v0.2.6), not in loom.
**Visual check:** the user was away. The check is parked in 102.
