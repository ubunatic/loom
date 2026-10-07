# 291 — loom edit: ^S saves file but Saved/Modified indicator is not updated

**Status**: Closed — implemented and verified in 29c6bf1
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [288](288-compose-loom-edit-panels-with-standard-widgets-and-clipped-responsive-layout.md)

---

## 1. Problem & Motivation
After `^S` in `loom edit` the file is written, but the user sees no reliable Saved state. The editor has two competing status surfaces and neither is driven by a library-level document-state contract.

## 2. Technical Specification / Findings
- `richtextedit_file.go:53` `SaveAs` already resets `savedDocument`, and `IsModified()` (`:14`) compares it by `reflect.DeepEqual`. The library state flips correctly.
- `richtextedit_filebar.go:133` `status()` duplicates that comparison instead of calling `IsModified()`. Two sources of truth.
- `cmd/loom/edit.go:383-395` draws its own header (`" File    <name> theme: ... mouse: ..."`) with no Saved/Modified field at all. `:418` status line shows only `statusMessage` (errors, screenshot path) and never a save confirmation.
- `^S` reaches `v.edit.Save()` only through the app's hint bar (`edit.go:214`). With the browser or search focused, the key goes to `filePicker`/`searchBar` first (`edit.go:576-625`), so save is focus-dependent.
- Fix in the library: `RichTextEdit` exposes one document state (`DocState()` → Untitled/Saved/Modified/Error) plus an `OnStateChange` callback. It owns `^S`/Save-as as editor actions so every host gets them, whatever has focus. The file bar and any host header read that state. State labels and colors come from `spec/defaults.yaml` `rich_text_edit` (labels) and the theme `ModifiedFG`/`SavedFG`.

## 3. Implementation & Verification Plan
- Reproduce first with a PTY probe: type, `^S`, capture the screen. Record which surface stays stale.
- Add `DocState()`/`OnStateChange`; make `status()` use it; move the state labels into the spec.
- Make `loom edit` show the library state in its header and drop its private status logic.
- Tests: unit (edit → Modified, Save → Saved, failed write → Error, callback fires once per transition); PTY (type, `^S`, header shows Saved, also with the browser focused).
- `make test-q1`, `make install`.

/goal Make RichTextEdit the single source of Saved/Modified state with a change callback and spec-defined labels, have loom edit display it so `^S` visibly flips to Saved, and verify with unit and PTY tests, or stop and report when blocked on a user decision or denied permission.
