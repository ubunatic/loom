# 316 — Configurable popup menu focus, navigation, and dismissal semantics in loom edit

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 307

---

## 1. Problem & Motivation
In `loom edit` / `RichTextEdit`, selecting text currently causes the format popover toolbar to appear and immediately capture keyboard navigation:
- Pressing Left/Right immediately navigates the popover actions instead of letting the user move the editor cursor or cancel the menu.
- A user selecting text often expects arrow keys to cancel or navigate the text unless they explicitly opt into menu focus (e.g., via `<tab>` or mouse interaction).
- Once actions are executed, or when dismissed, behavior around menu dismissal and selection retention needs to be predictable and spec-configurable.

## 2. Technical Specification & Proposed Config Keys
Gate these behaviors in `spec/` (e.g. `spec/defaults.yaml` under a popover/editor section) with Go struct bindings in `RichTextEdit`:

1. **`action_closes_menu`** (bool, default `true`):
   Whether applying an action (color, style, box style) automatically closes the popover.
2. **`menu_focus_key`** (string, default `"tab"`):
   Key required to transfer keyboard focus from the text selection to the popover toolbar when the menu is visible but unfocused.
3. **`arrows_cancel_unfocused_menu`** (bool, default `true`):
   When the popover is visible but unfocused, pressing arrow keys dismisses the popover and passes the arrow key to editor cursor movement.
4. **`unselect_after_action`** (bool, default `false`):
   Whether text selection is cleared after applying an edit action.
5. **`unselect_on_cancel`** (bool, default `false`):
   Whether text selection is cleared when dismissing/cancelling an unfocused menu or pressing Escape.
6. **`popup_open_key`** (string, default `"ctrl-space"`):
   Key to reopen or refocus the popover toolbar while text remains selected.
7. **`auto_popup_on_selection`** (bool, default `true`):
   Whether the popover appears automatically upon selecting text or only upon pressing `popup_open_key` / context click.
8. **`popup_open_delay_ms`** (int, default `0`):
   Debounce duration before popping up during mouse drag selection to avoid flicker.
9. **`cycle_navigation`** (bool, default `true`):
   Whether arrow/tab keys wrap around from the last item to the first item in the menu/palette.
10. **`position_preference`** (enum: `above`, `below`, `smart`, default `smart`):
    Preferred vertical placement relative to the selection rectangle.
11. **`dismiss_on_scroll`** (bool, default `true`):
    Whether editor scrolling automatically dismisses the popover.
12. **`mouse_hover_focus`** (bool, default `false`):
    Whether hovering over a popup item transfers keyboard navigation focus.
13. **Submenu & Escape Navigation**:
    Mouse selection of multi-stage items (e.g. FG/BG palette, Box choice) opens submenus; in focused popups and submenus, arrow keys navigate items and Escape cancels back one stage until final dismissal.

## 3. Implementation & Verification Plan
- Expose the configuration schema in `spec/defaults.yaml` and validate via `cmd/validate-spec`.
- Wire settings into `RichTextEdit` / `editView` in `richtextedit.go` and `cmd/loom/edit.go`.
- Write unit tests and PTY interaction tests in `cmd/loom/edit_*test.go` verifying:
  1. Arrow keys cancel unfocused menu and move cursor when `arrows_cancel_unfocused_menu: true`.
  2. Pressing `<tab>` focuses the menu, allowing arrow navigation.
  3. Action execution respects `action_closes_menu` and selection clearing flags.
  4. Pressing `^<space>` reopens the menu while text remains selected.
  5. Extended spec gates (`auto_popup_on_selection`, `cycle_navigation`, `dismiss_on_scroll`, `position_preference`) take effect as specced.
- Validate with `make test-q1`.

### Pre-Work for Milestone 2 (Test & Isolation Fixes):
1. **Fix Escape Selection Clearing in `editView.ConsumeKey` (`cmd/loom/edit.go`)**:
   - `TestEditEscapeClearsSelectionAfterOverlayCloses` failed because `res.Consumed` check blocked clearing selection on the second Escape. Restore `!wasBoxMode && key.Is("esc") && v.focused == focusEditor && v.edit.HasSelection` so second Escape clears the selection after any overlay closes.
2. **Fix Mouse Backdrop & Isolation in `RichTextEdit.consumeMouse` (`modal_mouse_test.go` & `edit_304_test.go`)**:
   - `Test304PopoverMouseIsolation` tests backdrop clicks and asserts `wantOpen := action != MousePress || button != MouseLeft`.
   - `dismiss_on_scroll: true` dismissed on `MouseScrollUp`/`MouseScrollDown` inside backdrop handling, breaking `Test304PopoverMouseIsolation` scroll assertion. Ensure backdrop mouse events outside popover hit bounds follow the established backdrop contract: only left press dismisses the popover. Scrolling inside the editor text area should dismiss on scroll when configured.
3. **Fix PTY Popover Backdrop Test Timing / Gestures (`edit_304_test.go`)**:
   - Ensure `s.Send("\x13\x1bOP")` in `Test304EditPTYPopoverBackdropIsolation` does not hang or race with menu focus/reopen.
4. **Run `make test-q1` and verify 100% green suite.**

/goal Implement configurable popup menu focus, navigation, and dismissal semantics gated by spec in RichTextEdit and verify with unit tests, or stop and report if blocked on a user decision or denied permission.
