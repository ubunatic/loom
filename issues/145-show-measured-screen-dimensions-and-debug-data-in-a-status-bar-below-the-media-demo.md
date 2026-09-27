# 145 — Show measured screen dimensions and debug data in a status bar below the media demo

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [143](143-make-the-media-example-full-width-and-play-video.md), [144](144-fix-examples-media-pane-initialization-to-take-full-terminal-size.md)

---

## 1. Problem & Motivation

Users running `examples/media` need visual feedback on current terminal geometry (columns and rows), media rendering mode, media filename/type, and dimension debug info at runtime.

## 2. Technical Specification / Findings

In `examples/media/main.go`, add a bottom status bar (at `r.Y + r.H - 1` with height 1) beneath the media rendering area:
- Media area bounds: `Rect{X: r.X, Y: r.Y + 1, W: r.W, H: r.H - 2}` (preserving 1 row for top title and 1 row for bottom status bar).
- Status bar contents: terminal size (e.g. `Cols: %d  Rows: %d`), media file path / name, rendering mode (`halfblock`, `quadblock`, `sextant`), and status style.
- Update `examples/media/media_pty_test.go` to verify the status bar displays columns/rows and stays on the bottom row.

## 3. Implementation & Verification Plan

- **/goal**: Render a bottom status bar in `examples/media` showing measured terminal width, height, mode, and media debug info, and verify via PTY tests, or stop and report when blocked on a user decision or denied permission.
