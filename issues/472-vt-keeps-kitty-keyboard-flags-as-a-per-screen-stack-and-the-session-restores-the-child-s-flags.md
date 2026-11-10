# 472 — vt keeps kitty keyboard flags as a per-screen stack and the session restores the child's flags

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: 458, 462, 321, 384

---

## 1. Problem & Motivation
`vt.VT` holds one kitty flag value (`vt/vt.go:58`): `CSI > N u` overwrites it, `CSI < u` clears
it (`vt.go:809-818`), and `alternate` leaves it alone (`vt.go:74-87`). `CSI = N u` (set flags) has
no branch and falls through to cursor restore (`vt.go:819`). 458 batch 7 canary
(`docs/studies/tuios/report-session.md`, canary a): push 1, push 3, pop gives `?0u`, not `?1u`;
flags pushed on the alternate screen leak to the main screen; `=31u` changes nothing and moves the
cursor.

The session then re-sends a fixed `ESC[>1u` on tab switch and reattach (`applyTabModes`,
`internal/session/session.go:525-528`, called at `:1396`), so a child that pushed `>31u` gets flag
1 in the outer terminal after a reattach (canary d3). A nested program (an editor started from a
shell that pushed flags) leaves the outer terminal with the wrong key encoding after it exits.

## 2. Technical Specification / Findings
- Kitty spec: per-screen stacks (main and alternate), push/pop with count, `=` with modes 1/2/3,
  bounded depth (TUIOS: 8 entries per screen, `vt/kitty_keyboard.go`). Pin the spec text and
  TUIOS revision from the report.
- Session: on tab switch and reattach send the active tab's current flags (`CSI = N u` or a pop
  plus push), not a fixed 1.
- Pane's own flag 1 push (321, 384) is unchanged.

## 3. Implementation & Verification Plan
/goal `vt.VT` keeps a bounded kitty flag stack per screen with push, pop by count, set (`=` with
modes) and query, and the session restores the active tab's exact flags on switch and reattach,
with tests; stop and report when blocked on an API decision or denied permission; on a loop or no
progress, clear the goal and write a failure report.

- Re-check live code first. Port canary a (scratchpad `b7-canary/vt_canary_test.go`
  `TestKittyStack`) into `vt/vt_modes_test.go` with the expected column from the report; add a
  session test for reattach after `>31u`.
