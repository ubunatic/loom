# 215 — NumberInput demo: range -100.00 to +100.00, right aligned

**Status**: Closed — fixed with tests, suite green
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Enhancement
**Related**: 207

---

## 1. Problem & Motivation
User feedback: the NumberInput demo should use a bigger number range and align the value on the right.

## 2. Technical Specification / Findings
Demo value range -100.00..+100.00 with two decimals, right aligned in its fixed-width field. If right alignment is not supported by the widget, add it as a widget option.

## 3. Implementation & Verification Plan
Test renders e.g. `-100.00` and `  5.00` right aligned in the field; PTY check of the gallery demo.

/goal The NumberInput demo shows -100.00..+100.00 right aligned, proven by tests; or stop and report when blocked on a user decision or denied permission.
