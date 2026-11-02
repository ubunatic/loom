# 275 — Decouple focus state updates from Popup and Widget Draw rendering

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Architecture
**Related**: [273](273-audit-richtextedit-workarounds-and-plan-loom-library-fixes.md)

---

## 1. Problem & Motivation
`Popup.Draw` (`popup.go:58-61`) calls `SetFocus(p.Open)` on `p.Inner` during drawing. Coupling state modifications and focus side-effects to `Draw` calls violates unidirectional rendering principles and can cause unexpected focus mutation during render passes.

## 2. Technical Specification / Findings
- Focus state should be updated explicitly during component lifecycle events (such as modal open/close or container focus routing), not inside `Draw`.
- Refactor `Popup` and focusable containers so that focus transitions are explicitly handled in focus setters or navigation event consumers.

## 3. Implementation & Verification Plan
/goal Ensure `Draw` methods in `Popup` and composite widgets do not mutate `SetFocus` state as a rendering side-effect.

## Jules audit (2026-11-01, GitHub main 5c3c603)

Verdict: still valid (Jules said "partly done"; nothing has been removed). 7/7 refs hold. `SetFocus` is still called from Draw in `popup.go:69-71` (ticket said 58-61), `tabs.go:303-305, 322-324, 336-338, 358-360`, `stack.go:84-86` and `grid.go:209-211`. Report: `reports/jules/275.md`.
