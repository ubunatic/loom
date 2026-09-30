# 173 — Add masked password mode to TextInput

**Status**: Closed
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: textinput.go

---

## 1. Problem & Motivation
loom has no way to enter secrets; `TextInput` echoes every character.

## 2. Technical Specification / Findings
Add a mask option to `TextInput` (e.g. `Mask rune`, `•` default from spec) that draws the mask per rune while keeping the real value, and disables copy of the value. Likely no new widget needed.

## 3. Implementation & Verification Plan
/goal Add masked mode to `TextInput` with tests (cursor, deletion, wide runes) and docs; stop and report if a separate widget turns out to be necessary.
