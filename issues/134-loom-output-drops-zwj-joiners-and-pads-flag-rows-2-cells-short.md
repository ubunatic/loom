# 134 — Loom output drops ZWJ joiners and pads flag rows 2 cells short

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: 096, 132, 048, 133

---

## 1. Problem & Motivation
With ansiviewer's ruler (81d5d6b) the terminal background now shows where each ANSI row ends. The
user's screenshots of docs/progress/096/M2-buttons.ansi (tilix, foot, 2026-09-27) show two Loom
output bugs, which the 096 review had put down to terminal behavior:

1. **ZWJ joiners are dropped.** None of the six 096 frames contains U+200D. The family 👨‍👩‍👧 is
   written as three separate emojis, which is why every terminal drew three. Loom measures the
   cluster as 6 columns (measure.go knows ZWJ) but the output loses the joiner.
2. **Flag rows are 2 cells short.** Loom counts 🇩🇪 as 4 columns, and terminals advance 2. Row
   padding is computed with 4, so each row with a flag ends 2 cells early and the rest of the row
   shows the terminal background. The label order in M2-buttons (sample first, then label) is in the
   file itself and is correct.

## 2. Technical Specification / Findings
- Check: `grep -a -c $'\u200d' docs/progress/096/*.ansi` gives 0 for all frames.
- Loom's ANSI box check (TestAllAnsiAssetsHaveValidBoxes) also counts the flag as 4 (see 048), while
  ansiviewer's replay via ParseANSI places the next text 2 later (132): Loom is inconsistent.
- Where the joiner is lost is unknown: cluster splitting in Canvas.Write/set (textClusters), the
  cell renderer, or the recorder.

## 3. Implementation & Verification Plan
- M1: failing tests: write a ZWJ family and a flag through Canvas and the renderer, and assert that
  the bytes keep U+200D and that a flag row's output is exactly as wide as the row.
- M2: keep the joiner in the output; count a regional-indicator pair as 2 columns everywhere
  (StringWidth, box check, ParseANSI) to match the terminals tested in 096.
- M3: re-record the 096 frames; the user checks them with ansiviewer's ruler (102).

## 4. Terminal test with the joiner kept (user, 2026-09-27, `echo` of 👨‍👩‍👧‍👦 with U+200D)
- Joins into one family glyph (2 columns): foot, kitty.
- Draws separate emojis (2 columns each): tilix, ptyxis, alacritty.
So the width of a ZWJ sequence depends on the terminal, like the flag (2 columns as letters or glyph
in all five). Keeping the joiner is still right (joining terminals need it), but no single width is
right everywhere.

Revised plan:
- M2 keeps U+200D in the output and counts a flag as 2 columns (all five terminals agree).
- New M2b: measure ZWJ width per terminal at startup with a canary (docs/Canary.md): write a ZWJ
  sequence off-screen or on the alternate screen, ask for the cursor position (DSR `ESC[6n`), and
  set Loom's ZWJ width mode from the reply; fall back to the separate-emoji width when there is no
  reply in time, and allow an override (env var, e.g. `LOOM_ZWJ=join|split`). Probe once, in
  tilix, foot, kitty, ptyxis and alacritty, before building on it.

## 5. Plan (dev-134, reviewed by the host)
- M1: failing tests that U+200D survives Canvas.Write/Set, the cell renderer (Canvas.Row) and ANSI
  replay; find and fix the loss point.
- M2: flags count 2 columns everywhere. Known outliers: the ANSI box check (ValidateAnsiBox) and
  ParseAnsiBuffer sum RuneWidth per rune; measure.go, parse_ansi.go and rawscreen.go handle pairs.
- M2b: a probe command `loom-probe` (installed by make install) that prints a ZWJ family, a flag and
  a plain emoji, asks for the cursor position (ESC[6n) after each, and prints each advance in
  columns. The user runs it in tilix, foot, kitty, ptyxis and alacritty and pastes the output here.
- M2c (after the probe results): startup detection and `LOOM_ZWJ=join|split` override.
- M3: re-record the 096 frames for the user's ruler check.

M1 (e2d4f56), M2 (5937073) and M2b (3815045) delivered. `loom-probe` is installed.

Host review finding (contradicts M1's conclusion): the textrender sample source *does* contain
U+200D (`textrender.go:32`), and the frames are rewritten by textrender_test on every test run, yet
all six frames still have zero joiners. A host probe shows `Canvas.Write` and `loom.Render` keep the
joiner for plain text. So a widget used by the textrender views (button, list, clipping or scroll
path) drops it.

Pre-Work / Required Refinements for M3:
- Add a test per textrender view (borders, buttons, clipping, scroll) that renders it with
  loom.Render and asserts the output contains U+200D; find and fix the widget that drops it.
- Then commit the re-recorded 096 frames (they must contain U+200D).

M3 delivered (0a28fe9): the PTY screen capture split clusters; fixed, and the four 096 frames that show the family now contain U+200D. Waiting: the user's loom-probe output (M2c). Side finding filed as 135 (LOOM_EVIDENCE=1 breaks a filebrowser test).

## Probe results

- tilix: ZWJ family 8, flag 2, plain 2 (splits).
Pre-Work for M2c: loom-probe prints results while still in raw mode (lines staircase, no trailing newline; zsh shows `%`). Restore the terminal before printing, or use `\r\n` and end with a newline.
- foot: ZWJ family 2, flag 2, plain 2 (joins).
- kitty: ZWJ family 2, flag 2, plain 2 (joins).

- Also verify in M2c: the gap before `>` in docs/progress/096 M2-buttons.ansi in tilix (likely the same join/split width mismatch).
