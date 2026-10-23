# 381 — Add session debug mode: detect ANSI leakage, rotating diagnostics, auto-SIGWINCH repair, and status indicator

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: `internal/session`, `cmd/loom`, 382, `docs/studies/2026-10-24-multiplexer-ansi-leakage-and-diagnostic-recovery.md`

---

/goal Implement session debug mode that detects ANSI/escape code leakage, logs rotating diagnostics and before/after screenshots, triggers a two-tier auto-repair (parser reset + SIGWINCH toggle) with a cooldown/rotation circuit breaker, and renders a status bar indicator, verified with PTY tests; or stop and report when blocked on a user decision or denied permission.

## 1. Problem & Motivation
When running complex, high-frequency terminal applications like `btop` inside `loom session`, escape sequences or ANSI parsing state can occasionally desynchronize or leak, resulting in unparsed escape characters or raw sequence fragments rendering directly onto the screen.

While issue #382 addresses parser streaming continuity and safe discard at the root, `loom session` requires a robust diagnostic, observability, and self-healing subsystem to catch, report, and recover from any unexpected terminal corruption:
1. Detect when escape sequences leak into the visible buffer or fail to parse.
2. Capture rich diagnostic information (context log, raw bytes from a 64 KB ring buffer, before-repair `.ansi` screenshot).
3. Automatically attempt to self-repair the display using a two-tier recovery: reset wrapper VT parser state to ground (`t.vt.ResetParser()`), followed by an auto-SIGWINCH resize toggle (forcing child applications like `btop` to re-query terminal dimensions and redraw cleanly).
4. Capture an after-repair `.ansi` screenshot to verify whether the redraw restored clean rendering.
5. Provide a circuit breaker (10s cooldown) and bounded rotation (max 5 rotating capture bundles) to prevent disk exhaustion and runaway loops.
6. Display a distinct warning indicator `[!]` in the status bar while leakage is observed or being recorded.

## 2. Technical Architecture & Constraints
1. **Activation & Entrypoint**:
   - CLI flags in `cmd/loom/session.go`: `--debug` (`bool`) and `--debug-dir` (`string`).
   - Environment variables: `LOOM_SESSION_DEBUG=1` and `LOOM_DEBUG_DIR=<path>` for headless testing.
   - Entrypoint in `internal/session`: `Options` and `RunWithOptions(...)`, keeping `Run(...)` backward-compatible.
2. **In-Memory Ring Buffer**:
   - Fixed 64 KB circular byte buffer per tab in `internal/session/debug.go` recording raw chunks read from `t.master`.
3. **Leakage & Anomaly Detector**:
   - VT safety-cap discard detection (`v.discardControl != 0` or discard hook).
   - Stalled incomplete escape sequences (`StallTimeout`, default 1s).
   - Stray control character / orphaned CSI parameter fragments (e.g. `38;2;...m`) in newly rendered screen cells.
4. **Two-Tier Auto-Repair**:
   - Tier 1: Reset parser state to Ground via `t.vt.ResetParser()`.
   - Tier 2: Momentary child PTY dimension toggle (`setPTYSize(t.master, cols, toggleRows)` then restore to `childRows`) to trigger `SIGWINCH`.
5. **Rotating Diagnostics & Circuit Breaker**:
   - Directory: `--debug-dir` or `$XDG_STATE_HOME/loom/debug` / `~/.local/state/loom/debug`.
   - Captured bundle per incident: `<prefix>.log`, `<prefix>-before.ansi`, `<prefix>-after.ansi`.
   - Bounded retention: FIFO rotation keeping at most 5 capture bundles.
   - Circuit breaker: 10s cooldown (`RepairCooldown`) between repair cycles and disk writes.
6. **Status Bar Indicator & Golden Compatibility**:
   - Render `[!]` in `sessionBarContent` when alert is active.
   - When no alert is active, bar output is 100% byte-identical to maintain existing golden tests (`TestSessionBarGoldens`).

## 3. Milestones & Delivery

### M1 Delivered: Debug Mode Options, CLI Flags, Ring Buffer & Detector Unit Tests
- **Status**: Delivered & verified in commit `9bda066`.
- Added `Options` struct and `RunWithOptions(...)` to `internal/session/session.go`, maintaining `Run` compatibility.
- Added `--debug` and `--debug-dir` flags to `cmd/loom/session.go` with `LOOM_SESSION_DEBUG` / `LOOM_DEBUG_DIR` environment precedence.
- Implemented 64 KB circular raw byte buffer and anomaly detector in `internal/session/debug.go`.
- Added unit tests in `internal/session/debug_test.go` covering ring wrapping, control parsing, stall timeouts, and false positive checks.

### M2 Delivered: Rotating Diagnostics, Two-Tier Auto-Repair & Status Indicator
- **Status**: Delivered & verified in commit `b14fd1d`.
- Implemented `writeDebugBundle` writing `.log` with raw stream chunks and `-before.ansi` / `-after.ansi` screen captures, with FIFO rotation keeping at most 5 bundles.
- Implemented two-tier auto-repair in `runSession`: `t.vt.ResetParser()` followed by momentary `TIOCSWINSZ` toggle to trigger child `SIGWINCH`, with 10s circuit breaker cooldown.
- Implemented status bar warning indicator `[!]` in `sessionBarContent`, maintaining 100% byte-identity when no alert is active (`TestSessionBarGoldens` passes).

### M3: Automated PTY Canary Integration Tests & Evergreen Docs
- **Scope**:
  - Automated PTY test in `internal/session/session_pty_test.go` with canary child verifying detection, `.ansi` captures, `SIGWINCH` redraw, and `[!]` indicator.
  - PTY test verifying cooldown rate-limiting and max-5 bundle rotation.
  - Update `docs/Session.md`.
  - Run full verification (`HTO=0 make test-q1`).
