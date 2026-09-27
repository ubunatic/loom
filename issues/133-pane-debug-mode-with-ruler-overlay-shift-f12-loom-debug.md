# 133 — Pane debug mode with ruler overlay (Shift-F12, LOOM_DEBUG)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Enhancement
**Related**: 132, 048, 089

---

## 1. Problem & Motivation
Debugging a Loom app's layout means guessing which cells a widget paints. ansiviewer's `r` ruler
(81d5d6b) showed that a 1-cell pad on the terminal's default background makes unpainted cells and
off-by-one widths visible at a glance. Every Loom app should get this for free, as a pane debug mode
that can later show more hints (box outlines, focus, mouse position, redraw counts).

## 2. Technical Specification / Findings
- Toggle: Shift-F12 in any pane, and `LOOM_DEBUG=1` to start with it on (for terminals that
  capture Shift-F12). Loom's key decoder has no `shift-f12` yet (event.go lists only f1..f12), so
  decoding `CSI 24;2~` is part of the work.
- The pane shrinks the app's rect by 1 on each side while debug is on, and shifts mouse
  coordinates by the same offset, so widgets still receive 0-based child-local positions
  (docs/Widgets.md). A resize is sent to the app on toggle.
- The ruler reuses ansiviewer's `drawRuler` look: dim dots, `┊` every 5, an orange digit every 10;
  row digits on the left. Cells use `PaintForeground` with a reset BG, which keeps the terminal
  default (canvas change in 81d5d6b).
- Off by default; no change to apps that never toggle it. ansiviewer keeps its own `r` preview ruler.

## 3. Implementation & Verification Plan
- M1: `shift-f12` key decoding with a test (and F-key modifier variants if cheap).
- M2: pane debug toggle, rect shrink, mouse offset, env var; tests for the offset of mouse events
  and for the ruler cells' default BG; move `drawRuler` into Loom and use it from ansiviewer.
- M3: a PTY test that toggles debug mode, clicks a widget, and checks the click still lands.
- Manual check (in 102): toggle in a few examples in tilix and foot.
