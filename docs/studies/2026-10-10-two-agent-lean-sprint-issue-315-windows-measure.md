# 2026-10-10 Lean Sprint Study: Windows Builds in Terminal Measurement (Issue 315)

## Summary
- **Ticket**: 315 (Support Windows builds in terminal measurement)
- **Role**: Host Orchestrator (Antigravity), Developer: `luna:high` (`codex:luna:high`)
- **Status**: In Progress

## Initial Findings
- Downstream build failures on Windows amd64 in `measure/zwj.go` due to Unix-specific polling (`unix.PollFd`, `unix.Poll`, `unix.Read`, `unix.POLLIN`, `unix.EINTR`).
- Probe CLI tools `cmd/loom-probe` and `cmd/loom-repaint-probe` also fail compilation on Windows (`unix.Poll` and `syscall.SIGWINCH`).
- Baseline linux test suite passes completely clean (`make test-q1`).

## Milestones
- **M1**: Separate Unix and Windows implementations of cursor reading / terminal probe in `measure` package so `measure` compiles cleanly on Windows.
- **M2**: Resolve Windows compile issues in `cmd/loom-probe` and `cmd/loom-repaint-probe` (or guard them with build tags / platform-appropriate fallbacks) and verify `GOOS=windows GOARCH=amd64 go build ./...`.
- **M3**: Run test suite (`make test-q1`), add/verify Windows cross-compile checks, and perform teardown/closing.
