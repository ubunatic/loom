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
- [x] **096 textrender vs terminal**: run the example in your terminal(s) (tilix, foot, VTE, kitty)
  and compare with the recorded `knownDivergences` (ZWJ family: Loom 6, terminal 2; flag: Loom 4,
  terminal 2). Decide whether `loom.StringWidth` changes; the width audit itself is tracked in 048.
  Result 2026-09-27 (tilix, foot, ptyxis, kitty, alacritty; notes in docs/progress/096/comments.txt):
  all six frames look right. Every terminal draws the ZWJ family as 3 separate emojis (6 columns),
  which matches Loom's width 6. CORRECTION (ansiviewer ruler, same day): the frames contain no U+200D
  joiner, so the three emojis come from Loom's output, not the terminals; see 134. The flag varies by terminal
  (letters "DE" in tilix, ptyxis and alacritty, a wavy flag in foot, a color flag in kitty), which is
  font or terminal behavior. Open: tilix shows a big gap between the text and '>' in M2-buttons.
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

## Sprint 124/128/064 (2026-09-27)

- [ ] **124 F10 global quit**: in textedit, ansiedit, ansiviewer, filebrowser and paint, press F10 while an editor or popup has focus. The app exits and the terminal is restored, and the footer shows "F10 Quit".
- [ ] **064 hosted tabs**: in loom-demo, open splash and monitor as two tabs. Both redraw live, switching tabs works, and when splash finishes it closes only its own tab.
- [ ] **064 standalone monitor**: `monitor --watch`, press F10, then check that no monitor process or collector is left running.

- [ ] **111 media widget:** run `media` (examples/media) with a PNG, and with a video if ffmpeg is installed. Check that the image keeps its proportions and is centered, that colors look right in halfblock, quadblock and sextant, that nothing is painted outside the image area, and that quitting leaves no ffmpeg process behind.
- [ ] **103 usage example:** run `usage` (examples/usage). Check the two boxes, that the heat colors match `harnez usage --compact --watch`, that the key switches between the plain and Loom views, and that resize and quit work.
- [ ] **132 flag row in ansiviewer:** `ansiviewer docs/progress/096`, open M2-clipping.ansi in tilix,
  alacritty and ptyxis. The flag row fills the full width with no default-background gap, and in
  ptyxis the label comes before the sample ("Flag DE"). If ptyxis still swaps the order, reopen 132.

- [ ] **089 cursor effects:** run `ansicanvas_demo --cursor-fx` in tilix and foot. Check glow
  strength, star trail, pulse, smooth motion without lag, and idle CPU near zero while the mouse rests.

- 134 (ZWJ width detection, 013fa84): open a Loom app with the family emoji (e.g. ansiviewer on docs/progress/096/M2-buttons.ansi) in tilix and foot; rows should end flush, no gap before `>`. `LOOM_ZWJ=join|split` forces a mode.

## Roadmap 180 — human feedback queue (2026-09-30)

Decided 2026-09-30 (user: host makes all decisions): the host builds these after the main list, choosing look and behavior itself. Review afterwards; they do not block. Order and dependencies are in issues/180.
- [ ] H1 155 Dialog — built (339f523); review with `loom widgets --show Dialog`. Host choices: centered by default (optional Rect), sized to content; Esc dismisses; Left/Right/Tab move the highlight, first button preselected; Enter calls OnSelect without quitting the app; Popup's sharp border, title and fill; selected button bold with ▶, others dim.
- [ ] H2 170 Menu and MenuBar — built (2818936); review with `loom widgets --show MenuBar`. Host choices: one-row bar, bordered dropdown under the active title; `---` separators, right-aligned shortcut labels, check marks, dim disabled items, underlined mnemonics; F10 or Alt+mnemonic opens, Left/Right switch titles, Up/Down move, Enter/Space activate; first Esc closes the dropdown, second unfocuses; hover follows the mouse, clicking outside closes.
- [ ] H3 193 Nested submenus — built (7df9741); review with `loom widgets --show MenuBar`. Host choices: submenu opens beside its parent row, flips left at the right edge, clamped vertically; Right opens, Left/Esc close one level; hover opens instantly (no delay); `›` marks items with a submenu.
- [ ] H4 175 DatePicker — built (9198d2f, 1f761dd); review with `loom widgets --show DatePicker`. Host choices: date only, no time of day; compact month grid, week starts Monday (configurable); arrows move day/week, PgUp/PgDn month, Ctrl+Left/Right year; Enter selects; typing YYYY-MM-DD + Enter also works; `*` today, `>` selected, underlined cursor; inclusive min/max bounds, navigation clamps.
- [ ] H5 192 Chart widget — built (5944999, 44828f1); review with `loom widgets --show Chart`. Host choices: thin braille line plots (2x4 dots per cell), solid-block grouped bars; Y axis on every row with round tick steps (1/2/5 × 10^n), range rounded outward; X ticks at sample positions; legend only for several series; distinct 16-colour palette defaults; auto-scale with optional fixed range.
- [x] H6 112 Media controls — built 6f5844f, 7c41e53, 3f62a8b. Host choices: two-row footer (key hints above; ▶ Play, [-] [+], zoom level, progress to the right), shrinks/omits on short widgets; space play/pause; +/-/wheel zoom ×1.25 in 1×–8×; arrows pan 12%; drag pans; 0 resets; pan clamps to bounds; ProgressBar shows activity only (no duration API).
- [x] Clipboard decided: filed 194 (OSC 52 copy only), last in the main list
