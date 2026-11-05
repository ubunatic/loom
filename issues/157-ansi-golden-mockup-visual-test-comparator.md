# 157 — ANSI Golden Mockup Visual Test Comparator

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [loom-games](../loom-games)

---

## 1. Problem & Motivation

Developing TUIs matching design mockups (such as `.ansi` golden files) requires rigorous visual parity testing. Currently, applications must manually split lines and inspect row text substrings or reconstruct cell grids to compare against ANSI files.

A dedicated test helper in `loom` or a testing subpackage would streamline visual regression testing across all Loom-based apps.

## 2. Technical Specification / Findings

- Provide `loomtest` or `Canvas.DiffANSI` helper:
  ```go
  // AssertCanvasMatchesANSI compares a Canvas against golden ANSI content.
  // Outputs detailed cell-by-cell visual diffs on mismatch (coordinate, expected/actual glyph, styles).
  func AssertCanvasMatchesANSI(t *testing.T, canvas *loom.Canvas, goldenANSI []byte)
  ```
- Uses `loom.ParseANSI` to decode expected grid cells and compares with canvas cell buffer.

## 3. Implementation & Verification Plan

### Goal
Provide an automated visual parity comparator for golden ANSI files.

### Acceptance Criteria
- [ ] Compares glyphs, continuation cells, foreground colors, background colors, and bold/dim attributes.
- [ ] Produces human-readable visual diff reports on failure highlighting mismatch coordinates.
- [ ] Unit tests covering exact matches and intentional mismatch diagnostics.

## Status 2026-11-05 (Jules audit PR 100, partly done, 4/4 refs hold within 1 line)
The comparator exists as `loomtest.CompareWidget` (loomtest/compare.go:23), not `AssertCanvasMatchesANSI`. It
compares glyph, FG and BG and prints a row/col report (compare.go:71). Left: compare bold/dim and continuation
cells, and add `loomtest/compare_test.go` (today only `colorpicker_reference_test.go` uses it).
