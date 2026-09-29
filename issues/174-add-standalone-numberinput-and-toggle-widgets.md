# 174 — Add standalone NumberInput and Toggle widgets

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: settings.go (KindNumber, KindBool), issues/169-*.md

---

## 1. Problem & Motivation
Number stepping (`◂ n ▸`, min/max/step, typed entry with validation) and on/off toggles exist only as `Settings` row kinds (`KindNumber`, `KindBool`). Forms outside a `Settings` list (issue 169) cannot use them.

## 2. Technical Specification / Findings
Extract the logic from `settings.go` into standalone `NumberInput` and `Toggle` widgets, and let `Settings` rows reuse them so behavior stays in one place.

## 3. Implementation & Verification Plan
/goal Provide `NumberInput` and `Toggle` widgets shared by `Settings`, with existing Settings tests and goldens still green; stop and report if Settings behavior would change.
