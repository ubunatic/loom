# 217 — ProgressBar demo: fill 0 to 100, fade, repeat

**Status**: Closed — fixed with tests, suite green
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Enhancement
**Related**: 207

---

## 1. Problem & Motivation
User feedback: the demo bar moves left and right; progress should read as 0 -> 10 -> ... -> 100 -> fade -> repeat.

## 2. Technical Specification / Findings
Keep the widget's start/end capability; change only the demo animation to fill from 0 to 100, fade out briefly, and restart.

## 3. Implementation & Verification Plan
Test the demo sequence: values increase monotonically to 100, then fade, then restart at 0.

/goal The ProgressBar demo fills 0-100, fades and repeats, proven by a test; or stop and report when blocked on a user decision or denied permission.
