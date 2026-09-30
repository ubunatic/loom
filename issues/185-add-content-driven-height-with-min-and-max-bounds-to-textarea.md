# 185 — Add content-driven height with min and max bounds to TextArea

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: textarea.go, issues/177, issues/180

---

## 1. Problem & Motivation
`TextArea` has a fixed height from its layout; inline UIs want it to grow with the content up to a limit, like the Bubbles text area (issue 177).

## 2. Technical Specification / Findings
Add `MinHeight`/`MaxHeight` so `Measure` reports the content line count clamped to the bounds; scrolling starts once `MaxHeight` is reached. Default behavior unchanged.

## 3. Implementation & Verification Plan
/goal Add content-driven height bounds to `TextArea` with Measure and scroll tests; stop and report if Frame layout cannot re-measure on edits without a pane-level change.
