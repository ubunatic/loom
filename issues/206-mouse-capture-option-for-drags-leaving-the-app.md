# 206 — Mouse capture option for drags leaving the app

**Status**: Closed
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 195, 197-202 gallery

---

## 1. Problem & Motivation
User feedback (FilePicker): dragging the scrollbar keeps working when the pointer leaves to the left, but stops when it leaves the app to the right. If the terminal allows it, add an option to keep capturing the drag outside the app area.

## 2. Technical Specification / Findings
Investigate first (Canary): which terminal mouse modes (1002 button-event, 1003 any-event, 1006 SGR, 1016 pixel) report motion outside the app cell area, and whether the right-edge stop is a loom clamp (coordinates beyond width dropped) or the terminal. If loom drops out-of-range events, keep delivering them to the widget holding the drag, behind an opt-in.

## 3. Implementation & Verification Plan
PTY test: press on a scrollbar, drag with x beyond the app width, the scroll keeps following.

/goal Drags continue outside the app where the terminal reports them, or document why not; or stop and report when blocked on a user decision.

## 4. Result (2026-09-30)
Frame kept dropping out-of-pane drag/release reports; it now keeps routing them to the pressed child (67a493d), proven by a PTY test with injected out-of-range SGR reports. Not verified in a real terminal (no live pointer in the agent PTY): check with `loom widgets --show FilePicker`, dragging the scrollbar out to the right, and `loom info --watch` to see what the terminal reports.
