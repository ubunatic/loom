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

## 4. More terminals (user screenshots in ansiviewer, 2026-09-27)
- alacritty: same as tilix. The flag is drawn as the letters "DE", and the flag row ends early with
  the default background at its end.
- foot: the flag is drawn as a glyph, and the flag row still leaves an unpainted cell at its end.
- ptyxis: the gap is there too, and in addition every sample row is drawn in the wrong order: the
  sample text comes first and the label second ("DE Flag [...]" instead of "Flag DE [...]"). This
  points to ansiviewer's cursor replay (column moves) and not only to width. Check whether the
  frame uses absolute column moves (CSI G / CSI H) that ansiviewer replays differently from ptyxis.

## 5. Plan (dev-132, reviewed by the host)
- M1: a test that replays a 096 frame through ansiviewer and asserts label/sample order and cell
  positions around the flag. The test comes first and fails.
- M2: `writeANSI` replays whole grapheme clusters (a regional-indicator pair is one cluster) and
  keeps rectangle clipping; the trailing cells of the row get the widget background.
- M3: fix cursor replay (CSI G / CSI H) only if M1 shows it is wrong.
Pre-work: use `make test-q1` once per change, write its output to a file, and grep it for `--- FAIL`.
