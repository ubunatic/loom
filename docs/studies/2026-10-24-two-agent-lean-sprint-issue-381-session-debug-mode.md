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

## 2. Sprint Milestones
- **M1**: Debug Mode Options, CLI Flags, Ring Buffer & Detector Unit Tests — Delivered in commit `9bda066`.
- **M2**: Rotating Diagnostics, Two-Tier Auto-Repair & Status Indicator — Delivered in commit `b14fd1d`.
- **M3**: Automated PTY Canary Integration Tests & Evergreen Docs

---

## 3. Execution Log
- [01:33] Launched architect refinement with `argon:low`.
- [01:36] Architect refinement completed with concrete specifications; updated and committed Issue #381 ticket (`8ecd21f`).
- [01:37] Dispatched developer `sol:med` for plan-first inspection.
- [01:38] Developer submitted concrete implementation plan. Authorized M1.
- [01:40] M1 completed and committed (`9bda066`). Host review verified diff, unit tests, and API snapshots. Embedded M2 pre-work in ticket (`6075b74`).
- [01:43] M2 completed and committed (`b14fd1d`). Verified bundle rotation, repair toggle, and bar goldens. Updated ticket (`c1dfdee`).
