# 444 — PTY data race and live tmux timeout flakiness

**Status**: Closed — Thread-safe ptytest Session accessors eliminate data race; live tmux timeout increased; verified with go test -race and full uncached go test ./...
**Closed-By**: session abe69541-ef92-47e9-b3da-b8dbf42f4cde
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: issues/438, issues/399

---

## 1. Problem & Motivation

Full uncached and race-instrumented test runs (`go test -race` and `go test -count=1 ./...`) fail with two concrete issues:
1. Data race in `cmd/loom/widgets_paste_pty_test.go` (`TestTextEntryGalleryCtrlUPTY`): test predicates call `gallery.VT().Cursor()` directly without mutex synchronization while `Session.pump()` concurrently writes PTY output to the virtual terminal (`vt.Write()`). `internal/ptytest.Session` lacks thread-safe accessor methods for cursor position, mouse modes, and alternate screen state.
2. Flaky timeout in `cmd/loom/live_tmux_test.go` (`TestLiveTmuxScriptRunsStepsAndMeasuresPause`): test hardcodes `--timeout 2` for `scripts/live-tmux.sh`. Under full uncached multi-package concurrency, spinning up an isolated tmux server, subshell, and child process takes ~2.6s, causing intermittent test failures.

## 2. Technical Specification / Findings

- `internal/ptytest/session.go`:
  - `Session` currently synchronizes `Screen()` and `Cells()` via `s.mu.Lock()`, but only provides `s.VT()` returning the raw `*VT` pointer.
  - Add thread-safe methods to `Session`:
    - `Cursor() (int, int)`
    - `MouseMode() int`
    - `MouseSGR() bool`
    - `AltScreen() bool`
- Update callers in:
  - `cmd/loom/widgets_paste_pty_test.go`: replace `gallery.VT().Cursor()` with `gallery.Cursor()`.
  - `internal/session/session_pty_test.go`: replace unprotected `session.VT().Cursor()`, `session.VT().MouseMode`, `session.VT().MouseSGR`, and `session.VT().AltScreen()` with thread-safe `Session` methods.
  - `internal/editor/edit_280_pty_test.go`: replace `s.VT().MouseMode` with `s.MouseMode()`.
- `cmd/loom/live_tmux_test.go`:
  - Increase the initial command wait timeout from 2s to 10s (scaled under load) so parallel execution does not trip on server spin-up.

## 3. Implementation & Verification Plan

- Developer: `agy:sonnet:low` (per owner instruction). Host orchestrator performs direct review.
- Verification:
  - `go test -race -count=1 ./cmd/loom` must pass with 0 data races.
  - `go test -count=1 ./...` must pass with 0 timeouts.
  - `go vet ./...` and `make test-q1` must pass.
  - Run `make install`.
