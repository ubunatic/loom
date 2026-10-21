# 359 — ColorPicker redesign: two-row FG/BG picker with slim edges (b2)

**Status**: Open
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
