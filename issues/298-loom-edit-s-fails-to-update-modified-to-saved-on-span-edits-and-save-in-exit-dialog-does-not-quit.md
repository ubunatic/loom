# 298 — loom edit: ^S fails to update Modified to Saved on span edits, and Save in exit dialog does not quit

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [291](291-loom-edit-s-saves-file-but-saved-modified-indicator-is-not-updated.md), [297](297-loom-edit-esc-exits-even-with-unsaved-changes-only-q-and-f10-should-quit.md)

---

## 1. Problem & Motivation
In `loom edit`, after editing and deleting text, pressing `^S` saves the file to disk in the background, but the UI label fails to update from "Modified" to "Saved". Because the document is still considered modified by `IsModified()` / `DocState()`:
1. When quitting via `^Q` / `F10`, the user is prompted with the "Save changes?" dialog even though the file was just saved.
2. In the confirmation dialog, selecting "Save" performs the save, but because `IsModified()` continues to report `true`, the quit action is aborted and the user is thrown back into the editor in a stuck "Modified" state without exiting.

Solve this problem on the library level: ensure `RichTextEdit`'s document comparison and dirty-state tracking (`IsModified()`, `savedDocument`, normalization of empty spans / text deletions, and `DocState`) accurately reflect document equivalence after saves and edits. Align save state defaults and indicators with `spec/defaults.yaml`.

## 2. Technical Specification / Findings
- **Dirty State Tracking & Span Normalization**:
  - `RichTextEdit.IsModified()` uses `reflect.DeepEqual(e.savedDocument, e.Document.Lines)`.
  - When editing and deleting text (e.g. Backspace or Delete), empty `RichSpan{Text: ""}` spans, fragmented adjacent spans with identical styles, or differing slice capacities cause `reflect.DeepEqual` to return `false` even if the text content or serialized document is identical to the saved state, or even right after `SaveAs` if `cloneRichDocumentLines` captures an unnormalized structure.
  - Normalizing document lines/spans (pruning empty spans, merging adjacent spans with identical styles) or comparing normalized content ensures `IsModified()` is strictly accurate.
- **Save on Exit & Guard Action Lifecycle**:
  - In `cmd/loom/edit.go:326-331`, `Save` in the unsaved-changes dialog only executes `action()` (`v.doQuit`) if `!v.edit.IsModified()`.
  - When `IsModified()` fails to reset to `false` after `Save()`, the pending close/quit action is ignored and the application remains open.
  - `v.edit.Save()` must reliably clear modified state for existing files, and `guardUnsaved` must properly execute the intended quit action upon successful save.

## 3. Implementation & Verification Plan
- Normalize `RichDocument` spans on edits and saves so that `IsModified()` compares canonical span representations.
- Ensure `Save()` / `SaveAs()` guarantees `IsModified() == false` immediately after writing an existing file.
- Fix `cmd/loom/edit.go` unsaved-changes dialog so selecting "Save" always proceeds with quitting/closing once the file is saved successfully.
- Tests:
  - Unit test: Type text, delete text with backspace/delete, call `Save()`, assert `IsModified() == false` and `DocState() == DocStateSaved`.
  - Unit test: Trigger quit confirmation dialog on modified document, select "Save", assert document is saved and `pane.Quit()` is invoked.
  - PTY test: End-to-end edit -> delete -> `^S` -> verify `· Saved` status, then `^Q` -> verify clean exit.
- Run `make test-q1` and `make install`.

/goal Ensure ^S reliably clears the Modified state after typing and deleting text by normalizing RichDocument spans and ensuring the exit confirmation dialog cleanly quits upon Save, or stop and report when blocked on a user decision or denied permission.
