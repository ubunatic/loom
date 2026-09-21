# 102 — Human review collection: manual checks for lean-sprint deliveries

**Status**: Open — manual checks
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Review / Manual test
**Related**: 037, 060, 081, 088, 092, 093, 096, 097, 099, 100, 048; `docs/progress/`

---

## 1. Purpose

Collection ticket for checks that no headless test can settle. The tickets above are closed
on tests and `.ansi` evidence; the items below need eyes or a real terminal. Tick an item
when checked, and file a separate ticket for any defect found (do not reopen the closed one).
Whoever completes the last item closes this ticket.

## 2. Checklist

- [ ] **All evidence frames**: `for f in docs/progress/*/*.ansi; do echo "== $f"; cat "$f"; done`
  in a real terminal. Spot check that frames look right, not just that tests pass (73 frames).
- [ ] **096 textrender vs terminal**: run the example in your terminal(s) (tilix, foot, VTE, kitty)
  and compare with the recorded `knownDivergences` (ZWJ family: Loom 6, terminal 2; flag: Loom 4,
  terminal 2). Decide whether `loom.StringWidth` changes; the width audit itself is tracked in 048.
- [ ] **092 scrollbar drag**: drag the View and Split scrollbars with a real mouse; check drag
  past the track ends, release outside the window, and drag inside a Split child.
- [ ] **093 filebrowser mouse**: click on row text, on trailing whitespace and on the border
  in a real terminal; only text should select.
- [ ] **060 Ticker**: run an example with a Ticker for a minute while moving the mouse and typing;
  check for smooth ticks, no freeze under input storms, no CPU spin when idle.
- [ ] **088 keys**: press the 3 terminal-limitation keys (Ctrl-I, Ctrl-J, Ctrl-M) and a sample of
  the 217 audited keys in your terminal; confirm `docs/KeyDefaults.md` matches what you see.
- [ ] **037 boxes**: eyeball the BoxStyle gallery and popup frames with your font; check
  border joins and glyph widths.
- [ ] **081 treemap**: check that squarified vs slice-dice look sensible at 3 sizes and the
  ColorScale is readable on dark and light themes.
- [ ] **097 ansiviewer**: run `ansiviewer`, record with `--record`, replay; see 099 and 100
  for the known defects (recording, terminal width, top-bar background).
- [ ] **Two old processes**: `splash --watch` PIDs 920646 and 920686 predate the sprints and may
  still be running; confirm and kill them if unwanted.
