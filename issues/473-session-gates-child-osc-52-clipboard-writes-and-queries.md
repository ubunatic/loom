# 473 — session gates child OSC 52 clipboard writes and queries

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 458, 465

---

## 1. Problem & Motivation
`loom session` forwards a child's OSC 52 bytes unchanged to the attached client whenever the tab
is active (`internal/session/session.go:764`, `:782-784`; `vt.WriteFramed` returns the OSC slice,
`vt/vt.go:220-234`). This holds for writes and for clipboard queries (`ESC]52;c;?`). The VT
neither stores nor answers them. 458 batch 7 canaries c and d2
(`docs/studies/tuios/report-session.md`). So any program in a session tab can set the host
clipboard, and can read it if the outer terminal answers queries. TUIOS gates writes with
`osc52_write` off/on/ask/focused (default focused) and answers queries with empty content
(`internal/app/clipboard_osc52.go`, pinned in the report).

465 covers loom's own Ctrl+C/X/V clipboard writes, not passthrough from child programs.

## 2. Technical Specification / Findings
- A spec value in `spec/defaults.yaml` `session:` for the write policy (e.g. off, active-tab,
  on); default to be decided by the owner.
- Queries: answer empty in the VT, or forward only by explicit opt-in; never forward from an
  inactive tab.
- Unknown: whether foot and kitty answer a forwarded query (owner check in foot/kitty, one small
  step).

## 3. Implementation & Verification Plan
/goal `loom session` applies a spec-configured policy to child OSC 52 writes and does not leak
clipboard queries to the host unless enabled, with tests; stop and report when blocked on the
default policy choice or a denied permission; on a loop or no progress, clear the goal and write a
failure report.

- Re-check live code first; port canary d2 (`TestB7OSC52AndKittyForwarded` in the report) to a
  session test per policy value.
