# 360 — ColorPicker look: bar panel colour, row-3 gap BG, active-row marker in selection colour, right padding

**Status**: Closed — M2 fixed edge glyph backgrounds to use editor BG and limited selection/dim marker colors to the two picked cells; per-cell canvas tests cover edges and picks inside and beyond the preview rug. ANSI checks and q1 passed; make install succeeded.
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 359 (two-row picker b2), 353 (Ctrl+Space bar), 355 (bar focus highlight)

---

/goal Fix the four ColorPicker look bugs below, with the colours specced (no Go duplicates of spec
values), canvas tests per bug, `loom eval/measure/check-box` clean and docs/progress/ColorPicker.ansi
regenerated; or stop and report when blocked on a user decision or denied permission.

## Bugs (user, 2026-10-22, after 359)

1. **Panel colour:** the picker uses `color_picker.panel_bg: 8`; it must use the same panel colour as
   the Ctrl+Space bar (`toolbar_bg: 239` in spec/defaults.yaml today). Edge glyphs follow it.
2. **Row-3 gap:** on the BG marker row, the cells between the preview rug and the picks (columns
   left of the upper edge, where no marker is) are transparent (editor BG). They must be panel BG.
3. **Active row not visible:** the FG and BG markers both use `marker_fg: 51`, so the user cannot see
   which row Left/Right/Space act on. The active row's marker uses the standard selection colour;
   the inactive row's marker is dimmed.
4. **Right edge:** the swatch rows end right after the last swatch; add one panel-BG padding column
   after it (mirror of `leading_padding`), and the edges span the new width.

## Decisions (host)

- "Standard selection colour" = the colour the editor selection and the focused bar item use today
  (`selection_bg: 24`, `popover_focus_bg: 24`). If the theme system (`theme: "julia256"`) has no
  named selection role, add one (e.g. `selection`) and point `selection_bg`, `popover_focus_bg` and
  the picker's active marker at it, so the colour is set in one place. Check the live theme code
  first; record what you found in this ticket.
- Inactive marker: a dim colour on the panel BG (e.g. `separator_fg: 8`), specced.
- Panel colour: reference the bar's colour rather than copying 239 into a second key; remove
  `color_picker.panel_bg` or make it default to the bar colour. Add an Upgrading.md row if a public
  key goes away.
- Live check by the host in tmux against the sketch `docs/data/colorpicker-ideas-001.ansi` (b2).

## M1 review (host, 2026-10-22)

0a60961: host q1 green. Live tmux (80x20, BG pick at palette 9, active row BG) shows:

### M2 pre-work

- **Edge rows use the panel colour as BG.** Rows 0 (`▂`), 3 right part (`🮂`) and 5 (`🮂`) are drawn
  `fg 239 on bg 239`, i.e. solid panel blocks; the slim look is gone. In the sketch the edge glyph
  is `fg panel` on the **editor/default BG** (the glyph itself extends the panel by a quarter cell;
  its BG clears the text under it). Only cells inside the panel (row 3 columns 0–8, rows 1, 2, 4) get
  panel BG.
- **Marker colour leaks along row 3.** The whole upper-edge run from column 9 to 35 is fg 24 (the
  selection colour), not only the two cells under the BG pick (columns 21–22 for palette 9). Every
  other edge cell stays fg panel. Same rule for row 0.
- Canvas tests must assert per cell: edge glyph FG and BG on rows 0, 3, 5, and that only the pick's
  two cells carry the marker colour (pick inside and right of the rug, active and inactive row).
