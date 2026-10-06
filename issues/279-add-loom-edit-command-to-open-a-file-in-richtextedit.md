# 279 — Add loom edit command to open a file in RichTextEdit

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: [270](270-pass-demo-arguments-to-loom-widgets-show-after.md), [271](271-richtextedit-save-ends-files-with-a-final-newline.md), [278](278-richtextedit-toolbar-redesign-key-caps-and-short-form.md), [125](125-build-ansiedit-standalone-ansi-file-editor-example.md)

---

## 1. Problem & Motivation
Editing a file today goes through the gallery: `loom widgets --show RichTextEdit -- <file>` (270). That works but carries gallery chrome (F8 BG, F9 Theme) and a long command. `loom view <file>` already opens a file for viewing; `loom edit <file>` would be its editing sibling.

## 2. Technical Specification / Findings
- New Cobra subcommand next to `viewCommand()` in `cmd/loom/main.go`: `loom edit <file>` opens a full-screen RichTextEdit with `ShowFileBar`, loads the file via `RichDocument.FromANSI`, binds `FilePath`, and saves with ^S (final newline per 271). A missing file starts empty, bound to that path.
- Bottom bar: the editor's `HotkeyBar()` (278) plus F10 Quit; no gallery theme/BG controls unless a `--theme` flag is wanted.
- Quit with unsaved changes: decide whether F10 asks to save or quits silently (gallery behaviour today). Record the choice before building.
- Share the load/bind code with the gallery demo (`gallery/richtextedit.go`) instead of copying it.
- Not a duplicate of `examples/ansiedit` (125): that edits ANSI art cell by cell; this edits rich text.

## 3. Implementation & Verification Plan
/goal `loom edit <file>` opens the file in RichTextEdit and ^S saves it back, sharing load code with the gallery demo; stop and report when blocked on a user decision (unsaved-quit behaviour) or denied permission.

Acceptance: CLI test for load, missing file and save round-trip; PTY test of the installed binary (open, type, ^S, F10, file content); `--help`, completion and man page list the command; `make install`.
