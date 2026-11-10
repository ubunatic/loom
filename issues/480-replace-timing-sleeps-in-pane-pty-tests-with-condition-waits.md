# 480 — Replace timing sleeps in pane PTY tests with condition waits

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Testing
**Related**: 458, 337, 377

---

## 1. Problem & Motivation
`docs/TestTiers.md:44` forbids timing sleeps. 458 batch 9 counted 119 sleep calls in repo tests; a
sample of five were all timing waits: `pane_kitty_test.go:151`, `pane_suspend_test.go:133`, `:298`,
`:418`, `pane_pty_test.go:323`. Closed 377 removed the session sleeps, not the pane ones. Not every
one of the 119 is a violation (child-process scripts may need them).

## 2. Technical Specification / Findings
- Classify each sleep: timing wait (replace with a condition wait with a deadline, e.g. a ptytest
  wait or a channel) or allowed (document why at the call).

## 3. Implementation & Verification Plan
/goal No timing sleep remains in the pane tests, each remaining sleep has a one-line reason, and the
tests pass under `-race -count=20`; stop and report when blocked on a decision or denied permission; on a
loop or no progress, clear the goal and write a failure report.

- Re-check live code first: `rg -n 'time\.Sleep' --glob '*_test.go'`.
- `make test-q1`.
