# 294 — loom edit: Add ^O to open file pane and ^W to close file

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Feature
**Category**: Keybindings
**Related**: [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [282](282-add-f2-side-panel-with-file-browser-to-loom-edit.md), [290](290-keep-the-loom-edit-file-browser-usable-after-cancel-and-mouse-activation.md)

---

## 1. Problem & Motivation
`loom edit` lacks the standard editor keys `^O` (open: show and focus the file browser) and `^W` (close the current file, with the unsaved-changes guard). The File menu also lacks Open/Close entries.

## 2. Technical Specification / Findings
- Today the browser opens only via F2 (`spec/defaults.yaml` `editor.hotkey_files_*`, `cmd/loom/edit.go:206` `toggleSidePanel`). There is no `ctrl-o`/`ctrl-w` handling anywhere in `richtextedit*.go` or `edit.go`.
- `spec/widgets.yaml:166` already shows `{Label: "Open", Shortcut: "Ctrl+O"}` as a menu example, but no binding backs it.
- Opening a file with a dirty buffer already goes through the guard via `pendingOpenPath` (`edit.go:152-156`, `:237-272`). That guard logic is app-local and duplicated for quit.
- Fix in the library: `RichTextEdit` defines document actions Open and Close (next to Save and Save-as). It emits requests (`OnOpenRequest`, `OnCloseRequest`) that a host fulfils, so the editor does not own a side panel. Close runs the shared unsaved-changes guard from 297 (Save / Discard / Cancel). After close the editor shows an empty Untitled buffer. Bindings, keycaps and labels go in `spec/defaults.yaml` (`pane.hotkey_open_*`, `pane.hotkey_close_*`). The file bar menu (293) lists them.
- `^O` while the browser is shown moves focus to it (it does not toggle it off). F2 remains the toggle.

## 3. Implementation & Verification Plan
- Add the spec keys, the editor actions, the callbacks and the menu items. In `loom edit`, wire Open to show and focus the browser and Close to the guarded reset.
- Make sure `^O`/`^W` work from every focus (editor, browser, search) through normal routing, not per-branch checks in `edit.go`.
- Tests: unit (actions fire, guard on dirty Close, Cancel keeps the buffer); PTY (`^O` focuses the browser; `^W` on a dirty buffer shows the dialog, Discard gives an empty Untitled buffer).
- `make test-q1`, `make install`.

/goal Add spec-bound Open (^O) and Close (^W) document actions to RichTextEdit with host callbacks and the shared unsaved-changes guard, wire them in loom edit and its File menu, and verify with unit and PTY tests, or stop and report when blocked on a user decision or denied permission.
