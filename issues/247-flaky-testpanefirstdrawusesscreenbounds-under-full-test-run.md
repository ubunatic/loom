# 247 — Flaky TestPaneFirstDrawUsesScreenBounds under full test run

**Status**: Closed — cancellation checked before each frame (41438d8)
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 243, 246

---

/goal The pane PTY first-draw test passes reliably in `make test-q1`, or stop and report when blocked.

## 1. Problem & Motivation
During the 243 M3 / 246 run (85511f4, 8c1ccfd), `make test-q1` failed once in an untouched test:
`TestPaneFirstDrawUsesScreenBounds/full` (`pane_pty_test.go:70`): first draw got `{X:0 Y:0 W:99 H:29}`,
want 100x29. In isolation, `go test -count=5 -run TestPaneFirstDrawUsesScreenBounds .` passes 5/5, so it
looks load- or timing-dependent. A flaky test blocks closing tickets under the one-run Quota-1 budget.

## 2. Technical Specification / Findings
- Suspects: the first draw reads the size before the PTY resize is applied, or a race between the size
  query and SIGWINCH under load. Check whether the off-by-one width is the test's or the pane's.

## 3. Implementation & Verification Plan
- Reproduce under load (`-count=50`, `-race`, or parallel with the full suite); fix the cause, not the
  assertion; then one green `make test-q1`.

## Recurrence (2026-10-04, during issue 261 M1)

Failed 3 runs in a row in different subtests (`inline`, `wrapped-request-auto-alt`): `pane_pty_test.go:70: first draw at row 7 with {X:0 Y:0 W:99 H:24}, want row 7 and 100x24`. M1 touched only RichTextEdit.

## Delivered (2026-10-06)

- `8a9ee13` M1 red test (test-only hook `paneBeforeSignalHandler`, nil in production); `b45cb88` M2: `refreshStartupSize()` re-reads the size after signal registration and the ZWJ probe, before the first canvas, in every screen mode; `41438d8` (247): the run loop checks `ctx.Err()` before each frame, so a SIGWINCH that wins `select` after cancellation can no longer redraw the first frame at the guarded width (99 instead of 100).
- Developer dev247 (codex:sol:med). Full `make test-q1` after the last commit: 0 FAIL (`/tmp/loom-final-test-q1.log`, 15:54). Host rerun was blocked by Quota-1 (no source change since); host ran the pane startup tests 20 times: all green. `make install` done.
