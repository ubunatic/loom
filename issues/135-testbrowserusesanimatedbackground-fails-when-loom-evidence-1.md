# 135 — TestBrowserUsesAnimatedBackground fails when LOOM_EVIDENCE=1

**Status**: Closed — test pins LOOM_EVIDENCE off (t.Setenv); make test-q1 green
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: 134

---

## 1. Problem & Motivation
`LOOM_EVIDENCE=1 make test-q1` (used to re-record evidence frames in 134 M3) fails
`examples/filebrowser/filebrowser TestBrowserUsesAnimatedBackground` with "filebrowser pane has no
animated background" (browser_test.go:197). Without the variable the test passes. An evidence run
should not change what the tests check.

## 2. Technical Specification / Findings
Likely the evidence mode turns off or replaces the animated background; the test reads the ambient
environment instead of setting it.

## 3. Implementation & Verification Plan
Make the test set or clear LOOM_EVIDENCE itself (t.Setenv), or make evidence mode keep the
background type; verify with and without the variable.
