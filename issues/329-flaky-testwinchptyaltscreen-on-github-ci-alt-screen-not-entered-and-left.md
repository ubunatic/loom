# Flaky TestWinchPTYAltScreen on GitHub CI: alt screen not entered and left

Status: Closed — PTY Wait drains to EOF before Raw assertions; race count=20 passed locally and in 2-CPU container (866b510).

## Problem
GitHub CI run 39010007435 (commit 2e3fda6) failed in `examples/winch/winch`:
`winch_pty_test.go:190: alt screen not entered and left (on=418 off=-1)`.
The alt-screen leave sequence was never seen. The next run (fe538e9, run
39011071715) passed, and local plus container `-race -count=3` runs pass, so
this is timing-dependent: the PTY reader probably stops before the child
writes the leave sequence under a slow, race-instrumented runner.

## Acceptance
- The test waits for child exit and drains the PTY to EOF before asserting,
  with no wall-clock sleeps.
- `go test -race -count=20 ./examples/winch/...` passes in the CI container
  limited to 2 CPUs.
