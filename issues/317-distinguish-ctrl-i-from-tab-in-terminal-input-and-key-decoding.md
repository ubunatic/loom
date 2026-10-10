# 317 — Directional popover focus (arrow towards popover) and decoupling Tab from popover focus

**Status**: Closed — implemented directional arrow popover focus and decoupled Tab to italic toggle on selection
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 316

---

## 1. Problem & Motivation
In terminal editors, `^I` (Ctrl+I) sends the ASCII byte `0x09`, which is identical to the `Tab` key.
Because `Tab` previously served as `menu_focus_key` for the format popover:
1. When text was selected and the popover appeared, pressing `Tab` (or `^I`) focused the popover rather than applying `ctrl-i` (italic styling) or advancing tabs/switching panels.
2. In legacy ASCII terminals without CSI-u keyboard enhancement, users wishing to format text with italics via `Ctrl+I` cannot do so cleanly if `Tab` is captured for popover focus.

## 2. Technical Specification & Design

### A. Directional Popover Focus (Arrow Towards Popover)
Instead of requiring `<tab>` to focus the popover menu:
- When the popover opens **above** the selection (`pointer == PointerDownGlyph` / `barY < anchorY`):
  - Pressing `up` (arrow towards the menu) **focuses** the menu (`popoverFocusSet = true`).
  - Pressing `down`, `left`, or `right` dismisses/cancels the menu and navigates text.
- When the popover opens **below** the selection (`pointer == PointerUpGlyph` / `barY > anchorY`):
  - Pressing `down` (arrow towards the menu) **focuses** the menu (`popoverFocusSet = true`).
  - Pressing `up`, `left`, or `right` dismisses/cancels the menu and navigates text.
- Clicking with the mouse on an item/submenu continues to focus as before.

### B. Decouple `<tab>` from Popover Focus
- `<tab>` is no longer the default popover focus trigger:
  - If text is selected in `RichTextEdit`, `key.Is("tab", "ctrl-i")` toggles italic text formatting (`s.Italic = on/off`).
  - If no text is selected (or when outside text editing), `<tab>` performs normal pane switching (between editor, file browser, and search panels) or indent insertion.
- The popover focus trigger can be specced via `spec/defaults.yaml` (e.g. `directional_focus: true` and `menu_focus_key: ""`).

### C. Popover Reopening
- `^<space>` (`popup_open_key`) or re-selecting text reopens the popover.

## 3. Implementation & Verification Plan
1. Update `RichTextEdit.consumePopoverKey` and `consumeKey` in `richtextedit.go`:
   - Detect popover placement relative to anchor (`isAbove := barY < anchorY`).
   - If unfocused:
     - If `isAbove` and key is `up` -> focus popover (`popoverFocusSet = true`).
     - If `!isAbove` and key is `down` -> focus popover (`popoverFocusSet = true`).
     - Other arrow keys cancel the popover when `arrows_cancel_unfocused_menu == true`.
   - When text is selected and popover is unfocused, route `tab`/`ctrl-i` to italic toggle `toggleAttribute(func(s *Style, on bool) { s.Italic = on }, ...)`.
2. Update `cmd/loom/edit.go` so `tab` routes to `RichTextEdit` when text is selected (applying italic toggle), and routes to panel focus switching when no text is selected.
3. Update unit and PTY tests in `richtextedit_test.go` and `cmd/loom/edit_*test.go`.
4. Verify with `make test-q1`.

/goal Implement directional arrow focus for textedit popover (arrow towards menu focuses) and route Tab/Ctrl+I to italic styling when text is selected, or stop and report if blocked on a user decision or denied permission.
