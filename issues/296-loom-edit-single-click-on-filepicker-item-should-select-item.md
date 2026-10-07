# 296 — loom edit: Single click on FilePicker item should select item

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: [172](172-add-a-reusable-filepicker-widget-on-top-of-the-directory-model.md), [277](277-standardize-click-only-focus-and-cursor-invariants-across-compound-widgets.md), [282](282-add-f2-side-panel-with-file-browser-to-loom-edit.md), [290](290-keep-the-loom-edit-file-browser-usable-after-cancel-and-mouse-activation.md)

---

## 1. Problem & Motivation
In `loom edit`'s file browser pane, clicking once on a file or directory item in the `FilePicker` does not select/highlight the clicked item as expected. Mouse interaction with list/picker items should immediately update the selected item index and trigger selection callbacks on single click (with double-click or Enter activating/opening the file/folder).

Solve this at the library level: fix mouse hit-testing and item selection semantics directly within the reusable `FilePicker` and `List` widgets. Ensure widget interaction behaviors adhere to standard mouse routing invariants and spec definitions in `spec/widgets.yaml`.

## 2. Technical Specification / Findings
- Inspect mouse event handling in `FilePicker` (and underlying `List` / `Directory` views).
- When a `MouseEvent` (press/release) occurs within the item area, compute the item index from local child coordinates and update `SelectedIndex` / cursor position immediately.
- Ensure single-click selection fires the appropriate selection change callbacks and renders the item in selected/highlighted state.
- Ensure mouse coordinates respect 0-based child-local invariants and container-owned translation (no subtracting 1).

## 3. Implementation & Verification Plan
- Fix `FilePicker.HandleMouse` to select the clicked item on single-click.
- Add unit tests simulating single-click mouse events at various row offsets and asserting selected item index updates.
- Test in `loom edit` and gallery demos with `make test-q1`.

/goal Update the reusable FilePicker library widget to select items on a single mouse click and verify mouse selection in loom edit, or stop and report when blocked on a user decision or denied permission.
