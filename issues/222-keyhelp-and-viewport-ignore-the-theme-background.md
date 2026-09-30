# 222 — KeyHelp and Viewport ignore the theme background

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 205

---

## 1. Problem & Motivation
User feedback (ThemeBackground): the KeyHelp pane and the Viewport do not use the theme background.

## 2. Technical Specification / Findings
Both must paint their empty and text cells with the theme background, like the other themed widgets.

## 3. Implementation & Verification Plan
Render test: with a non-default theme, every KeyHelp and Viewport cell has the theme background.

/goal KeyHelp and Viewport use the theme background, proven by tests; or stop and report when blocked on a user decision or denied permission.
