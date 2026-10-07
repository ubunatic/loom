# 291 — loom edit: ^S saves file but Saved/Modified indicator is not updated

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [288](288-compose-loom-edit-panels-with-standard-widgets-and-clipped-responsive-layout.md)

---

## 1. Problem & Motivation
In `loom edit`, pressing `^S` saves the file contents to disk, but the "Saved/Modified" status indicator on the top status pill or footer does not update to reflect the clean/saved state. This leaves the user unsure whether the save operation succeeded and causes the UI to persistently display a stale "Modified" indicator.

Fix the issue on the library level: ensure that document state tracking, dirty flag transitions, and status bar notification/repaint events are cleanly handled by the underlying `RichTextEdit` / status bar components rather than ad-hoc caller-side patches. Align state indicator styling and defaults with `spec/` YAML definitions.

## 2. Technical Specification / Findings
- Investigate `RichTextEdit` and `loom edit` save event dispatching (`handleSave` / `OnSave`).
- When a document is saved cleanly, the dirty/modified flag in `RichTextEdit` must reset to clean, triggering an event/repaint callback that updates the status pill / header.
- Ensure status indicator labels, colors, and dirty-state indicators leverage spec definitions in `spec/widgets.yaml` and `spec/defaults.yaml`.
- Fix the library, not the caller: ensure any compound widget or app hosting `RichTextEdit` receives accurate dirty/clean lifecycle notifications.

## 3. Implementation & Verification Plan
- Connect `RichTextEdit` save action to clear the dirty buffer state and notify parent/status indicators.
- Add unit and PTY regression tests asserting that modifying text sets dirty state to "Modified" and pressing `^S` immediately clears the state and updates the status indicator to "Saved".
- Verify with `make test-q1` and `make install`.

/goal Ensure `loom edit` and `RichTextEdit` immediately update the Saved/Modified state indicator upon saving via `^S` with clean library-level lifecycle handling and spec integration, or stop and report when blocked on a user decision or denied permission.
