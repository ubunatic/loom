# 200 — Tests: gallery tab switching and per-widget input

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Test
**Related**: 195 gallery

---

## 1. Problem & Motivation
The gallery has no tests proving that Tab switches demos or that each demo widget reacts to input, so regressions like the dropped media clicks (112) go unnoticed.

## 2. Technical Specification / Findings
`gallery/gallery_test.go` currently checks rendering only. For every registered demo, send a representative key (and click where the widget is mouse-driven) and assert the render or state changes; demos that are display-only are listed explicitly.

## 3. Implementation & Verification Plan
The test fails when a demo stops reacting (verify by breaking one).

/goal Every gallery demo is covered by a tab-switch and input-reaction test; or stop and report when blocked on a user decision.
