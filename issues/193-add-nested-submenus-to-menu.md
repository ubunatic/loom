# 193 — Add nested submenus to Menu

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: issues/170, issues/178, issues/180

---

## 1. Problem & Motivation
170 delivers `Menu`/`MenuBar` with one level of items. ncurses menus and desktop-style apps also need nested submenus (issue 178); split out to keep 170 one sprint.

## 2. Technical Specification / Findings
`MenuItem.Submenu` opens a second dropdown beside the item (flip left at the right edge); → opens, ← and Esc close one level; mouse hover opens after the pointer enters the item; clicking outside closes all levels.

## 3. Implementation & Verification Plan
/goal Add nested submenus to `Menu` with navigation, placement, and dismissal tests; stop and report when placement or hover behavior needs a user decision.
