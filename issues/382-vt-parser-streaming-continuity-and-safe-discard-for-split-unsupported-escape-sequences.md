# 382 — VT parser streaming continuity and safe discard for split/unsupported escape sequences

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: `vt`, `internal/session`, 381, `docs/studies/2026-10-24-multiplexer-ansi-leakage-and-diagnostic-recovery.md`

---

/goal Fix escape sequence leakage in vt/session by ensuring parser state continuity across PTY read boundaries, buffering incomplete output sequences without time-based flushes, and safely discarding unsupported/malformed controls, verified with adversarial split-byte PTY tests; or stop and report when blocked on a user decision or denied permission.

## 1. Problem & Motivation
Interactive TUI monitoring tools like `btop` emit high-frequency, complex ANSI sequences (truecolor SGR, private mode toggles, cursor addresses) that frequently cross OS read buffer boundaries. When `loom session` processes these PTY chunks:
1. Partial escape sequences split across `read()` calls can be misidentified as plain text or terminated prematurely, causing trailing sequence bytes (e.g. `[38;2;...`) to leak directly into visible cell contents.
2. Incomplete output sequences are vulnerable if time-based Escape timeouts (intended only for keyboard input ambiguity) are erroneously applied to the child PTY output stream.
3. Unsupported or malformed escape sequences fall through to printable characters rather than being cleanly consumed to their terminator and discarded.

## 2. Technical Specification & Architecture
Based on research into `tmux`, `Ghostty`, and `wezterm` (`docs/studies/2026-10-24-multiplexer-ansi-leakage-and-diagnostic-recovery.md`):

1. **Persistent Parser State Across Chunks**:
   - The VT parser state machine (`vt.Terminal`) must survive across individual PTY `read()` calls.
   - If a chunk ends after `ESC`, during CSI parameters, or inside an OSC/DCS payload, the parser must retain its intermediate state (`csi_param`, `osc_string`, etc.) and seamlessly resume when the next chunk arrives.
2. **No Time-Based Flushes on Output Stream**:
   - Ensure the PTY output parsing loop never applies standalone Escape timeouts. Bounded buffers (e.g. max 1024 bytes per control sequence) should prevent memory leaks on genuinely malformed/unterminated streams, resetting to ground state without dumping raw control characters onto the grid.
3. **Safe Discard of Unsupported Controls**:
   - Unsupported or unhandled CSI/OSC/DCS/APC sequences must be consumed through their valid terminator (e.g. `m`, `h`, `l`, ST `\x1b\\`, or BEL `\x07`) and dropped.
   - Under no circumstances should unparsed control fragments be written to printable screen cells.
4. **Adversarial Stream Tests**:
   - Unit and PTY tests that split ANSI sequences (colors, cursor motions, private modes) byte-by-byte across simulated read boundaries to assert zero visual leakage.

## 3. Implementation & Verification Plan
1. Audit `vt/` and `internal/session/` PTY read handling for chunk boundaries and parser continuity.
2. Implement strict state preservation and safe discard in the parser.
3. Add table-driven tests feeding split sequences (at every byte boundary) into `vt.Terminal`.
4. Verify with `go test ./vt ./internal/session` and run `make test-q1`.
