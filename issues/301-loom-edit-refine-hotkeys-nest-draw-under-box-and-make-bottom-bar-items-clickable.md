# 301 — loom edit: Refine hotkeys, nest Draw under Box, and make bottom bar items clickable

**Status**: Closed — implemented and verified in c239dc9
**Priority**: P1 (High)
**Severity**: Feature
**Category**: UX
**Related**: [292](292-loom-edit-f1-help-screen-looks-cluttered.md), [293](293-loom-edit-redundant-top-file-label-and-bottom-file-menu.md), [294](294-loom-edit-add-o-to-open-file-pane-and-w-to-close-file.md), [295](295-loom-edit-add-s-as-save-as-and-use-as-modifier-hints.md), [300](300-loom-edit-replace-top-title-bar-with-compact-bottom-status-icons.md)

---

## 1. Problem & Motivation
Several hotkey labels, toolbar structures, and mouse interactions in `loom edit` and `RichTextEdit` require refinement to provide an intuitive, polished editing workflow:
1. **Save as Hotkey**: "Save as" should not have a dedicated global shortcut keycap; it remains exclusively a File-menu option.
2. **Search Keycap Label**: Display only `^F` in hint bars and tooltips for Search, while keeping `F3` supported as a secondary hidden binding.
3. **F1 Help Visibility**: Explicitly display `F1 Help` in the bottom hotkey/hint bar.
4. **Popover Toolbar Structure**: In `RichTextEdit`'s format popover, nest "Draw" under the "Box" button/submenu instead of showing "Draw" as a separate top-level popover item.
5. **Draw Keybinding**: Rebind box-drawing mode to `^D` (Ctrl-D) instead of `F5`.
6. **`^O` Toggle Behavior**: `^O` should both open AND close (toggle) the file browser side panel (keeping `F2` as a secondary hidden toggle).
7. **Clickable Bottom Bar Controls**: Every button, shortcut keycap, and indicator in the bottom bar must be mouse-clickable to invoke its action (e.g. clicking `^O Files`, `^F Search`, `^S Save`, `F1 Help`, `F10 Quit`, or clicking the mouse mode / theme / altscreen indicators).

Solve these requirements on the library level: update `HintBar`, `RichTextEdit` popovers, and `EditorStatusBar` to support mouse interaction and spec-driven keybindings in `spec/defaults.yaml`.

## 2. Technical Specification / Findings
- **Spec & Bindings (`spec/defaults.yaml`)**:
  - `editor.hotkey_search_key`: Change display keycap to `^F` (binding handles both `ctrl-f` and `f3`).
  - `editor.hotkey_files_key`: Display `^O` (binding handles `ctrl-o` and `f2`).
  - `editor.hotkey_box_binding`: Update from `f5` to `ctrl-d` (and keycap to `^D`).
  - Remove global hotkey for Save As from default hint bars (retained in File menu).
  - Add explicit `F1 Help` entry to the bottom hotkey bar.
- **RichTextEdit Toolbar (`richtextedit.go`)**:
  - Update `popover_labels` in `spec/defaults.yaml` and popover rendering: remove top-level "Draw"; make "Box" open a submenu containing box styles (Plain, Rounded) and "Draw" mode toggle.
- **File Panel Toggle**:
  - In `cmd/loom/edit.go`, update `^O` handling to toggle `showSidePanel` (open if closed, close if open and focused).
- **Clickable Bottom Bar Interaction**:
  - In `HintBar` and `EditorStatusBar`: implement `ConsumeMouse(MouseEvent)` hit testing to translate click coordinates to individual entry indices and fire corresponding action callbacks (`OnClick`).
  - Wire bottom bar clicks in `cmd/loom/edit.go` to trigger:
    - `^O Files` -> toggle file browser
    - `^F Search` -> toggle search overlay
    - `^S Save` -> save buffer
    - `F1 Help` -> toggle F1 help overlay
    - `F10 Quit` -> trigger close request
    - Status icons: clicking theme cycles theme, clicking mouse icon toggles mouse tracking, clicking altscreen toggles altscreen mode.

## 3. Implementation & Verification Plan
- Update `spec/defaults.yaml`, `defaults.go`, and schemas for updated hotkey labels, bindings (`^D`, `^F`, `^O`), and popover structure.
- Refactor `RichTextEdit` popover to nest "Draw" inside the "Box" submenu.
- Implement mouse hit-testing on `HintBar` and `EditorStatusBar`.
- Update `cmd/loom/edit.go` to handle `^O` toggle and route bottom bar mouse clicks.
- Add unit and PTY tests for:
  - Clicking each bottom bar item and verifying the triggered action.
  - `^O` toggling the file browser open and closed.
  - `^D` toggling box draw mode.
  - Popover Box submenu showing Draw option.
- Verify with `make test-q1` and `make install`.

/goal Refine loom edit keybindings (^F search, ^D draw, ^O toggle), nest Draw under Box in the popover toolbar, and make all bottom bar buttons and status icons mouse-clickable, or stop and report when blocked on a user decision or denied permission.
