# 266 — RichTextEdit S-F5 box-push drawing mode

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Normal
**Category**: Feature
**Related**: `richtextedit.go`, `boxglyph.go`, issues 261 (box drawing mode, F5), 264 (box selection, draw button, cursor shape)

---

/goal Add a box-push drawing mode on S-F5 where box strokes push non-space characters out of the way instead of overpainting them, with tests and a live gallery check; stop and report when blocked on a user decision.

Re-check live code and recent commits first (264 changed F5 and draw mode).

## Behavior

- S-F5 toggles box-push mode (F5 stays the overpainting box mode). Same block cursor as draw mode; the popover/hint shows which mode is active.
- When a stroke enters a cell holding a non-space character, that character (with its style) is moved out of the way instead of being replaced. Space cells are overpainted as in normal box mode.
- Existing box glyphs are joined as today (four-neighbour junctions), not pushed.
- One undo step per stroke, including the pushed characters.

## Open questions

- Push direction: horizontal strokes push characters right along the row, vertical strokes push them down (insert a line or shift into the next row)? Proposal: push in the stroke direction, inserting padding; record the chosen rule.
- What happens when pushing would hit another box or the line end: shift the whole tail of the line (proposal) or stop the stroke.

## Acceptance

- Unit tests: horizontal and vertical pushes, spaces overpainted not pushed, styled characters keep their style, junctions with existing boxes, wide characters, undo/redo.
- Live gallery check by the user in Tilix (S-F5 arrives as `ESC[15;2~`).
