# 292 — loom edit: F1 help screen looks cluttered

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: UX
**Related**: [269](269-refine-richtextedit-hints-picker-focus-and-help.md), [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md)

---

## 1. Problem & Motivation
F1 help in `loom edit` is a flat wall of 18 long lines. It is hard to scan, and it repeats shortcuts that are hand-written rather than taken from the bindings, so it drifts (e.g. `^Shift+S` vs. the planned `⌃⌥S` in 295).

## 2. Technical Specification / Findings
- `richtextedit_help.go:14` `richTextEditHelpLines()` hardcodes every line as `"Topic: key desc; key desc"`; `:96` only word-wraps them. There are no groups, no aligned key column and no link to the bindings.
- `keyhelp.go` `KeyHelp` already renders a `KeyMap` (`keymap.go`) as a one-line joined string. It has no sectioned layout.
- `loom edit` adds its own keys (F2 Files, F3/^F Search, F5 Box, ^P Screenshot from `spec/defaults.yaml` `editor.hotkey_*`) that never appear in the help.
- Fix in the library: extend `KeyHelp` (or add `KeyHelpSections`) to render grouped `KeyMap` entries as a two-column table: an aligned key column using the shared modifier formatter from 295, a description column, and section headers. Group names, order and descriptions live in `spec/defaults.yaml` (e.g. `rich_text_edit.help_sections`). `RichTextEdit` builds its help from that. Hosts append their own `KeyMap` section instead of replacing text.

## 3. Implementation & Verification Plan
- Add the spec schema and data for help sections; generate the Go accessors via the existing `SpeccedDefaults` path.
- Implement the sectioned renderer (rune and display-width aware, scrollable inside the existing help overlay).
- Replace `richTextEditHelpLines`; register the `loom edit` app keys as an extra section.
- Validate the rendered help with `loom eval`, `loom measure` and `loom check-box`. Update `richtextedit_help_test.go` and add a golden test.
- `make test-q1`, `make install`.

/goal Replace the hardcoded RichTextEdit help text with a sectioned, aligned KeyHelp renderer fed from spec-defined bindings (including host-added sections like loom edit's), validated with loom eval/measure/check-box and tests, or stop and report when blocked on a user decision or denied permission.
