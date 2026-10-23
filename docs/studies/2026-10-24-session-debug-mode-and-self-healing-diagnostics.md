# Case Study: Session Debug Mode, Diagnostic Ring Buffers, and Self-Healing SIGWINCH Recovery

**Date**: 2026-10-24
**Scope**: Issue #381 (Add session debug mode: detect ANSI leakage, rotating diagnostics, auto-SIGWINCH repair, and status indicator)
**System**: `cmd/loom`, `internal/session`
**Sprint**: Two-agent lean sprint with `argon:low` (architect refiner) and `sol:med` (developer)

---

## 1. Header & Context
Following root-cause VT parser streaming hardening in Issue #382, Issue #381 implemented an observability and self-healing subsystem for `loom session`. The goal was to detect unexpected escape sequence leakage, maintain a 64 KB rolling raw I/O ring buffer, write rotating diagnostic capture bundles (`.log`, `-before.ansi`, `-after.ansi`), trigger a two-tier auto-repair mechanism (`ResetParser` + momentary `TIOCSWINSZ` toggle delivering `SIGWINCH`), enforce a 10-second circuit breaker cooldown, and render an active warning indicator (`[!]`) in the session status bar without disturbing non-alert golden visual tests.

---

## 2. Executive Summary
The session debug mode was architected and delivered across three lean milestones:
1. **Architect Discovery & Spec Refinement**: `argon:low` mapped out the CLI flag precedence (`--debug`, `--debug-dir`, `LOOM_SESSION_DEBUG`, `LOOM_DEBUG_DIR`), the 64 KB ring buffer layout, multi-condition leakage detection (parser stalls, overlong controls, and suspicious screen cell fragments), two-tier repair semantics, and strict preservation of existing status bar golden snapshots (`TestSessionBarGoldens`).
2. **Milestone 1 (Options, Flags & Detector)**: Added `session.Options` and `session.RunWithOptions` to keep `session.Run` backward-compatible. Bound `--debug` and `--debug-dir` CLI flags. Implemented a 64 KB circular byte buffer and stream observer in `internal/session/debug.go`, verified with unit tests.
3. **Milestone 2 (Diagnostics, Repair & Status Indicator)**: Implemented `writeDebugBundle` writing timestamped `.log`, `-before.ansi`, and `-after.ansi` files with FIFO rotation capping storage at 5 incident bundles. Integrated two-tier auto-repair (`t.vt.ResetParser()` + `TIOCSWINSZ` toggle) with a 10s circuit breaker. Added `[!]` in `sessionBarContent` during active alerts, keeping normal bars byte-identical to golden tests.
4. **Milestone 3 (PTY Canary Verification & Documentation)**: Created automated PTY integration tests using a canary child script trapping `WINCH`, verifying detection, `.ansi` captures, `SIGWINCH` delivery, status bar warning display, cooldown suppression, and max-5 bundle rotation. Documented debug mode in `docs/Session.md`. Full Quota-1 test suite passed (`HTO=0 make test-q1`).

---

## 3. What Worked Well
- **Zero-Coding Orchestrator with Strict Review**: All implementation was driven by developer `sol:med` while the host focused on review, API stability assertion, and milestone coordination.
- **PTY Canary Test Design**: Emitting an intentional truecolor leakage fragment from a child shell that traps `WINCH` allowed deterministic, end-to-end verification of signal delivery, screen redraw, before/after captures, and cooldown gating without relying on real flaky applications or sleep heuristics.
- **Golden Test Preservation**: By designing the `[!]` indicator to only appear during active alerts and reserving width dynamically, all existing status bar golden tests (`TestSessionBarGoldens`) passed without requiring golden file churn.

---

## 4. Honest Post-Mortem (Failures, Bugs & Near-Misses)
1. **CAN Byte in Split CSI (M1 Near-Miss)**: During M1 development, unit tests revealed that a `CAN` (`0x18`) byte embedded inside a split CSI sequence remained in the detector's pending buffer rather than immediately returning the detector to Ground state. The developer caught and corrected this parser transition before committing M1.
2. **PTY Child Echo Timing in Integration Tests**: In PTY canary tests, shells default to echo mode on PTYs. Sending newline commands during active escape sequences can inadvertently terminate open CSI sequences. Disabling child echo via `stty -echo` or sequencing input reads with timed barriers avoided false positive terminators.
3. **Kernel TIOCSWINSZ Idempotency**: As noted during research, calling `TIOCSWINSZ` with identical rows/cols is a no-op in the Linux tty layer and will not generate `SIGWINCH`. The developer ensured a positive momentary toggle (`childRows - 1` if `childRows > 1` else `childRows + 1`) followed immediately by restoring the true dimension.

---

## 5. Quality & Invariants Audit

| Invariant / Check | Status | Evidence / Measurement |
|---|---|---|
| **Zero Coding Orchestrator** | 100% compliant | Host edited zero source code; all changes executed by `sol:med` |
| **Backward Compatibility** | 100% compliant | `session.Run` preserved as compatibility wrapper; root API snapshots unchanged |
| **Disk Space Bounds** | Enforced | FIFO rotation caps disk usage to at most 5 bundles (15 files, ~400 KB max) |
| **Rate Limiting** | Enforced | 10-second circuit breaker prevents repair and write storms |
| **Visual Golden Integrity** | Verified | `TestSessionBarGoldens` passed byte-for-byte in normal state |
| **Full Repository Tests** | Passed | `HTO=0 make test-q1` green across all packages |

---

## 6. Efficiency & Velocity Assessment
- **Architect Refinement**: 2m45s (1.2M tokens processed).
- **Developer Implementation**:
  - M1 (Options, Flags, Ring Buffer & Detector): 1m16s, commit `9bda066`.
  - M2 (Diagnostics, Repair & Indicator): 1m50s, commit `b14fd1d`.
  - M3 (PTY Canaries & Docs): 3m14s, commit `36bf560`.
- **Total Sprint Duration**: ~15 minutes from prompt to ticket closure and documentation.

---

## 7. Key Learnings & Evergreen Upstream
1. **Always Nudge PTY Geometry to Deliver SIGWINCH**:
   Calling `ioctl(fd, TIOCSWINSZ, &ws)` with unchanged dimensions is silently discarded by the kernel. An auto-repair system must momentarily change a dimension to guarantee `SIGWINCH` delivery.
2. **Decouple Disk Operations from Session Mutex**:
   Writing `.log` and `.ansi` files to disk can block on slow filesystems. Snapshotting the ring buffer and VT screen under the session mutex and performing file writes in a separate goroutine keeps the PTY read loop responsive.
3. **Documented in Evergreen**:
   `docs/Session.md` now details debug mode invocation, capture directory precedence, and repair behaviors.

---

## 8. File & Diff Summary
- [`cmd/loom/session.go`](file:///home/uwe/projects/loom/cmd/loom/session.go): Added `--debug` and `--debug-dir` flags with env defaults.
- [`internal/session/session.go`](file:///home/uwe/projects/loom/internal/session/session.go): Added `Options`, `RunWithOptions`, two-tier auto-repair, and `[!]` status bar alert rendering.
- [`internal/session/debug.go`](file:///home/uwe/projects/loom/internal/session/debug.go): Implemented 64 KB ring buffer, leakage detector, and rotating bundle writer.
- [`internal/session/debug_test.go`](file:///home/uwe/projects/loom/internal/session/debug_test.go): Unit tests for ring buffer, control parser, stalls, rotation, and status bar.
- [`internal/session/session_pty_test.go`](file:///home/uwe/projects/loom/internal/session/session_pty_test.go): PTY canary tests for SIGWINCH repair, diagnostics, cooldown, and 5-bundle rotation.
- [`docs/Session.md`](file:///home/uwe/projects/loom/docs/Session.md): Evergreen documentation for session debug mode.
- [`docs/studies/2026-10-24-two-agent-lean-sprint-issue-381-session-debug-mode.md`](file:///home/uwe/projects/loom/docs/studies/2026-10-24-two-agent-lean-sprint-issue-381-session-debug-mode.md): Running sprint log and milestone records.
