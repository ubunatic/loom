# 186 — Add a standalone Spinner widget

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: splash_view.go, graph/spinner.go, issues/154, issues/177, issues/180

---

## 1. Problem & Motivation
Only `SplashView` draws a spinner; apps that want a small busy indicator next to a label must build one (issue 177).

## 2. Technical Specification / Findings
`Spinner` widget: frame set from spec (reuse `graph/spinner.go` frames), optional label, `Start`/`Stop`, animates through `Ticker` (with 154's cadence control, stops ticking when stopped). `SplashView` may reuse it.

## 3. Implementation & Verification Plan
/goal Ship `Spinner` with frame-advance and stop tests and a docs/Widgets.md row; stop and report if 154 is not done and ticking cannot pause cleanly.
