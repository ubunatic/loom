# 254 — implement richtextedit widget with inline styling, selection spans, and pill badges

**Status**: Open
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
  - Implement `RichDocument`, `RichSpan`, and parser/serializer in `richtext.go` and `richtext_test.go`.
- **M2 (RichTextEdit Widget & Rendering)**:
  - Implement `loom.RichTextEdit` widget with cursor, multi-span line drawing, and selection highlighting in `richtextedit.go`.
  - Add unit tests verifying bounds, wrapping, and clipping.
- **M3 (Input Handling & Formatting Commands)**:
  - Implement `ConsumeKey` and `ConsumeMouse` with editing commands (`Insert`, `Delete`, `ToggleBold`, etc.).
  - Implement atomic pill badge traversal.
- **M4 (Gallery Demo & Golden Visual Tests)**:
  - Add `RichTextEdit` interactive demo tab to `examples/gallery`.
  - Add golden mockup comparator test against `docs/data/richtext-widget-v1.ansi`.
  - Update `spec/widgets.yaml` and `docs/Widgets.md`.
