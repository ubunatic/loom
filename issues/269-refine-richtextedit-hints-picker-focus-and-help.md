# 269 — Refine RichTextEdit hints, picker focus, and help

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [268](268-add-save-and-save-as-to-ricktextedit-with-a-file-menu.md)

---

## 1. Problem & Motivation
The RichTextEdit file bar still spends too much space repeating standard shortcuts, and the row can be clipped at common terminal widths. The Save as picker also switches focus without reliably moving the text cursor, stays open after clicks outside it, and visually runs the filename field into the directory list. Add concise key guidance and an F1 help popup that lists all RichTextEdit bindings.

## 2. Technical Specification / Findings
Keep the Ctrl+Space text-formatting popover separate from document actions. Move shortcut detail out of the crowded file bar and into the existing lower hotkey information area; retain a visible `F1 Help` entry there.

## 3. Implementation & Verification Plan
/goal Refine RichTextEdit hints and Save as interactions so the file bar stays readable, picker focus and dismissal behave predictably, and F1 explains every RTE key binding; stop and report if a product decision or denied permission blocks completion.

Acceptance: remove redundant `Alt+F`, `Ctrl+S`, and `Ctrl+Shift+S` text from the file bar and make the relevant shortcuts discoverable in the lower hotkey information area and F1 help popup; Tab focus changes also place the cursor in the focused FilePicker input; clicking outside the Save as popup closes it; a divider separates directory search from filename entry; F1 opens a popup listing all RichTextEdit key bindings. Verify hint layout at narrow and normal widths and add regression tests for focus, outside-click dismissal, divider, and help contents.
