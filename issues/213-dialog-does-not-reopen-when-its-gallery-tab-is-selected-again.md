# 213 — Dialog does not reopen when its gallery tab is selected again

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: 203

---

## 1. Problem & Motivation
User feedback: after closing the Dialog, selecting the "Dialog" tab again does not show it again (203 fix incomplete).

## 2. Technical Specification / Findings
Re-selecting the tab must reopen the demo dialog. Check how the gallery resets or re-activates a demo on tab selection.

## 3. Implementation & Verification Plan
PTY test: open Dialog tab, close dialog, switch tab away and back (and reselect the same tab), dialog visible again.

/goal The Dialog demo reopens on every tab (re)selection, proven by a PTY test; or stop and report when blocked on a user decision or denied permission.
