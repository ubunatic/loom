# 297 — loom edit: Esc exits even with unsaved changes, only ^Q and F10 should quit

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [290](290-keep-the-loom-edit-file-browser-usable-after-cancel-and-mouse-activation.md), [291](291-loom-edit-s-saves-file-but-saved-modified-indicator-is-not-updated.md)

---

## 1. Problem & Motivation
In `loom edit`, pressing `Esc` immediately exits the editor without warning or confirmation, discarding unsaved modifications if the user has typed text. In a text editor and throughout apps/demos, this recurring issue stems from framework-level defaults.

The library-level framework defaults must be standardized as follows:
- **`Esc`**: Closes active dialogs, popups, overlays, or menus, or cancels active selections/modes, but **never** closes the App unless explicitly configured to do so.
- **`^Q` and `F10`**: **Always** close the app, unless explicitly configured to NOT do so.
- **Close Interception**: Compound applications like `loom edit` must be able to intercept the close action (`^Q`, `F10`, window close) to present a confirmation dialog asking the user whether to Save, Discard, or Cancel before exiting.

Solve this problem comprehensively at the library and specification level (`spec/defaults.yaml` and `spec/widgets.yaml`) so all current and future Loom applications and demos inherit consistent exit behavior by default.

## 2. Technical Specification / Findings
- **Framework Defaults (`spec/defaults.yaml`)**:
  - Audit and adjust `fallback_quit_keys`. `esc` must not be a default application exit key for interactive apps.
  - Establish `^Q` (Ctrl-Q) and `F10` as the standard application exit keys across the framework.
- **Widget & Pane Event Routing**:
  - `Esc` must be routed to the topmost active modal/dialog/popover/overlay to dismiss it. If no modal is active, `Esc` cancels selection or is consumed as a no-op, rather than falling through to terminate the pane.
  - `Pane` / App runner provides a clean close interception hook (e.g. `BeforeQuit` / `CanQuit` callback or event veto) allowing applications with unsaved state (`loom edit`) to pause exit and prompt the user.
- **`loom edit` Lifecycle**:
  - `loom edit` intercepts `^Q` / `F10` (and any external close request). If buffer has unsaved changes, display the Save/Discard/Cancel dialog.
  - `Esc` inside `loom edit` dismisses open side panels (F2 file browser, F3 search), context menus, or selection, but does nothing when at the top-level edit view.

## 3. Implementation & Verification Plan
- Update `spec/defaults.yaml` and Go spec bindings for default quit keys (`ctrl-q`, `f10`) and remove `esc` from global app quit fallbacks.
- Implement the close-interception contract in the core `Pane` / event loop.
- Wire `loom edit` to consume `Esc` contextually and guard `^Q` / `F10` with the unsaved-changes dialog.
- Add unit and PTY regression tests:
  - Typing text in `loom edit` and pressing `Esc` keeps the app running and preserves edits.
  - Pressing `Esc` with an open dialog/panel closes only the dialog/panel.
  - Pressing `^Q` or `F10` with unsaved modifications presents the Save/Discard/Cancel prompt.
  - Discarding quits the app; Canceling returns to the editor; Saving writes the buffer and then quits.
- Verify with `make test-q1` and `make install`.

/goal Enforce framework-wide defaults where Esc closes dialogs/overlays but never exits the app, while ^Q and F10 quit with close-interception in loom edit for unsaved changes, and verify with automated tests, or stop and report when blocked on a user decision or denied permission.
