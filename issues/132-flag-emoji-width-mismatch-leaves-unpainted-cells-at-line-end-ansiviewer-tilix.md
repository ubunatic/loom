# 132 — Flag emoji width mismatch leaves unpainted cells at line end (ansiviewer, tilix)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 096, 048, 102

---

## 1. Problem & Motivation
In ansiviewer on tilix (user screenshot, 2026-09-27), the "Flag DE" line of the 096 textrender
frame ends about 2 columns early, and the terminal's default background shows through in the gap
at the right edge. The ZWJ and mixed lines above and below fill the full width. A line that does
not paint its full row breaks the look of any themed Loom app that shows flag text.

## 2. Technical Specification / Findings
- 096 recorded the divergence `flag: Loom 4, terminal 2`. Tilix, ptyxis and alacritty draw the
  flag as two letters "DE"; foot and kitty draw a flag glyph.
- Likely cause: Loom reserves 4 columns for the flag and paints only the cells it thinks are free.
  The terminal advances only 2, so the row's last 2 cells are never written and keep the default
  background.
- Unconfirmed: whether this comes from ansiviewer's own SGR/cursor parser (`applySGR`/`writeANSI` in
  examples/ansiviewer/ansiviewer/viewer.go) or from Loom's cell/width model.

## 3. Implementation & Verification Plan
1. Reproduce with the 096 frames: compare `cat` and ansiviewer output of the flag line in tilix.
2. Decide the fix: count a regional-indicator pair as 2 columns (terminal consensus), or always paint
   or clear the cells past a wide cluster so any mismatch shows the widget background.
3. Add a regression test in which the flag line's row is fully painted with the widget background.
4. The user checks it in tilix.
