# 419 — loom edit: --files flag and directory argument open the file pane

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: owner request via taia (Ctrl+O o runs `loom edit` from taia's folder)

---

## 1. Problem & Motivation
`loom edit` always starts with the file pane (side panel, `showFileBrowser` in `internal/editor/edit.go`)
closed, and no flag opens it. The owner wants it open when taia launches the editor.

## 2. Specification
- New flag `--files` / `-f` (bool): start with the side panel shown and focused, as `showFileBrowser` does.
- `loom edit <dir>`: when the argument is a directory, open no file, root the file picker at that directory,
  and show the pane (same as `--files`). `loom edit` with no argument keeps today's behaviour unless `--files`.
- Plumb through `editor.Run` (an option or config field); do not add it to the editor.yaml config unless trivial.
- `loom edit --help` lists the flag; update the man page and docs if they list edit flags.

## 3. Verification
Tests in internal/editor: with the option set, the first frame has a non-empty browser rect and focus on the
browser; a directory argument roots the picker there and opens no buffer; default stays closed. Release v0.7.2.
