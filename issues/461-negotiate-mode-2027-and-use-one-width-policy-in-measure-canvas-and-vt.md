# 461 — Negotiate mode 2027 and use one width policy in measure, canvas and VT

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 458, 460

---

## 1. Problem & Motivation
Mode 2027 (Unicode core, "grapheme width") lets an application ask the terminal whether it
measures text by grapheme cluster and switch that on. loom never asks: `measure` only probes the
width of a ZWJ family emoji by cursor report (200 ms, `LOOM_ZWJ` override), and loom's VT answers
`CSI ? 2027 $ p` with 0 (not recognised) after both set and reset. ultraviolet queries 2027 at
start and switches its buffer and renderer width method on a positive reply. When loom and the
terminal disagree on width, cells drift and borders break (see closed 048, 132).
Evidence: 458 batch 2, `docs/studies/tuios/report-width.md` (V1, V3, V6).

## 2. Technical Specification / Findings
- Query DECRQM 2027 at Pane start with a bounded timeout; replies 0–4 (not recognised, set, reset,
  permanently set, permanently reset) and no reply each map to a defined width policy. Enable on
  reply 2, and restore the prior state on exit and around `Suspend`.
- One selected policy must drive `measure`, Canvas, the renderer and `vt.VT`; the ZWJ probe and
  `LOOM_ZWJ` remain the fallback and override.
- `vt.VT` should answer and honour 2027 so `loom session` children can negotiate too.
- x/vt is not a usable oracle: its cursor advance is grapheme-based with 2027 off and on.
- Owner check 2026-11-10 (report-width.md, "Owner manual check results"): foot 1.27.0 answers
  2027 = 1 (set by default); with 2027 reset the ZWJ family takes 8 cells instead of 2. kitty
  0.47.1 answers 0 (not recognised) yet uses grapheme widths. So DECRQM is the first step and
  the CPR probe stays the fallback; extend the probe to the samples where terminals differ
  (ZWJ family, Indic conjunct: foot 2, kitty 1). Ghostty not checked (not installed).
- 458 batch 3 (report-queries.md): build the startup query as one batch ending with DA1, as tcell
  (1 s backstop) and TUIOS (300 ms) do, and include DECRQM 2026 and the kitty `CSI ? u` query, so
  2026 wrapping and kitty push follow the reply instead of opt-in switches. Depends on 462 (Pane
  must consume replies, not pass them to widgets as keys).

## 3. Implementation & Verification Plan
/goal Pane queries 2027, selects one width policy from the reply, applies it to measure, Canvas
and VT, restores terminal state on exit, and is verified headless plus by the owner in at least
two real terminals; stop and report when blocked on a design decision, a denied permission or
missing terminal results, or when no progress is made (clear the goal and write a failure report).

- Re-check live code first (`measure/zwj.go`, `vt/vt.go` private mode state, Pane start/exit).
- Headless tests with `loomtest.VT`: each DECRQM reply, no reply (timeout), set/reset and restore.
- Manual check by the owner, per the project's foot checklist rule.
