# 306 — Pane Post to run widget mutations on the event loop from any goroutine

**Status**: Open
**Priority**: P1 (High)
**Severity**: Feature
**Category**: Pane / Event Loop
**Related**: harnez #737, #739, #741

---

## 1. Problem & Motivation
Streamed model output (harnez #739) and job logs (harnez #741) arrive on
background goroutines. `Pane.Invalidate` only requests a repaint; changing widget
state from the producer races with `Draw`. Today a caller must add its own
mutex or poll a channel from `RunWatch`'s collect callback.

## 2. Technical Specification / Findings
- Add `Pane.Post(fn func())`: safe from any goroutine; queues `fn` to run on the
  event-loop goroutine before the next frame, then repaints.
- Never blocks the producer: bounded queue; when full, fall back to an unbounded
  slice under a mutex (dropping callbacks would lose output).
- Run all queued callbacks in order, then draw once (coalesce bursts).
- Calls after `Close` are no-ops.

## 3. Implementation & Verification Plan
- Unit test with `-race`: 1000 posts from many goroutines append to a widget;
  all arrive in per-goroutine order and no race is reported.
- Test that Post before `Run` starts is delivered on the first loop iteration.
- Document Post vs Invalidate in `docs/Widgets.md`.
