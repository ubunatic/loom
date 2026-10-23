# 382 — VT parser streaming continuity and safe discard for split/unsupported escape sequences

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: `vt`, `internal/session`, 381, `docs/studies/2026-10-24-multiplexer-ansi-leakage-and-diagnostic-recovery.md`

---

/goal Fix escape sequence leakage in vt/session by ensuring parser state continuity across PTY read boundaries, withholding incomplete sequences from outer terminal writes before session control injections, and safely discarding unsupported/malformed controls, verified with adversarial split-byte PTY tests; or stop and report when blocked on a user decision or denied permission.

## 1. Problem & Root-Cause Audit
Inspection of `vt/vt.go` and `internal/session/session.go` by architect refinement identified the precise root causes of escape code leakage when running high-frequency TUI tools like `btop`:

1. **Outer Terminal Interleaving Corrupts Split Sequences in `internal/session/session.go`**:
   In `runSession`, `t.vt.Write(buf[:n])` buffers trailing incomplete sequences in `t.vt.pending`. However, `session.go` forwards raw `out.Write(buf[:n])` to the outer terminal and immediately follows with `syncSessionMouse` or `drawSessionTabBar` (which emits `\x1b7\x1b[24;1H...`). If `buf[:n]` ended mid-sequence (e.g. `\x1b[38;2;10;`), the tab bar sequence aborts the child's incomplete CSI on the outer terminal. When the next PTY chunk arrives (`20;30mHello`), the outer terminal is in Ground state and prints `20;30mHello` literally as visible text!
2. **Intermediate Byte Handling in `VT.escape`**:
   Standard charset/DEC sequences with intermediate bytes (`0x20..0x2F`) like `ESC ( B` currently return `2, true` in `VT.escape`, leaving `'B'` or `'0'` to be printed as literal screen text.
3. **Interrupted or Cancelled CSI Sequences**:
   If a CSI sequence is interrupted by a new `\x1b` or cancelled by `CAN` (`0x18`) or `SUB` (`0x1A`), `VT.escape` previously scanned past `0x1B`, mistook `[` as the final byte of the first sequence, and dumped subsequent parameters to visible cells.
4. **String Controls (OSC/DCS/APC/PM/SOS) and Unbounded Buffer**:
   Missing `SOS` (`'X'`), missing `0x9C` ST, and missing abort-on-`ESC`. An unterminated string control could grow `pending` indefinitely or swallow valid commands.
5. **SGR Colon Subparameters & Truncated Tail**:
   `strings.ReplaceAll(s, ":", ";")` breaks ITU T.416 colon subparameters (`38:2::R:G:B` with empty colorspace ID, or `4:3`), and truncated `38;2;...` sequences fall through to misinterpret color numbers as SGR attributes.

## 2. Technical Architecture & Constraints
1. **DEC VT Parser Hardening (`vt/vt.go`)**:
   - `ESC` intermediate bytes (`0x20..0x2F`): consume through final byte (`0x30..0x7E`).
   - Abort & Cancellation: `0x18` (`CAN`) and `0x1A` (`SUB`) abort sequence and return to Ground. A new `0x1B` inside an in-flight CSI or string control aborts the prior sequence and starts fresh.
   - String Controls: Support OSC (`]`), DCS (`P`), PM (`^`), APC (`_`), and SOS (`X`), terminated by `BEL` (`\x07`), ST (`\x1b\\`), `0x9c`, or cancelled by new `ESC` / `CAN` / `SUB`.
   - Max Sequence Cap: Enforce `maxPendingBytes = 1024` on unclosed control sequences; reset safely to Ground without dumping raw bytes onto screen cells.
   - SGR Subparameters: Correctly parse colon-separated subparameters (`38:2::R:G:B`, `48:2::R:G:B`, `4:x`, `58:...`) and advance past truncated `38`/`48` sequences.
   - Add `ResetParser()` method on `vt.VT` (clears pending buffer).
2. **Session Output Stream Integrity (`internal/session/session.go`)**:
   - Withhold incomplete trailing sequences from `out.Write` before injecting session UI controls (`syncSessionMouse`, `drawSessionTabBar`). Only complete sequence/cluster bytes are forwarded to the outer terminal before UI draws.

## 3. Milestones

### M1: Harden `vt.VT` Escape & Control String Parser (`vt/vt.go`)
- Implement DEC-compliant escape dispatch in `VT.escape` and `VT.Write` for intermediate bytes (`0x20..0x2F`).
- Support abort on embedded `ESC` and cancellation on `CAN`/`SUB` in CSI and string controls.
- Add support for SOS (`X`), C1 ST (`0x9C`), and colon SGR subparameters.
- Enforce `maxPendingBytes = 1024` safety cap and add `ResetParser()`.

### M2: Harden `internal/session` Output Stream Framing (`internal/session/session.go`)
- Buffer/withhold trailing incomplete escape sequences and incomplete UTF-8 clusters so `out.Write` never interleaves `drawSessionTabBar` or `syncSessionMouse` inside a child's incomplete sequence.
- Ensure tab switches mid-sequence do not corrupt virtual or outer screens.

### M3: Adversarial Split-Byte, Malformed Sequence & PTY Verification (`vt/vt_test.go`, `internal/session/session_pty_test.go`)
- Table-driven unit tests in `vt/vt_test.go` verifying byte-by-byte split continuity, interrupted/cancelled sequences, colon SGR, and overlong sequence caps.
- PTY integration test in `internal/session/session_pty_test.go` asserting zero leakage with active tab bar updates during split PTY chunks.
- Pass `HTO=0 make test-q1`.

## 4. Edge Cases & Verification Plan
- Split immediately after `ESC`, inside CSI params, across multi-byte intermediates, and across 2-byte ST.
- Split multi-byte UTF-8 and ZWJ grapheme clusters.
- Active tab bar redraw during split sequence chunks.
- Quota-1 test: `HTO=0 make test-q1`.
