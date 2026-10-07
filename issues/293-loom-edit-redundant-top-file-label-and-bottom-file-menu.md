# 293 — loom edit: Redundant top File label and bottom File menu

**Status**: Closed — implemented and verified in e0d7442
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: UX
**Related**: [268](268-add-save-and-save-as-to-ricktextedit-with-a-file-menu.md), [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [288](288-compose-loom-edit-panels-with-standard-widgets-and-clipped-responsive-layout.md)

---

## 1. Problem & Motivation
`loom edit` shows "File" twice: a static `File` word at the start of the top header, and the interactive `File` menu in the bottom file bar. The top label does nothing and suggests a second menu.

## 2. Technical Specification / Findings
- `cmd/loom/edit.go:394`: `headerText := fmt.Sprintf(" File    %-48s %s", fileName, rightStatus)`. The static label is a caller-side literal.
- `richtextedit_filebar.go:17-30`: `ensureFileBar` builds `NewMenuBar(Menu{Title: "File", ...})` with `Bottom = true`. This is the real menu. `:71-87` draws it next to `status()` ("name · Saved"), so the file name also appears twice (header and file bar).
- Root cause: `RichTextEdit` already owns a file bar (name, state, menu), and the app paints a second hand-made header beside it. The menu title and items (`"Save"`, `"^S"`, `"^Shift+S"`) are hardcoded rather than taken from `spec/defaults.yaml` `pane.hotkey_save_*`.
- Fix in the library: the file bar is the only place that shows document identity and state plus the File menu. Its title, items and shortcuts come from the spec. Hosts that want a header put app-level info in it (theme, mouse, alt screen), never "File" or the file name.

## 3. Implementation & Verification Plan
- Move the menu title and items to the spec (`rich_text_edit.file_menu`) and build the menu from it, sharing the shortcut formatter from 295.
- Remove `File` and the file name from the `loom edit` header. Keep app status there or drop the header row if it ends up empty.
- Coordinate with 291 (state display) and 294 (Open/Close items in the same menu).
- Validate with `loom eval`/`loom measure`; update the file bar golden tests and add a PTY capture of `loom edit`.
- `make test-q1`, `make install`.

/goal Make the RichTextEdit file bar the single spec-driven owner of the File menu and document name/state, remove the duplicate File label and name from loom edit's header, and verify with layout tests, or stop and report when blocked on a user decision or denied permission.
