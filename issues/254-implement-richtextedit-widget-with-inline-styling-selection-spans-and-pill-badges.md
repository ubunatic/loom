# 254 — implement richtextedit widget with inline styling, selection spans, and pill badges

**Status**: Closed — implemented RichTextEdit widget with inline styling, selection spans, atomic pill badges, and gallery demo
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature
**Related**: `docs/data/richtext-widget-v1.ansi`, `docs/data/richtext-widget-v2.ansi`, `docs/data/richtext-widget-v3.ansi`, `ansibuffer.go`, `ansieditor.go`, `spec/widgets.yaml`, `docs/Widgets.md`

---

## 1. Problem & Motivation

Loom provides `loom.TextInput` and `loom.TextArea` for plain text input, and `loom.AnsiEditor` / `loom.AnsiBuffer` for fixed-dimension 2D spatial cell grid editing (pixel/cell painting).

However, terminal applications, chat tools, note editors, and LLM prompt inputs frequently require a **flow-based rich text editor** (`loom.RichTextEdit`) that supports:
1. Inline text editing with styled spans (bold, italic, underline, strike, custom FG/BG colors, code spans, hyperlinks).
2. Text range selections with shortcut/popover formatting actions (`Ctrl+B`, `Ctrl+I`, `Ctrl+U`, `Ctrl+K`).
3. Embedded inline chips/pills (e.g. `@mentions`, `#issues`, status badges) as non-breakable atomic tokens.
4. Export and import to/from ANSI SGR text streams and Markdown.

Design mockups for this widget are captured and validated in:
- `docs/data/richtext-widget-v1.ansi` (Inline WYSIWYG & Styling)
- `docs/data/richtext-widget-v2.ansi` (Selection Range & Floating Popover)
- `docs/data/richtext-widget-v3.ansi` (Mentions, Pills & Metadata)

## 2. Technical Specification

### A. Data Model (`RichDocument` / `RichSpan`)
- `RichTextEdit` maintains lines composed of text spans (`Span{Text, Style, PillData}`).
- Spans support Loom `loom.Style` (foreground color, background color, bold, dim, italic, underline, strike).
- Support atomic "Pill / Badge" spans that move and delete as a single unit when navigating or deleting.

### B. Widget Contract & Interactions (`loom.Widget`)
- Implements `loom.Widget` (`Draw`, `ConsumeKey`, `ConsumeMouse`).
- Key navigation: cursor arrows, word jumps (`Ctrl+Left`/`Ctrl+Right`), home/end, backspace/delete across span boundaries.
- Selection: `Shift+Arrows` expands selection range.
- Formatting toggles: `Ctrl+B` (bold), `Ctrl+I` (italic), `Ctrl+U` (underline), `Ctrl+K` (link).
- Mouse interaction: click to position cursor, drag to select range.

### C. Serialization & Tooling
- `ToANSI() string` & `FromANSI(string)` parsing ANSI SGR codes into rich spans.
- `ToPlainText() string` for unformatted text extraction.
- Golden mockup tests verifying visual rendering against `docs/data/richtext-widget-v*.ansi`.

## 3. Implementation & Verification Plan

- **M1 (Core Model & ANSI Serialization)**:
  - Delivered in commit `1134365` (`feat(richtext): add core document and ANSI serialization`).
  - Added `RichDocument`, `RichLine`, `RichSpan`, and `RichPill` in `richtext.go`.
  - Added SGR styling attributes `Italic`, `Strike`, `Invert` to `loom.Style` and updated ANSI parsing in `parse_ansi.go`.
  - Unit tests verified roundtrip and ANSI parsing in `richtext_test.go`, `style_test.go`, and `parse_ansi_test.go`.
- **M2 (RichTextEdit Widget & Rendering)**:
  - Delivered in commit `1c706e2` (`feat(richtext): render document widget viewport`).
  - Added `loom.RichTextEdit` in `richtextedit.go` implementing `loom.Widget` (`Draw`, `ConsumeKey`, `ConsumeMouse`).
  - Implemented multi-span line drawing, grapheme cluster width calculation with `measure`, cursor mapping, viewport scrolling (`ScrollX`, `ScrollY`), and selection range highlight styling.
  - Added tests in `richtextedit_test.go` and catalog entry in `spec/widgets.yaml`.
- **M3 (Input Handling & Formatting Commands)**:
  - Delivered in commit `99ad346` (`feat(richtext): add editing and selection input`).
  - Implemented `ConsumeKey` on `RichTextEdit`: text insertion with active style, newline splitting, span joining, backspace & delete with atomic pill badge deletion.
  - Implemented navigation (arrows, home/end, `Ctrl+Left`/`Right` word hops, `Shift+Arrow` selection expansion).
  - Implemented formatting shortcuts (`Ctrl+B`, `Ctrl+I`, `Ctrl+U`) for selected ranges and active input style.
  - Implemented `ConsumeMouse` with click-to-position and drag-to-select range.
  - Verified with comprehensive unit test suite in `richtextedit_test.go`.
- **M4 (Gallery Demo & Golden Visual Tests)**:
  - **Pre-Work / Requirements**:
    - Add `RichTextEdit` interactive demo tab in `examples/gallery` demonstrating inline formatting, badges/mentions, and formatting shortcuts.
    - Add golden mockup comparator test in `richtextedit_golden_test.go` verifying visual frame rendering matches `docs/data/richtext-widget-v1.ansi`.
    - Update `docs/Widgets.md` with documentation for `loom.RichTextEdit` and `loom.RichDocument`.
    - Format code, run `make test-q1`, run `make install`.
    - When all tests pass, close the ticket with `harnez issues close 254 "implemented RichTextEdit widget with inline styling, selection spans, atomic pill badges, and gallery demo"`.
