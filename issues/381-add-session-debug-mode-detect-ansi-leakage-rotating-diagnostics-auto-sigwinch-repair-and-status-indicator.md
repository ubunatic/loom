# 381 — Add session debug mode: detect ANSI leakage, rotating diagnostics, auto-SIGWINCH repair, and status indicator

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: `internal/session`, `internal/editor`, 377, 374, 382, `docs/studies/2026-10-24-multiplexer-ansi-leakage-and-diagnostic-recovery.md`

---

/goal Implement session debug mode that detects ANSI/escape code leakage, logs rotating diagnostics and before/after screenshots, triggers a two-tier auto-repair (parser reset + SIGWINCH toggle) with a cooldown/rotation circuit breaker, and renders a status bar indicator, verified with PTY tests; or stop and report when blocked on a user decision or denied permission.

## 1. Problem & Motivation
When running complex, high-frequency terminal applications like `btop` inside `loom session`, escape sequences or ANSI parsing state can occasionally desynchronize or leak, resulting in unparsed escape characters or raw sequence fragments rendering directly onto the screen.

While issue #382 addresses parser streaming continuity and safe discard at the root, `loom session` requires a robust diagnostic, observability, and self-healing subsystem to catch, report, and recover from any unexpected terminal corruption:
1. Detect when escape sequences leak into the visible buffer or fail to parse.
2. Capture rich diagnostic information (context log, raw bytes, before-repair `.ansi` screenshot).
3. Automatically attempt to self-repair the display using a two-tier recovery: reset wrapper VT parser state to ground, followed by an auto-SIGWINCH resize toggle (forcing child applications like `btop` to re-query terminal dimensions and redraw cleanly).
4. Capture an after-repair `.ansi` screenshot to verify whether the redraw restored clean rendering.
5. Provide a circuit breaker and bounded rotation to prevent disk exhaustion and runaway loops.
6. Display an indicator in the status bar while leakage is observed or being recorded.

## 2. Technical Specification & Architecture
Informed by research into `tmux`, `Ghostty`, and `wezterm` (`docs/studies/2026-10-24-multiplexer-ansi-leakage-and-diagnostic-recovery.md`):

### Diagnostic & Recovery Workflow
1. **Leakage Detection**:
   - Detect orphaned ESC sequences (`\x1b[...`, partial CSI/OSC strings) appearing as printable cell contents in the screen buffer or unconsumed in the VT parser buffer.
2. **Rotating Diagnostic Logging**:
   - Maintain an in-memory ring buffer (e.g. last 64 KB of raw I/O stream).
   - When an anomaly is detected, flush timestamped diagnostic entries, sequence details, and raw stream chunks into a dedicated rotating log file.
   - Capture a "before" `.ansi` screenshot of the corrupt screen state.
3. **Two-Tier Auto-Repair via Parser Reset + SIGWINCH**:
   - Step 1: Force reset the internal VT parser state machine back to Ground.
   - Step 2: Nudge child PTY geometry via `TIOCSWINSZ` (momentary dimension toggle) to deliver `SIGWINCH`. Full-screen TUI apps (e.g. `btop`, `htop`, `tmux`) handle `SIGWINCH` by clearing and repainting the entire viewport.
4. **Post-Repair Verification**:
   - Capture an "after" `.ansi` screenshot once the repaint arrives and return the session to normal processing.
5. **Circuit Breaker & Bounded Retention**:
   - Implement a rate-limiting cooldown (minimum 10s cooldown per session/process before triggering another repair cycle).
   - Maintain a bounded FIFO ring buffer of captures/screenshots per application (e.g. max 5 before/after snapshots and rotated log chunks) so disk space cannot be exhausted even during sustained misbehavior.
6. **Status Bar Indicator**:
   - Render a distinct warning indicator in the `loom session` status bar (e.g. `[!]` or alert glyph) whenever leakage is actively observed, logged, or recovering.

## 3. Implementation & Verification Plan
- **M1 (Advisor Design)**: Audit `internal/session` PTY event loop, `vt.Terminal`, and signal dispatch mechanism against findings in `docs/studies/2026-10-24-multiplexer-ansi-leakage-and-diagnostic-recovery.md`.
- **M2 (Detection, Ring-Buffer Logging & Screenshots)**: Implement detector, rotating capture directory, and before/after `.ansi` captures with bounded retention.
- **M3 (Two-Tier Repair & Circuit Breaker)**: Implement parser reset + SIGWINCH toggle, enforce cooldown timer, and add the status bar warning indicator.
- **M4 (PTY Verification)**: Create automated PTY integration tests using a canary process that emits deliberately malformed/leaked escape sequences to verify detection, logging, SIGWINCH dispatch, and status bar display.
