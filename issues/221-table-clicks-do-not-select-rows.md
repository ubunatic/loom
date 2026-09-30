# 221 — Table: clicks do not select rows

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: 204, 209

---

## 1. Problem & Motivation
User feedback: in the Table demo only the arrow keys select; mouse clicks do nothing.

## 2. Technical Specification / Findings
A click on a row must select it (child-local 0-based coordinates, header offset considered).

## 3. Implementation & Verification Plan
PTY test: click a data row, that row becomes selected.

/goal Clicking a Table row selects it, proven by a PTY test; or stop and report when blocked on a user decision or denied permission.
