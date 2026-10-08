# 312 — Preserve .tui block source (box, table, image) on save

**Status**: Draft
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 309 (boxes), 310 (tables), 311 (images)

---

## 1. Problem & Motivation
`ParseMarkup` draws `[box]`, `[table]` and `[image]` blocks into plain
`RichLine`s at parse time. `RichDocument` keeps no block structure, so
`ToMarkup`, and saving a `.tui` file in `loom edit`, writes the drawn glyphs
instead of the block tags. For images this is worst: every cell becomes a
`[#rrggbb on #rrggbb]▀[/]` run, the file grows a lot and the `src` link is
lost. Editing box content also does not re-flow the border.

## 2. Technical Specification / Findings
- Block rendering: `markupBox.render` (`markup_box.go`),
  `markupTable.render` (`markup_table.go`), `renderMarkupImage`
  (`markup_image.go`).
- Accepted as a known limit in issue 311 §2.4 and `docs/TuiMarkup.md`.
- Options to evaluate: keep a block-source range on the lines a block
  produced and write the source back when those lines are unchanged; or
  model blocks in `RichDocument` and re-render on edit.

## 3. Implementation & Verification Plan
To be decided after a design pass. Verification: a load/save round trip of
`examples/markup/{box,table,image}.tui` without edits reproduces the source.
