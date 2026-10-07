# 303 — Temporary mouse grab for popups, menus, and dialogs in mouse-off mode

**Status**: Closed — Implemented in 180cf4a; make test-q1 and make install passed
**Priority**: P1 (High)
**Severity**: Feature
**Category**: UX
**Related**: [300](300-loom-edit-replace-top-title-bar-with-compact-bottom-status-icons.md), [301](301-loom-edit-refine-hotkeys-nest-draw-under-box-and-make-bottom-bar-items-clickable.md), [302](302-loom-edit-ux-refinements-theme-default-mouse-disable-backdrop-clicks-and-untitled-buffer.md)

---

## 1. Problem & Motivation
In `loom edit` and general Loom applications, mouse tracking can be toggled between off (`"m"`) and on (`"M"`).
- In `"m"` mode, terminal mouse reporting is fully disabled so that users have unobstructed native terminal text selection (drag-to-copy).
- However, when an interactive overlay opens (such as the `Ctrl-Space` format popover, `F1 Help` screen, context menus, or `Save changes?` dialog), the user cannot interact with or click buttons in the popup with the mouse without first toggling `"M"` globally.
- Solution: Support **temporary mouse grab** during modal overlay lifecycles. When an overlay is active, Loom temporarily enables mouse tracking so the user can click options, scroll, or dismiss the popup. When the overlay closes, Loom automatically restores the user's base mouse mode (`"m"`).

## 2. Technical Specification / Findings
- **Overlay State Tracking**:
  - In `Pane` or compound container widgets (`editView`, `RichTextEdit`), track the active modal stack (`Popup`, `Dialog`, `Popover`, `Help`).
  - When an overlay opens while the base config is `MouseGrab: false` (`"m"`):
    - Emit mouse tracking enable escape sequences (`\x1b[?1000h\x1b[?1006h` or `\x1b[?1003h\x1b[?1006h`).
  - When all modal overlays are closed:
    - If `MouseGrab: false` (`"m"`), emit mouse tracking disable escape sequences (`\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l`).
  - If `MouseGrab: true` (`"M"`), mouse tracking remains permanently enabled regardless of overlay state.
- **Backdrop & Action Dispatch**:
  - While temporary mouse grab is engaged, clicks within the overlay trigger its actions/buttons; clicks outside the overlay dismiss the popup or cancel the dialog and immediately release mouse grab.

## 3. Implementation & Verification Plan
- Implement temporary mouse grab hooks in `Pane` and `cmd/loom/edit.go` on overlay open and close transitions.
- Add unit and PTY tests verifying:
  - Opening F1 help or popover in `"m"` mode emits mouse enable sequences and handles clicks.
  - Closing the popup or clicking the backdrop emits mouse disable sequences and restores native terminal mode.
  - Global `"M"` mode maintains persistent mouse capture across popup opens/closes.
- Verify with `make test-q1` and `make install`.

/goal Implement temporary mouse grab for popups, menus, and dialogs in mouse-off mode, and verify with make test-q1, or stop and report when blocked on a user decision or denied permission.
