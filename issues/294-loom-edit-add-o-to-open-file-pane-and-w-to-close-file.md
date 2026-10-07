# 294 — loom edit: Add ^O to open file pane and ^W to close file

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Feature
**Category**: Keybindings
**Related**: [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [282](282-add-f2-side-panel-with-file-browser-to-loom-edit.md), [290](290-keep-the-loom-edit-file-browser-usable-after-cancel-and-mouse-activation.md)

---

## 1. Problem & Motivation
In `loom edit`, standard editor keybindings for file management are missing:
- `^O` (Ctrl-O) should toggle or open the file browser side pane.
- `^W` (Ctrl-W) should close the current file (prompting for save if modified, or closing the document/tab).

Solve this on the library level: define standard file action commands in the library's editor widget / compound frame key routing rather than ad-hoc switch statements in the CLI runner. Keymaps and action bindings must be declared and documented in `spec/widgets.yaml` and spec definitions as the single source of truth.

## 2. Technical Specification / Findings
- In `RichTextEdit` / compound editor controller, map `Ctrl-O` (`^O`) to the action that toggles or focuses the file browser panel.
- Map `Ctrl-W` (`^W`) to the action that requests document close, routing through unsaved change confirmation guards.
- Declare these bindings in `spec/widgets.yaml` under editor actions and update keycap hints accordingly.
- Ensure event routing adheres to the global/local key routing contract (`EventResult`) without breaking existing bindings (e.g. F2).

## 3. Implementation & Verification Plan
- Implement `^O` and `^W` handlers in the editor library widget / event dispatcher.
- Add unit and PTY tests verifying `^O` toggles the file browser and `^W` closes the document with unsaved change detection.
- Verify with `make test-q1` and `make install`.

/goal Add `^O` to open/toggle the file pane and `^W` to close the active file in `loom edit` and library editor components backed by spec keymap definitions, or stop and report when blocked on a user decision or denied permission.
