# 359 — ColorPicker redesign: two-row FG/BG picker with slim edges (b2)

**Status**: Closed — M2 applies the palette-8 panel background, panel-colored edge glyphs, corrected row-3 layout and canvas coverage; design validation passed, q1 exited 0 with no FAIL markers, and make install succeeded.
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: UX
**Related**: 347 (one-row FG/BG mode picker, replaced here), 350 (∅ reset cell), 357 (Enter applies), 358 (arrows inside submenus)

---

/goal Replace the one-row FG/BG mode picker with the two-row design b2 from
`docs/data/colorpicker-ideas-001.ansi`, specced in spec/defaults.yaml, validated with `loom eval`,
`loom measure` and `loom check-box`, with key and PTY tests and docs/F1 help updated; or stop and
report when blocked on a user decision or denied permission.

## Design (user, 2026-10-22)

Variant b2 in `docs/data/colorpicker-ideas-001.ansi`:

- Top edge `▂` row with a coloured segment marking the FG pick; FG row; BG row; a `🮂` marker row
  for the BG pick; a 7+2 column preview "rug" (`Preview` in the chosen FG on the chosen BG) hanging
  below the panel; bottom `🮂` edge.
- FG row: swatches as coloured glyphs on the panel BG. BG row: swatches as BG cells. Swatches are
  not restyled by the other pick; only the preview shows the combination.
- Each row starts with `∅` (reset that row's colour) and shows all 16 palette colours (~35 columns).
- Edge chars spread the panel BG and clear the overlaid text so the popup stands out.

## Decisions

- Edge glyphs are spec values: default quarters (`▂` lower, `🮂` upper); documented fallback eighths
  (`▁`/`▔`) for fonts without `🮂` (Symbols for Legacy Computing, Unicode 13).
- Keys: Up/Down switch rows, Left/Right move within a row, Space applies the current row's pick and
  stays open, Enter applies both rows and closes picker and bar (selection cleared per 358).
- Mouse: click on a swatch applies it (as Space); no FG/BG mode toggle any more (347's Tab/b/
  right-click toggle goes away; update docs and Upgrading.md if it is public API).

## Notes

- Check live code first (347/350/357 implementation) before planning.
- Uncertain: width when the bar sits near the right terminal edge; reuse the popover clamping rules.

## M1 review (host, 2026-10-22)

Delivered a09c9f5 + fe6b21a: keys, Space/Enter, both channels, save work (live tmux: FG 1 on BG 2
saved). Host q1 green. Reopened for the look, which misses the core of the sketch:

### M2 pre-work

- The panel has no background: swatch rows, padding and preview pads use the editor BG (237 here),
  and the edge glyphs `▂`/`🮂` are drawn in the editor FG (250). In the sketch the panel has its own
  BG (palette 8) and the edge glyphs use that panel BG as their FG, so they "extend" the panel by a
  quarter cell and the popup stands out from the text. Add a spec'd panel BG (default palette 8, or
  the bar BG if that reads better; decide and document) and draw edges in it. Selection markers keep
  the marker colour.
- Marker row (row 3): columns 0–8 are panel BG (with the BG marker if the pick is there), columns
  9–34 are upper-edge glyphs, as in the sketch; today the whole row is `🮂`.
- Swatches: the sketch uses one letter plus a space per cell (`a `); `Aa` is fine if it reads
  better on the panel BG, but say which you kept.
- Add a canvas test asserting the panel BG on rows 1, 2 and 4 and the edge glyph FG; re-run
  `loom eval/measure/check-box`, regenerate docs/progress/ColorPicker.ansi, q1, install, close.
