# 337 — Stabilize PTY race tests with screen-state waits

**Status**: Closed — resolved in a1b4372
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**:

---

## 1. Problem & Motivation
Race-instrumented PTY tests use fixed sleeps or short polling deadlines, so screen
updates and child output can arrive after assertions on slower two-CPU CI jobs.
The PTY drain test also writes more data than needed to exercise EOF draining.

## 2. Technical Specification / Findings
`ptytest.Session` already renders PTY output through the public `loomtest.VT`.
Waiters should wake when the VT consumes new output and inspect its cell grid;
waits after input can require a newer output revision. Race deadlines use the
shared `ptytest.RaceFactor` build-tag constant.

## 3. Implementation & Verification Plan
The drain test sends five 64 KiB chunks and retains its final `END` assertion.
The editor and filebrowser PTY tests wait for screen state and inspect cells or
visible screen text instead of ANSI strings and fixed sleeps.

Verification: `go test ./...`, `make container-test CONTAINER_CPUS=2`, and the
three requested PTY targets with `go test -race -count=20` in a two-CPU
container.
