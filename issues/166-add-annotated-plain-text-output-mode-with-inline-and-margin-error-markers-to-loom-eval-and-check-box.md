# 166 — Add annotated plain-text output mode with inline and margin error markers to loom eval and check-box

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature
**Related**: #048, #129, #165

---

## 1. Problem & Motivation

Commands like `loom eval`, `loom check-box`, and `loom measure` identify formatting discrepancies in ANSI art and TUI frames: ragged rows, mismatched box border widths, misaligned Unicode corners, and trailing whitespace.
Currently, they output summary text or list error messages pointing to line coordinates (e.g. `line 4: 78 columns (ragged)` or `box validation failed`).

When debugging complex ANSI layouts, users need visual localization of exactly where each defect occurs within the rendered character grid:
- Identifying which row is short or overhanging.
- Identifying which column has a misaligned box corner or unexpected emoji/glyph expansion.
- Viewing the plain text output with direct annotations, footnotes, and callout arrows pointing at the exact issue.

## 2. Technical Specification / Findings

- **Annotated Plain Text Output Mode**:
  - Add an `--annotate` / `-a` (or `--explain`, `--show-errors`) flag to verification commands: `loom eval`, `loom check-box`, and potentially `loom view` / `loom measure`.
  - Echo back the rendered plain-text grid (leveraging `loom.AnsiBuffer` or plain-text rendering from #165).
- **Error Annotation Schemes**:
  - **Margin Callouts**:
    - For row-level errors (ragged row lengths, trailing whitespace), append indicators at the margin:
      - e.g. `<-- 1` on the right side of the row, or `2 -->` on the left margin.
      - Footnote index below the text grid: `1: row width is 78 columns (expected 80)`.
  - **Inline Glyph / Column Markers**:
    - For column-specific issues (misaligned vertical borders, broken corners, emoji width mismatches):
      - Support inline markers such as `❌¹`, `⚠️¹`, or superscript numbers directly replacing/tagging the erroneous coordinate.
      - Or print an underline pointer line directly below the offending row (e.g., `    ^--- ¹: expected '│', found ' '`).
  - **Color vs. Plain Modes**:
    - In plain mode (or non-TTY pipes): use ASCII/Unicode superscript markers and callout arrows (`<-- 1`, `¹: <text>`).
    - In color mode: highlight the offending cell in red/yellow background and use warning/error emoji (`❌¹`, `⚠️`).
  - **Footnote / Diagnostic Legend**:
    - At the bottom of the output, print numbered diagnostic details:
      ```text
      1: line 4: ragged width (78 columns, expected 80)
      2: column 40: vertical border broken (found ' ' instead of '│')
      ```

## 3. Implementation & Verification Plan

- **/goal**: Implement annotated plain-text output with inline and margin error callouts for `loom eval` and `loom check-box`, verifying with unit tests and broken/ragged ANSI fixtures, or stop and report when blocked on user input or denied permission.
- Create test fixtures with deliberate defects (ragged lines, mismatched box borders, emoji drift).
- Add tests in `cmd/loom/main_test.go` verifying the rendered annotations, error callouts, and footnote explanations.
- Verify with `make test-q1` and `make install`.
