# 302 — loom edit: UX refinements, theme default, mouse disable, backdrop clicks, and untitled buffer

**Status**: Open
**Priority**: P1 (High)
**Severity**: Feature
**Category**: UX
**Related**: [292](292-loom-edit-f1-help-screen-looks-cluttered.md), [295](295-loom-edit-add-s-as-save-as-and-use-as-modifier-hints.md), [300](300-loom-edit-replace-top-title-bar-with-compact-bottom-status-icons.md), [301](301-loom-edit-refine-hotkeys-nest-draw-under-box-and-make-bottom-bar-items-clickable.md)

---

## 1. Problem & Motivation
Several UX inconsistencies, defaults, and interactions across `loom edit` and library overlay widgets need refinement:
1. **Altscreen Toggle Extra Line**: In inline mode (`"a"`), an extra line appears at the bottom below the app compared to fullscreen/altscreen mode (`"A"`).
2. **Default Theme**: `julia256` should be the default theme for `loom` commands and apps in general.
3. **Mouse Disable ("m" vs "M")**: The `"m"` status mode should fully disable terminal mouse grab/tracking so that native terminal selection (drag-to-copy) works unobstructed. Currently, "m" mode still captures click/drag events inside the app.
4. **Backdrop Clicks on Popups & Dialogs**: Clicking outside a popup (such as F1 help) should close the popup. For `Dialog`, clicking outside the dialog window should trigger the default "Cancel" option.
5. **Box Draw Mode Exit**: Pressing `<Enter>` should finish and exit box drawing mode.
6. **Remove `⌃⌥S` Save As Binding**: Remove the `⌃⌥S` hint from the help screen and unbind the shortcut completely; Save As is strictly a File-menu action.
7. **Mouse Scroll in Help Panes**: Standard help popups/panes must support mouse wheel scrolling by default without per-widget workarounds.
8. **No-argument `loom edit`**: Running `loom edit` without arguments should open an empty, Untitled buffer rather than failing or requiring a file path.

## 2. Technical Specification / Findings
- **Inline Mode Row Calculation**:
  - Check `pane.go` inline mode height clamp (`termRows - 1` vs `termRows`) and `cmd/loom/edit.go` resize logic to ensure inline mode does not produce an extra trailing empty row.
- **Default Theme (`spec/defaults.yaml`, `defaults.go`)**:
  - Update specced default theme across CLI apps to `julia256`.
- **Mouse Tracking Disabling**:
  - In `cmd/loom/edit.go` and `pane.go`, distinguish full mouse tracking (`EnableMouse()`), click-only (`EnableMouseClicks()`), and full disable (`DisableMouse()` / `\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l`). When mouse grab is off (`"m"`), disable mouse reporting completely so the terminal handles selections natively.
- **Popup & Dialog Backdrop Click Handling**:
  - In `Popup.ConsumeMouse`: if mouse click is outside popup bounds, dismiss / close popup.
  - In `Dialog.ConsumeMouse`: if mouse click is outside dialog rect, execute the "Cancel" action.
- **Box Draw Mode (`richtextedit.go`)**:
  - In `RichTextEdit.ConsumeKey`, handle `key.Is("enter")` while in `BoxMode` by turning off `BoxMode` and consuming the key.
- **Unbind `⌃⌥S`**:
  - Remove `ctrl-alt-s` binding and help table entry from `spec/defaults.yaml` and `RichTextEdit`.
- **Mouse Wheel Scroll in Help**:
  - Ensure `Popup` and `Help` widgets process `MouseScrollUp` and `MouseScrollDown` events when mouse reports are active.
- **`loom edit` CLI Invocation (`cmd/loom/edit.go`)**:
  - Make the file path argument optional (`cobra.MaximumNArgs(1)`). When omitted, initialize `newEditView` with `loom.NewRichTextEdit()` on an empty Untitled buffer.

## 3. Implementation & Verification Plan
- Update `spec/defaults.yaml`, `defaults.go`, and schema files.
- Update `RichTextEdit`, `Popup`, `Dialog`, `Pane`, and `cmd/loom/edit.go`.
- Add unit and PTY tests covering:
  - `loom edit` without args opening Untitled document.
  - Altscreen vs inline row height.
  - Enter exiting box draw mode.
  - Popup and Dialog outside-click dismissal / cancel.
  - "m" mode disabling mouse tracking in terminal.
  - Mouse scroll inside F1 help.
- Verify with `make test-q1` and `make install`.

/goal Implement the 8 UX refinements for loom edit, themes, mouse modes, popup backdrop clicks, draw exit, and untitled buffer, and verify with make test-q1, or stop and report when blocked on a user decision or denied permission.
