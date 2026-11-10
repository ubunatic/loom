# 459 — Canvas flush writes only changed cells, measured against Bubble Tea v2

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Performance
**Related**: 457, 458

---

## 1. Problem & Motivation
On every dirty redraw, `Canvas.FlushWithConfig` (`canvas.go`) serialises all rows, even when one
cell changed. The 458 batch 1 canary (80×24, one run, `docs/studies/tuios/report-render.md`)
measured loom Canvas and Pane at 2269 bytes for a one-cell change, 2301 for a bold change and 2270
for a wide cell; Bubble Tea v2.1.1 (ultraviolet renderer) sent 9, 11 and 25 bytes. Idle panes
already write nothing (redraw is dirty-gated in `pane_run.go`). The cost matters most over SSH,
in `loom session` and for frequent small updates (cursor blink, spinners, typing).

## 2. Technical Specification / Findings
- ultraviolet keeps the previous screen and compares cells per touched line (`paintLine`, cell
  comparison in `terminal_renderer.go`); its line hash is only a fullscreen scroll optimisation and
  is not needed here.
- Keep what loom has: SGR only on style change within a row, one writer call per frame
  (`AtomicFlush`), optional mode 2026 wrap, stale-row clearing on shrink.
- Any full invalidation (resize, suspend/resume, alternate-screen switch, inline origin change,
  explicit `Invalidate`) must force a full repaint, because the terminal state is then unknown.
- Wide cells and continuation cells must be compared and rewritten as a unit.
- Not yet measured: Pane write-call count, playback of the stream in a terminal model.

## 3. Implementation & Verification Plan
/goal Canvas keeps the last flushed frame and a dirty flush writes only changed cells or runs,
with a full repaint after every invalidation, verified by byte counts and stream playback; stop and
report when blocked on a design decision or denied permission, or when no progress is made (clear
the goal and write a failure report).

- Re-check live code first (`canvas.go`, `pane_run.go`, `pane_screen.go`), the code may have moved.
- Tests: rerun the batch 1 matrix as a Go test with a counting writer; play each stream into
  `loomtest.VT` and compare the screen with a full-repaint stream for the same frames (blank,
  unchanged, one cell, style only, wide cell, shrink, resize, resume).
- Expected: one-cell change well under 100 bytes; screens identical to full repaint.
