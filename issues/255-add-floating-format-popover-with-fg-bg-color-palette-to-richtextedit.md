# 255 — add floating format popover with fg/bg color palette to richtextedit

**Status**: Closed — implemented floating format popover with FG/BG color picker for RichTextEdit
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
  - Delivered in commit `84ef5fa` (`feat: add rich text format popover M1`).
  - Implemented popover geometry, anchor rendering (`▲`/`▼`), button rendering (`B`, `I`, `U`, `S`, `Link`, `#FG`, `#BG`), and click hit testing.
  - Added unit tests in `richtextedit_test.go`.
- **M2 (Inline FG/BG Color Picker & Palette Popdown)**:
  - Delivered in commit `359bad1` (`feat: add rich text color palettes M2`).
  - Implemented 16-color swatch mini-palette popdown triggered by `#FG` and `#BG` buttons.
  - Implemented color application across all selected spans while preserving other styles.
  - Added unit tests in `richtextedit_test.go`.
- **M3 (Gallery Integration, Golden Coverage & Docs)**:
  - **Pre-Work / Requirements**:
    - Update `gallery/richtextedit.go` to enable/showcase the floating format popover and FG/BG color picker on selections.
    - Add golden mockup comparator test in `richtextedit_golden_test.go` verifying visual rendering matches `docs/data/richtext-widget-v2.ansi`.
    - Update `docs/Widgets.md` to document popover functionality and FG/BG palette actions.
    - Run `make test-q1`, run `make install`.
    - Close issue 255 with `harnez issues close 255 "implemented floating format popover with FG/BG color picker for RichTextEdit"`.
