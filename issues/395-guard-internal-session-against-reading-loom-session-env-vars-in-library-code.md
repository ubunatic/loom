# 395 — Guard internal/session against reading LOOM_SESSION env vars in library code

**Status**: Closed — lint_test guards LOOM_SESSION_* reads via Getenv/LookupEnv (752fe35); probe injection caught; make test-q1 green
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Test
**Related**: [393](393-fix-hang-in-testserverlasttabexitcleanssocket-in-internal-session.md) (the hang this guards against), [394](394-cover-waitforattach-wiring-in-servesession-and-remove-sleep-in-control-status-test.md)

---

## 1. Problem & Motivation
Ticket 393 was caused by `serveSession` reading `LOOM_SESSION_WAIT_ATTACH` from the environment. Every shell inside a loom session inherits that variable, so the library behaved differently for tests that ran there. The fix moved the gate into `Options.WaitForAttach`. Nothing stops the same kind of read from coming back, and the existing review found it only because the hang was reproduced.

## 2. Technical Specification / Findings
- The library is `internal/session`. `cmd/loom` may read env vars and pass them as options.
- 393 removed the `LOOM_SESSION_WAIT_ATTACH` read from `serveSession`. Other library reads of `LOOM_SESSION_*` exist, such as `LOOM_SESSION_SERVER_NAME` (`session.go:81–82`), which the library consumes and unsets. A guard must allow those explicitly, and this ticket should list them before it enforces anything. (Not audited in full here.)
- Options: a `go vet`-style check, a test that scans `internal/session` for `os.Getenv("LOOM_SESSION_` outside an allow list, or a `make lint` rule.

## 3. Implementation & Verification Plan
1. Decide the mechanism. A test that scans the package source is simplest and runs in `make test-short`.
2. Add an allow list with the existing `LOOM_SESSION_SERVER_NAME` read and a comment that explains it.
3. Verify the guard fails when `os.Getenv("LOOM_SESSION_WAIT_ATTACH")` is added to `server.go`, then passes on the current tree.
4. Verify with `make test-short` and `make test-q1`.
