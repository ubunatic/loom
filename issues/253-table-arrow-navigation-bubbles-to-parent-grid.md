# 253 — Table arrow navigation bubbles to parent Grid

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: 209 (EventResult contract), 242 (Dialog arrow keys bubble to Grid)

---

## 1. Problem & Motivation
In the widget gallery's All tab, arrow keys sent to the Table both change its row or cell selection and move focus to another Grid cell. The Table updates its internal state but returns `Ignored()`, so the parent Grid treats the same key as unhandled and navigates. This is a library-level event-consumption bug; callers embedding a Table should not need workarounds.

## 2. Technical Specification / Findings
`Table.ConsumeKey` handles Up/Down and, when `CellCursor` is enabled, Left/Right, but falls through to `Ignored()`. `Grid.ConsumeKey` dispatches ordinary arrows to its focused child and moves its own focus whenever the child result is not consumed. Determine the intended boundary behavior for arrows the Table cannot use (for example Left at the first column), so parent navigation remains predictable.

## 3. Implementation & Verification Plan
Fix the library-level event result/routing so a Table navigation key is not also applied by its parent Grid. Verify that arrows move Table selection without changing Grid focus, and that arrows not handled by the Table follow the agreed boundary behavior. Stop and report if that behavior requires a user decision.
