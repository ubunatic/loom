# 318 — Cancel focused popover on arrow away from menu bar

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 316, 317

---

## 1. Problem & Motivation
In `RichTextEdit`, directional arrow focus was added so:
- Arrow towards the popover (Up when above, Down when below) focuses the main menu bar.
- Once the main menu bar is focused, arrow keys navigate horizontally between menu actions.

However, once focused on the main menu line (when no submenu is open):
- If the menu is **below** and the user presses **Up** (moving away from the popover, back toward the text/selection), the menu currently stays focused or ignores the vertical arrow.
- If the menu is **above** and the user presses **Down** (moving away from the popover, back toward the text/selection), the menu currently stays focused or ignores the vertical arrow.

Pressing the vertical arrow directed away from the focused main menu bar back towards the editor should cancel/dismiss the popover cleanly and return control to the editor.

## 2. Technical Specification & Findings
In `richtextedit.go` (`consumePopoverKey`):
- When `e.popoverFocusSet == true` and no submenu is open (`!submenuOpen`):
  - Determine whether the popover toolbar is positioned **above** or **below** the anchor:
    - If positioned **below** (`bar.Y > anchorY` / `pointer == PointerUpGlyph`) and key is **`Up`**:
      cancel the popover (`e.cancelPopover()`) and let editor navigation handle the key or consume as dismissal.
    - If positioned **above** (`bar.Y < anchorY` / `pointer == PointerDownGlyph`) and key is **`Down`**:
      cancel the popover (`e.cancelPopover()`) and let editor navigation handle the key or consume as dismissal.
  - Submenu navigation (e.g. FG/BG palette swatches or Box choices) is unaffected: vertical/horizontal arrows inside submenus continue to wrap/navigate choices as specced until Escape or action is taken.

## 3. Implementation & Verification Plan
- Update `RichTextEdit.consumePopoverKey` in `richtextedit.go`.
- Add test cases in `richtextedit_test.go` and `cmd/loom/edit_317_test.go` / `edit_318_test.go`:
  - Focused menu below + Up -> cancels popover.
  - Focused menu above + Down -> cancels popover.
  - Submenu open + Up/Down -> navigates submenu items normally without cancelling.
- Verify with `make test-q1`.

/goal Implement cancel on arrow away from focused main menu bar (Up when menu is below, Down when menu is above) in RichTextEdit and verify with tests, or stop and report if blocked on a user decision or denied permission.
