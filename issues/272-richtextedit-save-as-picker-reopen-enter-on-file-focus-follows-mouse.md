# 272 — RichTextEdit Save as picker: reopen, Enter on file, focus follows mouse

**Status**: Open
**Priority**: P1 (High)
**Severity**: Moderate
**Category**: Bug
**Related**: [269](269-refine-richtextedit-hints-picker-focus-and-help.md), [273](273-audit-richtextedit-workarounds-and-plan-loom-library-fixes.md)

---

## 1. Problem & Motivation
User-reported problems with the RichTextEdit Save as FilePicker (installed gallery, after issue 269 M1–M6):

1. After saving once through the picker, the picker does not open again.
2. Open picker → type `test` → select the existing `test.ansi` → Enter only moves focus to the filename input instead of choosing that file.
3. The input cursor follows the mouse. This is not wanted. Remove the workarounds that cause it; file separate Loom library tickets if the proper fix needs library changes (rule: fix the library, not the caller).

## 2. Technical Specification / Findings
Starting points (verify against live code first):
- (1) Once a picker save sets `FilePath`, `Save()` (Ctrl+S) writes directly and never opens the picker; that part is intended. Check whether "Save as…" (menu and Ctrl+Shift+S) still reopens it, and whether picker/popup state (`savePicker`, `savePopup`, `FilePicker.done`, `Popup.Open`) is left stale after `OnSave`. Reproduce the user's exact sequence first.
- (2) User decision: Enter on a listed file puts that file's name into the filename input, moves focus to the input, and places the text cursor in it (end of the name). It does not save; a second Enter in the input saves. Today `FilePicker.activate()` sets the value and `nameFocus = true` but never calls `updateFocus()`, so the list keeps its focus look and the cursor does not move to the input.
- (3) `FilePicker.ConsumeMouse` switches `nameFocus` on every mouse event, including `MouseHover`, so moving the pointer over the name row or the list moves focus and the cursor. Also check `Choice` hover selection (`c.sel = fi` on hover) and the pane's cursor-proximity cursor placement (`pane.go`, `SetCursorPosition(p.mouseX, p.mouseY)`). Focus should change only on click or keys.

Event-routing change: per AGENTS.md, start on `codex:sol:med`.

## 3. Implementation & Verification Plan
/goal Save as reopens reliably, Enter on a listed file fills and focuses the filename input with the cursor in it, and focus/cursor change only on click or keys, with library-level fixes where needed; stop and report when blocked on a user decision or denied permission.

Acceptance: regression tests per item (reopen after picker save; Enter on a listed file fills the input, focuses it and places the cursor at the end of the name; hover does not change focus or cursor); installed-binary PTY check of the user's sequence; any needed library change gets its own ticket linked here.

---

## Delivered

- **Save as Reopen (272a)**: Cleared `savePicker` and `savePopup` when `savePopup.Open` becomes false or Escape/outside click closes the popup.
- **Enter on File in List (272b)**: `FilePicker.activate()` in save mode calls `updateFocus()`, focusing filename entry, positioning the caret at the end of the filename, and shifting focus state from the directory list. A second Enter in the input saves.
- **Hover Focus Stability (272c)**: `FilePicker.ConsumeMouse` only toggles focus on left mouse press, preventing `MouseHover` from stealing focus or moving the cursor.
- **Commits**: `8a41b2a`
- **Tests**: `make test-q1` passed. Unit and PTY regressions verified.
