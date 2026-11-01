# 394 — Cover WaitForAttach wiring in serveSession and remove sleep in control status test

**Status**: Closed — serveSession WaitForAttach test added (111ec99); fails when server.go wiring removed; make test-q1 green
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Test
**Related**: [393](393-fix-hang-in-testserverlasttabexitcleanssocket-in-internal-session.md) (introduced `Options.WaitForAttach`)

---

## 1. Problem & Motivation
Ticket 393 moved the attach gate from `LOOM_SESSION_WAIT_ATTACH` into `Options.WaitForAttach`. Two gaps remain, found in its review:
- `TestServeSessionWaitForAttachHoldsChildUntilAttach` drives `runSessionWithServer` with `server.waitForClient` set directly. Nothing proves that `serveSession` copies `options.WaitForAttach` into `server.waitForClient` (`internal/session/server.go:181`). A regression there would pass the suite.
- `TestServerControlStatusListDetachAndKill` still waits with `time.Sleep(20 * time.Millisecond)` (`internal/session/server_test.go:246`). The no-sleep rule in `docs/TestTiers.md` applies, and this test predates ticket 393.

## 2. Technical Specification / Findings
- The wiring gap needs a test that calls `serveSession` with `Options{WaitForAttach: true}` and checks the gate, with no env variable set.
- The sleep is in an untouched test. Replace it with a channel or an existing helper, without weakening its assertions.
- Found in review of commits `4a19821` and `67f922a`.

## 3. Implementation & Verification Plan
1. Add a test that calls `serveSession` with `Options{WaitForAttach: true}` and asserts the child does not start before attach, using a channel or event rather than a sleep.
2. Remove the sleep at `server_test.go:246` and keep its assertions.
3. Verify with `go test -count=1 ./internal/session/` with `LOOM_SESSION_WAIT_ATTACH` both set and unset, then `make test-q1`.
