# 486 — session PTY tests time out under go test -race on a busy host

**Status**: Backlog
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: 474, 480

---

## 1. Problem & Motivation
On 2026-11-11, `go test -race ./internal/session/` failed twice on main cf5177dc with
`timed out waiting for screen state` in `TestSessionFailedScreenshotReportedOnStderr` and
`TestSessionStderrCleanLineStart` (`internal/session/session_pty_test.go:836`, `:870`), while
Jules runs and other tests were active on the host. The same tests passed when run alone without
`-race` and in two `make test-q1` runs the same night. Seen in
`docs/studies/2026-11-11-tuios-research-and-jules-sweep-story.md`.

## 2. Technical Specification / Findings
- Both tests use `session.WaitFor(..., time.Second*5)` with fixed budgets, not scaled by
  `ptytest.RaceFactor`; other tests in the same file scale it (`session_pty_test.go:56`).
- Not reproduced on demand yet; the failing run was the package alone with `-race`.

## 3. Implementation & Verification Plan
/goal Find out whether the two tests fail from a fixed timeout under race or from a real order
problem, and make them pass reliably under `-race` without sleeps; stop and report when blocked on
a design decision or denied permission; on a loop or no progress, clear the goal and write a
failure report.

- Try `go test -race -count=20 -run 'TestSessionFailedScreenshot|TestSessionStderrCleanLine'
  ./internal/session/` under load first; scale budgets by `RaceFactor` per docs/TestTiers.md.
