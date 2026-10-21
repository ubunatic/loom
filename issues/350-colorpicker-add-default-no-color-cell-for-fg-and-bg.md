# 350 — ColorPicker: add default/no-color cell for FG and BG

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 347 (one-row FG/BG mode picker), 346

---

/goal Add a swatch that resets the colour to the terminal default ("default FG" in FG mode, "no BG"
in BG mode) to the ColorPicker, applied by RichTextEdit like any other pick; spec its glyph and
position, test it, update docs and `docs/progress/ColorPicker.ansi`, run `make test-q1` and
`make install`, or stop and report when blocked on a user decision or denied permission.

## Problem (user, 2026-10-21)

The picker offers only the 16 palette colours; there is no way to remove a FG or BG colour again.

## Decision (user, 2026-10-21)

Variant B (`docs/data/colorpicker-default-b.ansi`, 080ecd5): the default cell is the **first** cell and
shows `∅ ` (1-wide glyph plus space) in terminal default colours; FG mode resets to default FG, BG
mode to no BG. Palette swatches shift right by one; Home reaches the reset cell. Delete the unused
variant A file in the implementation commit.

## Notes

- Open: placement (first or last cell, 2 columns, row becomes 34 wide) and look (e.g. the `Aa`
  sample in default colours, or a `∅` glyph). Propose 1–2 .ansi variants before implementing.
- Needs a "default" colour value in the style model (check `Color` zero value) and in the saved ANSI
  output (reset code, not a palette index).
