# 474 — session client that stops reading blocks every tab of its session

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 458, 381

---

## 1. Problem & Motivation
Code read in 458 batch 7 (`docs/studies/tuios/report-session.md`), not yet canaried. Each tab's
PTY reader holds `sessionMu` (`internal/session/session.go:758`) while it writes the frame to the
client (`:783` → `sessionServer.writeOutput`, `server.go:88-98` → `writeMessage`, `ipc.go:217`).
That write has no deadline and also holds `writeMu`. If the client stops reading (frozen SSH link,
stopped client process), the socket buffer fills, the write blocks, and every other tab's reader
then waits for `sessionMu`, so all child programs of the session stall on output. `detach` needs
`writeMu` too (`server.go:142-143`), so a new client may not take over. Other sessions run in
their own processes and are not affected.

## 2. Technical Specification / Findings
- Canary first: an isolated session over a Unix socket whose client stops reading; measure
  whether other tabs' PTYs stop draining and whether detach or a new attach works.
- Fix direction if confirmed: a write deadline or a bounded per-client queue written outside
  `sessionMu`; on overflow drop the client (it repaints on reattach).

## 3. Implementation & Verification Plan
/goal Confirm or disprove the stall with a canary, and if confirmed, a client that stops reading
no longer blocks the session's tabs or a later attach, with tests and no sleeps; stop and report
when blocked on a design decision or denied permission; on a loop or no progress, clear the goal
and write a failure report.

- Re-check live code first; the batch 7 flow note saw the same lock order deadlock over
  `net.Pipe` in-process tests.

### Reproduced (2026-11-11, Jules test-draft GH #131, host rerun)
- `TestBackpressure` (`internal/session/backpressure_test.go` on the PR branch) attaches a
  `net.Pipe` client that never reads while the child loops `echo`; `server.detach(true)`
  does not return within 2 s x RaceFactor. Host rerun: `FAIL ... server.detach() blocked`.
- Chain: `writeOutput` holds `writeMu` during `writeMessage(conn, ...)`
  (`internal/session/server.go:89`, `:97`); `detach` takes `writeMu` first (`:143`) and,
  with `notify`, writes `msgDetached` to the same stuck conn (`:154`). A fix must bound that
  write too, or close the conn before notifying.
- Lead from the failed Jules run 16549792241857724638: a write deadline plus `go detach`
  broke the deadlock but made `TestSessionCtrlRightBracketPassesThrough` and
  `TestSessionFailedScreenshotReportedOnStderr` fail; check those first.
- Test notes: `net.Pipe` has no buffer, so it blocks on the first write; add one Unix-socket
  case that fills the kernel buffer. The test's `t.Cleanup` waits 5 s x RaceFactor on a
  stuck server; after the fix it should end at once. Take the test from GH #131, then close it.
