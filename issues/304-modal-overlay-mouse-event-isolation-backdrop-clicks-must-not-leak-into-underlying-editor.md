# 304 — Modal overlay mouse event isolation: backdrop clicks must not leak into underlying editor

**Status**: Closed — implemented in 896b1b0 and verified with make test-q1
**Priority**: P1 (High)
**Severity**: Bug
**Category**: Events
**Related**: [302](302-loom-edit-ux-refinements-theme-default-mouse-disable-backdrop-clicks-and-untitled-buffer.md), [303](303-temporary-mouse-grab-for-popups-menus-and-dialogs-in-mouse-off-mode.md)

---

## 1. Problem & Motivation
When running `loom edit --mousegrab` (or whenever mouse grab is active) and the user opens the format popover (`Ctrl-Space`), clicking outside the popover dismisses the menu but leaks the mouse event directly into the editor text. This unexpectedly relocates the text cursor or initiates text selection in the document.

While a modal overlay is open, mouse events must never leak into underlying widgets or secondary pipelines. Any click outside an active modal overlay must be consumed by the modal dismissal action alone and blocked from reaching underlying text editors.

This must be solved strictly at the library level (`RichTextEdit`, `Pane`, and overlay routing) without per-app or per-widget workarounds.

## 2. Technical Specification / Findings
- **Modal Event Isolation**:
  - In `RichTextEdit`: when `popoverOpen` (or any internal modal/menu) is active, `ConsumeMouse` must route mouse events to the popover/menu first.
  - If a click occurs outside the popover bounding box, `RichTextEdit` must close the popover and return `loom.Handled()` (consumed).
  - The mouse event must NOT fall through to the underlying document text mouse handler, cursor positioning, or selection logic.
- **Library-Level Architecture**:
  - Ensure clean, consistent modal event interception across `Pane`, `Popup`, `Dialog`, and `RichTextEdit`.
  - Remove any fragile or ad-hoc registration hacks added in previous commits in favor of a robust modal capture contract.
  - When a modal overlay is active, the overlay has exclusive mouse capture until dismissed.

## 3. Implementation & Verification Plan
- Refactor `RichTextEdit.ConsumeMouse` so modal popovers fully consume backdrop clicks on dismissal.
- Verify `Pane` and overlay event routing to ensure backdrop clicks never leak to underlying views.
- Add unit and PTY tests:
  - Open `Ctrl-Space` popover with mouse grab enabled.
  - Record cursor position and selection state.
  - Click outside the popover in the text area.
  - Verify popover is dismissed, and cursor position / document state remain completely unchanged.
- Verify with `make test-q1` and `make install`.

/goal Ensure modal overlay backdrop clicks dismiss the overlay without leaking mouse events to the underlying text editor or document, verified with make test-q1, or stop and report when blocked on a user decision or denied permission.
