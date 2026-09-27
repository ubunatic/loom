# 146 — Set pane.MaxCols = 0 in examples/media to uncap terminal width

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [143](143-make-the-media-example-full-width-and-play-video.md), [144](144-fix-examples-media-pane-initialization-to-take-full-terminal-size.md), [145](145-show-measured-screen-dimensions-and-debug-data-in-a-status-bar-below-the-media-demo.md)

---

## 1. Problem & Motivation

Loom's default pane configuration sets `MaxCols = SpeccedDefaults.Pane.MaxCols` (which defaults to 50 columns in `spec/defaults.yaml`). Full-screen apps like `filebrowser`, `ansiviewer`, `ansiedit`, `textedit` explicitly set `pane.MaxCols = 0` to use the actual terminal width. `examples/media` did not set `pane.MaxCols = 0`, causing the canvas to be hard-capped at 50 columns regardless of terminal width.

## 2. Technical Specification / Findings

In `examples/media/main.go`, set `pane.MaxCols = 0` after `loom.New(1 << 16)`:
```go
pane, err := loom.New(1 << 16)
if err != nil {
    return err
}
pane.MaxCols = 0
```
Update `examples/media/media_pty_test.go` to verify that `Cols` matches wide terminal dimensions (> 50, e.g. 120 cols).

## 3. Implementation & Verification Plan

- **/goal**: Set `pane.MaxCols = 0` in `examples/media` and verify with wide PTY tests, or stop and report when blocked on a user decision or denied permission.
