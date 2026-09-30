# 187 — Add Timer and Stopwatch widgets

**Status**: Closed
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: widget.go (Ticker), issues/154, issues/177, issues/180

---

## 1. Problem & Motivation
Countdown and elapsed-time displays are common in inline tools; Bubbles ships Timer and Stopwatch (issue 177). Loom has only the raw `Ticker`.

## 2. Technical Specification / Findings
`Timer` (countdown with `OnDone` callback) and `Stopwatch` (elapsed), both `Start`/`Stop`/`Reset`, formatted `mm:ss` or custom; tick once per display-change interval via `Ticker`. Inject the clock for tests.

## 3. Implementation & Verification Plan
/goal Ship `Timer` and `Stopwatch` with fake-clock tests and a docs/Widgets.md row; stop and report if 154's cadence control is missing and needed.
