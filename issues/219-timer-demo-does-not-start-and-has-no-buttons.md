# 219 — Timer demo does not start and has no buttons

**Status**: Closed — fixed with tests, suite green
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: 207

---

## 1. Problem & Motivation
User feedback: the Timer demo stays at 00:00, has no controls and never starts.

## 2. Technical Specification / Findings
Timer demo must run and offer start/stop/reset controls (mouse and keys) like the Stopwatch demo from 207. Check whether ticks reach the widget (see 211 hook discovery).

## 3. Implementation & Verification Plan
PTY test: the timer changes value over time; the buttons start, stop and reset it.

/goal The Timer demo runs and has working controls, proven by PTY tests; or stop and report when blocked on a user decision or denied permission.
