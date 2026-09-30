# 214 — MenuBar: second Alt-F/E/B press does not close the menu

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 204

---

## 1. Problem & Motivation
User feedback: pressing Alt-F (or Alt-E, Alt-B) opens the menu, but pressing the same shortcut again does not close it.

## 2. Technical Specification / Findings
The menu accelerator should toggle: same accelerator on the open menu closes it; another accelerator switches menus.

## 3. Implementation & Verification Plan
PTY test: Alt-F opens File, Alt-F again closes it; Alt-F then Alt-E switches to Edit.

/goal Menu accelerators toggle their menu, proven by a PTY test; or stop and report when blocked on a user decision or denied permission.
