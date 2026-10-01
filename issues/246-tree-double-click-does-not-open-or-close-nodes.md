# 246 — Tree double click does not open or close nodes

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 243

---

/goal Double clicking a Tree node toggles it open or closed, with key and mouse tests, verified in the
gallery All tab, or stop and report when blocked on a user decision.

## 1. Problem & Motivation
In the gallery All tab, double clicking a Tree node does nothing (user report during the 243 review).
Users expect a double click to toggle a node, as in file browsers.

## 2. Technical Specification / Findings
- `tree.go` has no click or double-click handling (grep finds none).
- Follow the library's existing double-click model (check how other widgets, e.g. the filebrowser or List,
  detect it) rather than adding a Tree-only timer. A single click selects the node.
- Mouse events reaching widgets are 0-based and child-local (docs/Widgets.md).

## 3. Implementation & Verification Plan
- Tests: double click on a closed node opens it, on an open node closes it, on a leaf only selects; single
  click selects; the key path (Enter/Space or the existing toggle key) still works.
- Check in the gallery All tab inside a focused and an unfocused Grid cell.
