# 430 — loom session: 6 tab PTY tests time out on v0.10.0 (TestSessionNewTabCtrlN and others)

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: 429 (found during Jules wave 1), 392f0da (q1 last reported green)

---

/goal The six session tab PTY tests pass again on main, with the root cause named (code, test or machine change); stop and report when blocked on a user decision.

## 1. Problem
`make test-q1` on 2026-11-06 fails in `internal/session` (log: `.git/harnez/quota-1-logs/test-20261106T002200.813372938Z.log`):
TestSessionNewTabCtrlN, TestSessionTabSwitchPrefixNAndP, TestSessionTabSwitchPrefixSelect,
TestSessionPrefixNewTabAndCloseTab, TestSessionAltScreenTabSwitch, TestSessionDirectTabSelectAltDigits, each after 5.04s.

## 2. Findings
- Reproduces on the v0.10.0 release commit 7511513, so it is not caused by the wave 1 merges (421, 414).
- Reproduces with a clean environment (`env -i HOME PATH TERM=xterm-256color SHELL=/bin/bash`).
- 392f0da (one commit before v0.10.0) reported q1 green, so something outside the code may have changed (machine, shell, bash version), unverified.
- The screen dump shows the new tab's output (`$ TAB2_RUNNING`) but `WaitFor` times out at `internal/session/session_pty_test.go:902`; the tab bar reads `1: bash  2: sh`.

## 3. Next step
Bisect between 392f0da and 7511513 if the tree differs; otherwise compare the shell the new tab starts (`sh` vs `bash`) and what `WaitFor` matches.
