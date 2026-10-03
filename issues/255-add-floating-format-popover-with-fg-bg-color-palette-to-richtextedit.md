# 255 — add floating format popover with fg/bg color palette to richtextedit

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature
**Related**: `docs/data/richtext-widget-v2.ansi`, `richtextedit.go`, `richtextedit_test.go`, `gallery/richtextedit.go`, `docs/Widgets.md`

---

## 1. Problem & Motivation

`loom.RichTextEdit` (added in issue #254) supports inline styling via hotkeys (`Ctrl+B`, `Ctrl+I`, `Ctrl+U`) and mouse selection. However, discovering formatting shortcuts and applying custom foreground/background colors or hyperlinks interactively requires a contextual floating popover / toolbar bubble attached to active selections, as mocked up in `docs/data/richtext-widget-v2.ansi`:

```text
   [ B │ I │ U │ #FG │ #BG │ Link ]
   ▲
Here is some selected text ready for styling.
```

## 2. Technical Specification

### A. Popover Geometry & Trigger Lifecycle
- `ShowPopover bool`: Configurable on `RichTextEdit` (default: true or opt-in).
- When a non-empty text selection exists (`HasSelection`), calculate the bounding anchor column and line above the selection start.
- Position the popover box either above the selection start or below if close to the top boundary. Render anchor pointer (`▲` or `▼`).
- Hide popover when selection collapses or `Esc` is pressed.

### B. Popover Actions & Interactive Sub-palettes
- **Format Toggles**: `B` (Bold), `I` (Italic), `U` (Underline), `S` (Strike).
- **Color Pickers**:
  - Clicking/activating `#FG` or `#BG` opens an inline mini-palette popup (16 ANSI colors / 6-hue palette row).
  - Selecting a color applies `Style.FG` or `Style.BG` to all spans within the active selection range.
- **Link Action**: Prompt / toggle URL link on selected span.

### C. Input Routing & Mouse Hits
- Popover consumes mouse clicks directly on its action buttons (`[ B ]`, `[ I ]`, `[ #FG ]`, `[ #BG ]`, etc.).
- Keyboard navigation: pressing `Alt+F` or `Ctrl+Space` focuses popover items with arrow keys / enter, while regular editing or typing replaces the selection and dismisses the popover.

## 3. Implementation & Verification Plan

- **M1 (Popover Geometry, Rendering & Actions)**:
  - Implement popover layout, anchor positioning, and button rendering above selection in `richtextedit.go`.
  - Handle mouse clicks on format action buttons (`B`, `I`, `U`, `S`).
  - Unit tests in `richtextedit_test.go`.
- **M2 (Inline FG/BG Color Picker & Palette Popdown)**:
  - Implement mini-palette popdown for `#FG` and `#BG` selection.
  - Apply chosen FG/BG color to active selection range.
  - Unit tests verifying color mutation across multi-span selections.
- **M3 (Gallery Integration, Golden Coverage & Docs)**:
  - Update `gallery/richtextedit.go` with popover demonstration.
  - Add golden mockup test asserting rendering matches `docs/data/richtext-widget-v2.ansi`.
  - Update `docs/Widgets.md`.
