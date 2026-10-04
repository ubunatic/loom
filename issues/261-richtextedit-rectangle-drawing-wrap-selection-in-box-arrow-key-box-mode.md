# 261 — RichTextEdit rectangle drawing: wrap selection in box, arrow-key box mode

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature
**Related**: `richtextedit.go`, `gallery/richtextedit.go`, issues 255 (format popover), 257 (keys, C-S-* decoding), 260; box glyph code in `examples/loomoji/loomoji/loomoji.go`, frame validator (128)

---

/goal Add box drawing to RichTextEdit in the milestones below, each with tests and a live gallery check, or stop and report when blocked on a user decision.

Re-check live code and recent commits first.

## M1 — Wrap selection in a box

Select text → format popover opens → choose "Box" → the selected text is wrapped in a rectangle of box-drawing glyphs (`┌─┐│└┘`). Multi-line selections get one box around the block, width = widest line.

## M2 — Box mode with arrows

`C-S-b` toggles "box" mode (shown in the hint bar; Esc also exits). In box mode each arrow key draws: the glyph under the cursor is assessed, the next cell in the arrow direction gets a line glyph, and the cursor moves there. Corners and junctions (`┌┐└┘├┤┬┴┼`) are completed correctly when a line turns or crosses an existing one.

## M3 — Four-neighbour completion

When placing a glyph, look at all 4 neighbours (up/down/left/right) and choose the glyph whose arms connect to every adjacent line arm, so crossings and T-junctions with existing boxes come out right.

## Notes / uncertainties

- RichTextEdit is a flow document; drawing below/right of line ends needs padding with spaces, and drawing up/down needs column alignment by display width (wide runes). Decide and document how lines are extended.
- Glyph-from-arms lookup (4-bit mask → glyph) should be a small reusable library helper, not RichTextEdit-only; check loomoji and existing box code for one to reuse.
- Light single-line glyphs only for now; heavy/double/rounded styles are out of scope unless trivial.
- Undo: one box-mode stroke (enter to exit) or one wrap = one undo step.
- Verify `C-S-b` decodes (CSI-u only, per 257); legacy terminals send plain C-b. Note the fallback key if needed.

## Progress

- **M1 delivered (f80aeb9): wrap selection in box.** "Box" popover action, reuses `getBoxBorderGlyphs`, one undo step. Host rerun: RichTextEdit tests green; only failure is the known flaky `TestPaneFirstDrawUsesScreenBounds` (issue 247).

### M2 Pre-Work / Required Refinements

- Mid-line selection: currently the prefix text is glued before the top border (`abc┌──┐`) and the suffix after the bottom border, so the box columns don't line up with the side borders. Put prefix and suffix on their own lines (as planned), so all box rows start in the same column. Add a test for a selection that starts and ends mid-line.
- **M2 delivered (57c1059): arrow-key box mode** (F5 / C-S-b, Esc exits, one undo per stroke; mid-line wrap pre-work done). Host rerun green; live tmux check drew a closed 4×3 box with correct corners, including the closing corner, and padded the short rows.

### M3 Pre-Work / Required Refinements

- None from M2. For M3 add a live-style test that crosses an existing horizontal line vertically (expect `┼`) and one that ends a stroke on an existing side (expect `├`/`┤`/`┬`/`┴`).
