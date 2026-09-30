# 201 — Theme switch in the widget gallery

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 195 gallery

---

## 1. Problem & Motivation
`loom widgets` renders one theme. Users should be able to cycle the available themes live to judge each widget.

## 2. Technical Specification / Findings
Use the existing theme registry and `ApplyTheme` (docs/Themes.md). Pick a key that demos do not use (e.g. F2) plus a `--theme` flag; show the theme name in the gallery chrome.

## 3. Implementation & Verification Plan
Test that switching changes the rendered colours of a demo.

/goal The gallery can cycle themes via key and flag, with a test; or stop and report when blocked on a user decision.
