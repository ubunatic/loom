# 277 — Standardize click-only focus and cursor invariants across compound widgets

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Architecture
**Related**: [273](273-audit-richtextedit-workarounds-and-plan-loom-library-fixes.md)

---

## 1. Problem & Motivation
`FilePicker` (`filepicker.go:295-303`) previously shifted input field focus (`p.nameFocus`) on every `MouseEvent` including `MouseHover`, stealing keyboard focus and hiding the text input caret whenever the pointer hovered over the file list. Similar issues occurred in `Choice` (`choice.go:627`) where hover changed selection state.

## 2. Technical Specification / Findings
- Enforce core library invariant: keyboard focus, field activation, and hardware text cursor placement must change ONLY on click (`MousePress` with `MouseLeft`) or explicit keyboard navigation keys, NEVER on passive `MouseHover`.
- Audit all compound input/navigation widgets (`FilePicker`, `Choice`, `Form`, `Table`, `DatePicker`) to ensure `MouseHover` does not alter focus or cursor coordinates.

## 3. Implementation & Verification Plan
/goal Enforce click-only/key-only focus and cursor placement invariants across all Loom compound widgets.
