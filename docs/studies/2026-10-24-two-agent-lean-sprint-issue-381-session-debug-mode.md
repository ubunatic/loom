# Two-Agent Lean Sprint Study: Issue #381 Session Debug Mode & Self-Healing Diagnostics

**Date**: 2026-10-24
**Issue**: [#381](file:///home/uwe/projects/loom/issues/381-add-session-debug-mode-detect-ansi-leakage-rotating-diagnostics-auto-sigwinch-repair-and-status-indicator.md)
**Roles**:
- Host Orchestrator: Antigravity
- Architect Refiner: `argon:low` (Session: `refiner-381`)
- Developer: `sol:med` (Session: `dev-381`)

---

## 1. Refinement & Architecture
Architect refinement completed in 2m45s (1.2M tokens processed). Key architectural decisions:
1. **CLI & Entrypoint**: Add `--debug` and `--debug-dir` to `cmd/loom/session.go`, paired with `session.Options` and `session.RunWithOptions` in `internal/session/session.go` to keep `Run(...)` backward-compatible.
2. **Ring Buffer**: 64 KB circular byte buffer per tab in `internal/session/debug.go`.
3. **Detector**: Multi-condition detector covering VT safety-cap discards, stalled incomplete escape sequences, and stray control/CSI fragments in newly rendered cells.
4. **Two-Tier Auto-Repair**: Tier 1 `t.vt.ResetParser()`, followed by Tier 2 momentary `TIOCSWINSZ` toggle to trigger child `SIGWINCH`.
5. **Diagnostics & Retention**: Correlated bundles (`.log`, `-before.ansi`, `-after.ansi`) with FIFO rotation capped at 5 bundles and a 10s cooldown circuit breaker.
6. **Status Bar Indicator**: Warning glyph `[!]` in `sessionBarContent` during active alerts, preserving 100% byte-identity when no alert is active.

---

## 2. Sprint Milestones & Verification
- **M1 (Debug options, CLI flags, 64 KB ring buffer & detector)**: Delivered in commit [`9bda066`](file:///home/uwe/projects/loom/cmd/loom/session.go).
  - Added `Options` and `RunWithOptions` to `internal/session/session.go`.
  - Added `--debug` and `--debug-dir` CLI flags and env support (`LOOM_SESSION_DEBUG`, `LOOM_DEBUG_DIR`) in `cmd/loom/session.go`.
  - Implemented 64 KB circular buffer and multi-rule leakage detector in `internal/session/debug.go`.
  - Verified with unit tests in `internal/session/debug_test.go`.
- **M2 (Rotating diagnostics, two-tier auto-repair & status indicator)**: Delivered in commit [`b14fd1d`](file:///home/uwe/projects/loom/internal/session/debug.go).
  - Implemented `writeDebugBundle` writing `.log` with raw stream chunks and `-before.ansi` / `-after.ansi` captures, capped at 5 bundles via FIFO rotation.
  - Implemented two-tier auto-repair in `runSession`: `t.vt.ResetParser()` followed by momentary `TIOCSWINSZ` toggle delivering `SIGWINCH`, with a 10s circuit breaker cooldown.
  - Added status bar warning indicator `[!]` in `sessionBarContent` during active alerts, maintaining 100% byte-identity when no alert is active (`TestSessionBarGoldens` passed).
- **M3 (Automated PTY canary tests & documentation)**: Delivered in commit [`36bf560`](file:///home/uwe/projects/loom/internal/session/session_pty_test.go).
  - Added PTY canary integration tests (`TestSessionDebugDetectsLeakRepairsViaSIGWINCHAndWritesDiagnostics` and `TestSessionDebugDiagnosticRotationCapsAtFiveBundles`) in `internal/session/session_pty_test.go`.
  - Verified detection, `.log` dump, `-before.ansi` / `-after.ansi` creation, `SIGWINCH` delivery, `[!]` warning indicator, 10s cooldown suppression, and max-5 bundle rotation.
  - Updated [`docs/Session.md`](file:///home/uwe/projects/loom/docs/Session.md) with debug mode documentation.
  - Full repository test suite passed via `HTO=0 make test-q1`.

---

## 3. Execution Log
- [01:33] Launched architect refinement with `argon:low`.
- [01:36] Architect refinement completed with concrete specifications; updated and committed Issue #381 ticket (`8ecd21f`).
- [01:37] Dispatched developer `sol:med` for plan-first inspection.
- [01:38] Developer submitted concrete implementation plan. Authorized M1.
- [01:40] M1 completed and committed (`9bda066`). Host review verified diff, unit tests, and API snapshots. Embedded M2 pre-work in ticket (`6075b74`).
- [01:43] M2 completed and committed (`b14fd1d`). Verified bundle rotation, repair toggle, and bar goldens. Updated ticket (`c1dfdee`).
- [01:44] Dispatched developer for M3.
- [01:47] M3 completed and committed (`36bf560`). Quota-1 test suite passed. Ran `make install`. Closed Issue #381 (`4b40072`).
