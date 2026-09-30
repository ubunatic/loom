# 184 — Scroll TextInput horizontally when the value is wider than the field

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: textinput.go, issues/177, issues/180

---

## 1. Problem & Motivation
When the `TextInput` value is wider than the field, text past the edge is clipped and the caret can leave the visible area (issue 177).

## 2. Technical Specification / Findings
Keep a horizontal offset so the caret stays visible; show a `…` or arrow marker on clipped sides; measure with display width (wide runes, see docs/Go.md). Works with the masked mode of 173.

## 3. Implementation & Verification Plan
/goal Make `TextInput` scroll horizontally with the caret, with tests for wide runes, Home/End, and deletion at the edges; stop and report if Choice's filter prompt needs a different behavior.
