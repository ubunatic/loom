# 175 — Add a DatePicker widget

**Status**: Closed
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: choice.go, issues/169-*.md

---

## 1. Problem & Motivation
loom has no date input; consumers would have to parse free text from `TextInput`.

## 2. Technical Specification / Findings
A `DatePicker` with a month grid, arrow-key/mouse navigation, min/max bounds, and a typed ISO-date fallback. Open question: whether time-of-day is in scope (record, don't decide).

## 3. Implementation & Verification Plan
/goal Ship `DatePicker` with tests and docs, or stop and report when scope (time-of-day, locale week start) needs a user decision.
