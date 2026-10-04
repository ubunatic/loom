# 264 — RichTextEdit popover keyboard navigation, box selection and box style submenu

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature
**Related**: `richtextedit.go`, `boxglyph.go`, `BoxBorderStyleSharp/Rounded` (existing border styles), issues 255 (popover, color picker), 257 (C-space), 261 (box drawing), 263 (Home/End)

---

/goal Make the RichTextEdit format popover fully keyboard-driven and box-aware per the milestones below, each with tests and a live gallery check; stop and report when blocked on a user decision.

Re-check live code and recent commits first. Stop for host review after each milestone.

## M1 — C-space on box art selects the box

When the cursor is on box glyphs (a box from 261), `C-space` selects the whole box (its bounding rectangle, found by following connected glyph arms) and opens the popover. Off a box, `C-space` keeps selecting the word.

## M2 — Tab/Enter in the popover bar

With the popover open, `Tab` / `S-Tab` cycle the bar items with a visible focus highlight, `Enter` applies the focused item, `Esc` closes. Initial focus follows the selection: bold → "B", italic → "I", underline → "U", box → "Box", else the first item.

## M3 — Box style dropdown

"Box" opens a dropdown with **plain** (sharp, `┌┐└┘`) and **rounded** (`╭╮╰╯`). Choosing a style wraps the selection, or restyles the selected box (M1). Reuse the existing `BoxBorderStyle*` glyph sets; labels/defaults go in the spec per `docs/Spec.md`.

## M4 — Keyboard navigation in submenus

Flow: `C-space` → bar opens → `Tab`… → `Space` opens the focused item's submenu (FG/BG color picker or box style) → `Tab`/arrows… → `Enter` applies → back to text; `Esc` steps back one level.

## Notes / uncertainties

- Box detection on a flow document: decide how a box with text inside or with crossing lines (┼) is bounded; document the rule.
- Box-mode junction completion (261 M3) must handle rounded corners: `╭` etc. need the same arm mask as `┌`.
- Tab inside the popover must not insert a tab into the text or leak to the host's focus cycling (event-routing; per AGENTS.md start on `codex:sol:med`).
