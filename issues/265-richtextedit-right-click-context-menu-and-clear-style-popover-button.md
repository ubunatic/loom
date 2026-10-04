# 265 — RichTextEdit right-click context menu and clear-style popover button

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature
**Related**: `richtextedit.go`, issues 255 (popover, color picker), 257 (internal clipboard, C-space), 264 (popover keyboard navigation, submenus)

---

/goal Add a right-click context menu and a clear-style popover button to RichTextEdit, with tests and a live gallery check; stop and report when blocked on a user decision.

Re-check live code and recent commits first. Build on 264's popover/submenu keyboard navigation (Tab/Enter/Esc) if it has landed; do not fork a second menu model.

## Pre-Work: color picker focus marker (user, 2026-10-04, after 264 closed)

The keyboard focus in the #FG/#BG color picker (264 M4) is hard to see: a highlight moving over a bar of colors does not stand out. Mark the focused swatch with foreground helper characters in the neighbouring cells, e.g. `>█<`: the cells left and right of the focused swatch show `>` and `<` (in a contrasting fg), instead of relying on a color/inverse highlight. Keep the swatch itself unchanged. If the picker's layout has no free neighbouring cells, add one-cell gaps or draw the markers over the neighbouring swatches; choose and document. Test the rendered row for focus at the first, middle and last swatch.

## Right-click context menu

Right-click in the editor opens a menu at the click position with:

- **Copy / Cut / Paste**: the existing internal clipboard (257). Right-click on a selection keeps it; on unselected text it moves the cursor there first.
- **Paste style**: apply the style of the clipboard content to the selection (or word at cursor) without changing its text. If the clipboard has mixed styles, best-effort guess (e.g. the style covering the most characters). Disabled when the clipboard is empty.
- **Change case** -> submenu: lower, upper, title, snake_case, PascalCase. Applies to the selection (or word at cursor), keeps span styles where the text length allows, one undo step.

Esc or a click outside closes it. Keyboard navigation follows the popover (264).

## Popover: clear style

Add a "🧹" button to the main format popover that removes all styling (bold, italic, underline, fg/bg colors) from the selection. Links keep their link target. One undo step.

## Open questions

- Snake/Pascal conversion of multi-line or multi-word selections: convert per line, and treat spaces/punctuation as word breaks (proposal).
- Does 🧹 render with width 2 in all target terminals? Measure with the existing display-width code; fall back to a text label if not.

## Acceptance

- Unit tests for each action, incl. empty clipboard, mixed-style paste-style guess, case conversions with punctuation and non-ASCII, undo/redo.
- Live gallery check by the user in Tilix.
