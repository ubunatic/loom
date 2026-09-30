# 212 — Choice: clicking an item exits the gallery

**Status**: Closed — fixed with tests, suite green
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: 203, 209

---

## 1. Problem & Motivation
User feedback on `loom widgets`: clicking a Choice item quits the gallery.

## 2. Technical Specification / Findings
Only q, F10 and Esc may exit the gallery, and Esc only when no child consumed it (bubbling contract from 209). A mouse click must never exit. Find why the click result reaches the quit path; fix in the library/gallery routing, not per widget.

## 3. Implementation & Verification Plan
PTY test: click a Choice item, gallery still running and item selected. Tests that q/F10 exit, and Esc exits only when unconsumed (e.g. an open popup closes first).

/goal Clicks never exit the gallery and only q/F10/unconsumed Esc do, proven by PTY tests; or stop and report when blocked on a user decision or denied permission.
