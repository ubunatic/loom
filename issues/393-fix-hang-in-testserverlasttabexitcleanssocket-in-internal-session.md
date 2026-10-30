# 393 — Fix hang in TestServerLastTabExitCleansSocket in internal/session

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [287](287-fix-search-overlay-mouse-routing-and-selection-isolation-in-loom-edit.md) (its final `make test-q1` gate is blocked by this test)

---

## 1. Problem & Motivation
`TestServerLastTabExitCleansSocket` in `internal/session/server_test.go:161` never returns. `make test-q1` fails on it with `panic: test timed out after 10m0s`, so the full suite cannot pass. The test covers the rule that the session socket is removed when the last tab exits. Its run also blocks every ticket that needs a green full suite before closing.

Reproduced on 2026-10-30 with `go test -count=1 -timeout 90s -run 'TestServerLastTabExitCleansSocket$' ./internal/session/`: the test panics after 1m30s. The package is not touched by the editor work in issue 287.

## 2. Technical Specification / Findings
- The test sets `LOOM_SOCKET_DIR` to a temp dir and `LOOM_SESSION_SERVER_NAME=finished`, so it does not share sockets with a live `loom session`. A live server is an unlikely cause.
- The stack at the timeout shows the test blocked in `RunWithOptions` → `serveSession` (`internal/session/server.go:209`, the `select` loop). The session goroutine is in `runSessionWithServer` → `runSessionCore`. Find out why the run goroutine does not finish and signal `runDone`, or why the accept loop does not see it.
- The test command is `bash -c "exit 0"` with an empty stdin. Check whether the session core waits on input or on the last client's detach.
- Introduced or first seen before 2026-10-30; the file history is `f2146fb`, `47389a2`, `d544475`.

## 3. Implementation & Verification Plan
1. Reproduce with `-timeout 90s` and capture the goroutine dump (the lines above).
2. Fix the shutdown path so `runDone` is observed when the command exits, without a sleep in the test.
3. Verify: `go test -count=1 -run 'TestServerLastTabExitCleansSocket$' ./internal/session/` passes, then `make test-q1` passes in full.

/goal `TestServerLastTabExitCleansSocket` returns, the session socket is removed after the last tab exits, and `make test-q1` is green. Stop and report when blocked.
