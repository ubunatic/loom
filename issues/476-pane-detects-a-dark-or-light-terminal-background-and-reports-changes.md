# 476 — Pane detects a dark or light terminal background and reports changes

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 458, 462, 461

---

## 1. Problem & Motivation
loom has no dark/light background detection: no OSC 11 query, no mode 2031 colour-scheme reports,
no DSR 996 (458 batch 8, `docs/studies/tuios/report-desktop.md`; `theme.go`, `event.go:707-731`).
Apps pick a theme by name only. Lip Gloss v2 queries OSC 11 with a timeout and defaults to dark;
Bubble Tea v2 has an OSC 11 request and typed colour/scheme events; TUIOS enables 2031 for local
clients and requeries colours on a scheme event (pins in the report and `report-queries.md`).
Glamour picks dark/light from `GLAMOUR_STYLE` and does not query.

## 2. Technical Specification / Findings
- Depends on 462: replies must be consumed centrally first, else an OSC 11 reply is typed into
  the focused TextEntry (batch 8 canary C2).
- Opt-in Pane query at start (OSC 11, ended by DA1 with a backstop, as 461), plus mode 2031
  reports while running; deliver a typed "background is dark/light" value to the app, which
  chooses the theme. No query in SSH or when the terminal does not answer.
- Owner decision: default when silent (dark, like Lip Gloss) and whether themes in
  `spec/themes.yaml` get dark/light pairs.

## 3. Implementation & Verification Plan
/goal Pane can report whether the terminal background is dark or light at start and when it
changes, through a typed value that never reaches widgets as keys, with tests; stop and report when
blocked on 462, an owner decision or a denied permission; on a loop or no progress, clear the goal
and write a failure report.

- Re-check live code and 461/462 first; verify replies in foot and kitty with the owner (one small
  step each).
