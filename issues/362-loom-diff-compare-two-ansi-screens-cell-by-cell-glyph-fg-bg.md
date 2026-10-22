# 362 — loom diff: compare two ANSI screens cell by cell (glyph, FG, BG)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 359–361 (ColorPicker colour bugs found only by per-cell checks), harnez 852 (tui-check skill)

---

/goal `loom diff <want.ansi> <got.ansi>` lists every cell whose glyph, FG or BG differs (row, column by display
width, want vs. got), exits non-zero on differences, and is documented and tested; or stop and report when
blocked on a user decision or denied permission.

## Why

In 359–361 the host compared tmux captures with the sketch by hand (`capture-pane -pe` plus a perl escape
dump). A loom command makes that one call for hosts, developers and the harnez `/tui-check` skill.

## Notes

- Inputs: `.ansi` sketches and `tmux capture-pane -pe` output; reuse loom's existing ANSI parser
  (as `loom eval`/`measure`).
- Useful options to consider: a region (row/column window, since a sketch may show only the widget) and
  ignoring glyph or colour.
