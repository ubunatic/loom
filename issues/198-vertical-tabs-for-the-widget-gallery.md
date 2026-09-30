# 198 — Vertical tabs for the widget gallery

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 195 gallery

---

## 1. Problem & Motivation
`loom widgets` shows ~40 demos as horizontal tabs, which overflow the width. A vertical tab list on the left scales and reads better.

## 2. Technical Specification / Findings
`loom.Tabs` has no vertical placement yet. Add it to the library (non-breaking, e.g. an option or `NewVerticalTabs`), then use it in the gallery.

## 3. Implementation & Verification Plan
TDD in `tabs_test.go` for layout and selection; update gallery demo and `docs/progress/Tabs.ansi`.

/goal loom.Tabs supports a vertical tab list and the gallery uses it; or stop and report when blocked on a user decision.
