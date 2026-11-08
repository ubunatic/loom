# 451 — Accelerate fast mouse-wheel scrolling in the core event system

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 449, 450, 452

---

## 1. Problem & Motivation
Owner report (Viewport demo, foot): scrolling with the wheel feels slow. `View` and `Viewport` move one line per wheel report (`view.go`, `viewport.go`), so a long scroll needs many notches. Fix it once in the core event system, not per widget.

## 2. Technical Specification / Findings
- Unverified: what a wheel gesture looks like at the terminal (reports per notch, gaps between reports, differences between mice, touchpads and terminals). `loom-probe mouse` now shows wheel up/down counts, the last and smallest gap between reports, and the current and longest burst (gap under 150 ms). Measure first (docs/Canary.md), record the numbers here, then pick the curve.
- Design direction: `Pane` measures the gap between consecutive wheel reports in the same direction and delivers the report to the widget N times, N growing with the speed (1 for a slow notch). Widgets stay unchanged, so every scroller (View, Viewport, Tree, lists, TextEntry) gains it. A direction change or a pause resets it.
- The curve (burst gap, max gain, ramp) lives in `spec/defaults.yaml`, not in Go (docs/Spec.md). The values come from the measurements; do not guess them.
- Time comes from an injectable clock on `Pane` so tests need no sleeps (docs/TestTiers.md).
- Mouse-path only (wheel has no key twin); document the keys that already scroll by page.

- Measurement 1 (owner, foot, `loom-probe mouse`, burst = reports with gaps under 150 ms): slow scrolling climbs from burst 3 to 21 in steps; medium scrolling reaches 21 quickly (sometimes above); fast scrolling reaches 21 at once and often 24 or 30. Reading: the steps of 3 suggest foot sends 3 wheel reports per notch, so each notch already moves 3 lines. Unverified: the gap values (last gap, min gap) were not reported, so the gap inside a notch versus between notches is still unknown. A notch of 3 reports must count as one step in the curve, not as a fast scroll.

## 3. Implementation & Verification Plan
- Run `loom-probe mouse` in foot, kitty, ghostty (and ptyxis if used); record gap and burst numbers in this ticket.
- Tests: a slow notch scrolls 1 line; a fast burst scrolls more than the same count of slow notches; a pause and a direction change reset the gain; the cap holds; the clock is injected.
- Docs: `docs/TuiInput.md` (wheel behavior) and `docs/Widgets.md`; check in `loom widgets --show Viewport` (100 lines).
