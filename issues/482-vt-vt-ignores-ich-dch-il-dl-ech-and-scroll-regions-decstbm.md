# 482 — vt.VT ignores ICH, DCH, IL, DL, ECH and scroll regions (DECSTBM)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: 481, 477, 382, 373

---

## 1. Problem & Motivation
`vt.VT` has no handler for the CSI editing sequences ICH (`@`), DCH (`P`), IL (`L`), DL (`M`),
ECH (`X`) or for scroll margins DECSTBM (`r`): the CSI switch in `vt/vt.go` (cases from `'m'` at
about line 832 to `'h', 'l'`) skips them. Full-screen programs (vim, less, htop, shells with line
editing) use all of them. `loom session` keeps each tab's screen in a `vt.VT` and repaints a tab from
it on a tab switch (`internal/session/session.go:544`, `repaint := tab.vt.Repaint()`), so such a tab
comes back with stale or shifted lines; `ptytest` screens are wrong in the same way.

## 2. Technical Specification / Findings
Found by the Jules vt-diff session (GH #124, `experiments/vtdiff`, compares loom `vt` with
`charmbracelet/x/vt` on 10x4), host-checked on main `976b96b4`:
- `A\x1b[1@B\x1b[1PC\x1b[1LD\x1b[1ME`: loom gives `ABCDE`, cursor 5,0 (all four ops ignored);
  expected (xterm, upstream) `EBC`, cursor 1,0.
- `\x1b[2;3r1\n2\n3\n4`: loom scrolls the whole screen (`1`, ` 2`, `  3`, `   4`, cursor 4,3);
  expected only rows 2-3 scroll: `1`, `  3`, `   4`, empty, cursor 4,2.
- Also seen, smaller: `abcdefghij\b\b` on 10 columns leaves the cursor at 8, xterm and upstream at 7
  (BS from the pending-wrap state should count from the last column, not from column 10).
- The report labels the first two as upstream bugs; that is wrong, loom is the one that differs.
  Width differences in the same run (VS16, skin tone, ZWJ) are measure policy, handled by 433/460
  (GH #120) and `LOOM_ZWJ`; incomplete UTF-8 holding back the previous cell is F2 in 481.

## 3. Implementation & Verification Plan
/goal `vt.VT` implements ICH, DCH, IL, DL, ECH, DECSTBM (with origin of LF/RI/IND scrolling inside
the margins and cursor home on DECSTBM) and BS from pending wrap, matching xterm, with table tests in
`vt`; stop and report when blocked on a design decision or denied permission; on a loop or no
progress, clear the goal and write a failure report.

- Re-check live code first; port the vtdiff streams as `vt` table tests (no upstream dependency in
  the root module).
- Keep `Repaint()` correct for a screen with margins set (re-emit DECSTBM).
- `make test-q1`, then `make install`; owner check: vim and less in a `loom session` tab, switch tabs
  and back.

### Progress (2026-11-11)
Part 1 landed (Codeberg PR 96, from Jules GH #134): ICH, DCH, ECH and BS after pending wrap, with
`vt/vt_rowedit_test.go`. Left: IL, DL, DECSTBM and the `Repaint()` margin re-emit; split-every-byte
streams for the new sequences were not added.
