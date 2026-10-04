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

## Push rules (user, 2026-10-04)

A "natural" push: text gets out of the way of the stroke, spaces are drawn over.

- **Vertical stroke (Up/Down) hits a non-space:** the word under the stroke moves right, out of the box path (ideally the whole word, minimum the character), and the rest of that line shifts right with it. Example: three lines of text with no blank lines between; a vertical stroke through them pushes each line's text to the right as the stroke reaches it.
- **Horizontal stroke (Left/Right) hits a non-space:** the whole line moves down (a blank row is inserted), so the stroke continues in free space. Continuing over spaces pushes nothing.
- Moving onto a cell that is a space never pushes. After a push, the stroke paints through free cells until it hits the next non-space, which triggers the next push.
- Pushed text keeps its styles. A word that would be cut by a vertical stroke moves as a whole to the right of the stroke.

### Pushing through existing boxes (user, 2026-10-04)

The document model has no decoration layer: box glyphs are ordinary characters in `RichLine` spans, recognised only by glyph (`BoxGlyphArms`). Rule for v1:

- A box glyph under the stroke is **not pushed**; the stroke joins it with a junction (as in F5 mode).
- Text characters are pushed per the rules above. Example: an existing box around three lines of text, a vertical push stroke from the top through the middle: top and bottom edges get junctions (`┬`/`┴`), the text rows inside are pushed right, which shifts the outer box's right edge on those rows and breaks it. Accepted for the first version.
- Follow-up (not in this ticket): compensate broken boxes, e.g. widen the outer box so its right edge stays aligned. File it after the live check if still wanted.

Uncertain (user): the vertical case "depends on the situation"; implement the rules above first and refine after a live check.

## Acceptance

- Unit tests: horizontal and vertical pushes, spaces overpainted not pushed, styled characters keep their style, junctions with existing boxes, wide characters, undo/redo.
- Live gallery check by the user in Tilix (S-F5 arrives as `ESC[15;2~`).
