# Two-Agent Lean Sprint Study: Issue #382 VT Parser Streaming Continuity and Safe Discard

**Date**: 2026-10-24
**Issue**: [#382](file:///home/uwe/projects/loom/issues/382-vt-parser-streaming-continuity-and-safe-discard-for-split-unsupported-escape-sequences.md)
**Roles**:
- Host Orchestrator: Antigravity
- Architect Refiner: `argon:low` (Session: `refiner-382`)
- Developer: `sol:med` (Session: `dev-382`)

---

## 1. Refinement & Root-Cause Discovery
Architect refinement completed in 3m15s (710k tokens processed). It identified critical root causes in `vt/vt.go` and `internal/session/session.go`:

1. **Outer Terminal UI Interleaving during Split Sequences (`session.go`)**:
   `runSession` wrote `out.Write(buf[:n])` followed immediately by `syncSessionMouse` or `drawSessionTabBar` (`\x1b7\x1b[24;1H...`). If `buf[:n]` ended with an incomplete sequence (e.g. `\x1b[38;2;10;`), the tab bar write aborted the child's sequence on the outer terminal. When the next PTY chunk arrived (`20;30m`), the outer terminal rendered it as visible text.
2. **Intermediate Bytes Handling (`VT.escape`)**:
   Charset sequences (`ESC ( B`) consumed only 2 bytes and fell through, leaving `'B'` to leak to the screen.
3. **CSI & OSC Interruption/Cancellation**:
   Embedded `ESC` or `CAN`/`SUB` inside CSI or OSC sequences were improperly parsed, misidentifying `[` as a final byte and leaking parameter tails.
4. **Buffer Bounds**:
   Unbounded `pending` slice for unterminated escape sequences.
5. **SGR Colon Subparameters**:
   Colon flattening disrupted ITU T.416 format (`38:2::R:G:B`), and truncated `38;2;...` sequences fell through to attribute parsing.

## 2. Sprint Milestones
- **M1**: Harden `vt.VT` Escape & Control String Parser (`vt/vt.go`) — Delivered in commit `7703a90`.
- **M2**: Harden `internal/session` Output Stream Framing (`internal/session/session.go`) — Delivered in commit `f8a9d37`.
- **M3**: Adversarial Split-Byte, Malformed Sequence & PTY Verification (`vt/vt_test.go`, `internal/session/session_pty_test.go`)

---

## 3. Execution Log
- [01:14] Launched architect refinement with `argon:low`.
- [01:18] Architect refinement completed with concrete findings; updated and committed Issue #382 ticket (`9195930`).
- [01:19] Dispatched developer `sol:med` for plan-first inspection.
- [01:20] Developer submitted concrete implementation plan. Authorized M1.
- [01:22] M1 completed and committed (`7703a90`). Host review verified diff and unit tests; identified API snapshot update needed for new exports. Embedded pre-work in ticket.
- [01:24] M2 completed and committed (`f8a9d37`). Updated API snapshot and switched `runSession` to `WriteFramed`.
