# 297 — loom edit: Esc exits even with unsaved changes, only ^Q and F10 should quit

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [290](290-keep-the-loom-edit-file-browser-usable-after-cancel-and-mouse-activation.md), [291](291-loom-edit-s-saves-file-but-saved-modified-indicator-is-not-updated.md)

---

## 1. Problem & Motivation
In `loom edit`, pressing `Esc` immediately exits the editor without warning or confirmation, discarding unsaved modifications if the user has typed text. In a text editor, `Esc` should dismiss open overlays, modals, search bars, or side panels, or clear selections—it must not unconditionally exit the app.

Only `^Q` (Ctrl-Q) and `F10` should trigger app exit, routing through the unsaved-change guard dialog when modified.

Solve this problem on the library level: ensure `RichTextEdit` and compound editor container key routing strictly consume `Esc` for local cancellation/dismissal actions, leaving application termination exclusively to standard exit keys (`^Q` / `F10`). Bindings and exit behaviors should be specified in `spec/widgets.yaml` and `spec/defaults.yaml`.

## 2. Technical Specification / Findings
- Audit `RichTextEdit.HandleKey` and `loom edit`'s event loop.
- Currently, `Esc` may bubble up or be treated by the container/runner as an unhandled key triggering default quit behavior, or is explicitly bound to quit without checking dirty state.
- Ensure `Esc` is consumed by the active editor/view for contextual cancellation (e.g. closing search overlay, closing file picker, clearing selection) and does not quit the application.
- Standardize quit keys across editor components to `^Q` and `F10`, which must check document modification status and prompt to Save/Discard/Cancel when modified.
- Align key action mappings with `spec/widgets.yaml`.

## 3. Implementation & Verification Plan
- Update key routing so `Esc` is handled contextually and cannot exit the application.
- Verify `^Q` and `F10` invoke the quit flow with unsaved change prompts.
- Add regression and PTY tests:
  - Typing text and pressing `Esc` keeps the editor running with modified content intact.
  - Pressing `^Q` or `F10` with unsaved changes prompts before quitting.
- Verify with `make test-q1` and `make install`.

/goal Prevent `Esc` from exiting `loom edit` when changes are present, restrict application exit to `^Q` and `F10` with unsaved-change protection, and verify with automated tests, or stop and report when blocked on a user decision or denied permission.
