# 191 — Add cell cursor and frozen header columns to Table

**Status**: Closed
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: table.go, issues/179, issues/180

---

## 1. Problem & Motivation
`Table` selects whole rows only and scrolls headers and key columns out of view. tview and Textual's DataTable offer a cell cursor and fixed rows/columns (issue 179).

## 2. Technical Specification / Findings
Add an optional cell-cursor mode (←/→ move between cells, `OnCellSelect(row, col)`) and `FrozenCols int` so the leading columns stay visible during horizontal scroll; header row stays fixed. Rich renderable cells are out of scope.

## 3. Implementation & Verification Plan
/goal Add cell cursor and frozen columns to `Table` with navigation and scroll tests, row mode unchanged; stop and report if Table's horizontal scroll model must be redesigned first.
