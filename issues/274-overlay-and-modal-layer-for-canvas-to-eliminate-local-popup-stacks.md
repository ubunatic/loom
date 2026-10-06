# 274 — Overlay and modal layer for Canvas to eliminate local popup stacks

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Architecture
**Related**: [273](273-audit-richtextedit-workarounds-and-plan-loom-library-fixes.md)

---

## 1. Problem & Motivation
Widgets like `RichTextEdit` (`richtextedit.go:58-59`, `260-269`, `720-732`, `910-922`) currently maintain local modal state fields (`helpPopup`, `savePopup`, `savePicker`), manually calculate popup dimensions, manually route keys/mouse to popups ahead of themselves, and manage pointer teardown. This leads to duplicate modal boilerplate and stale modal bug patterns across compound widgets.

## 2. Technical Specification / Findings
- Implement a reusable top-level or `Canvas`/`Pane` overlay layer for floating popups and dialogs.
- The overlay layer should own Z-ordering, bounds positioning, Escape key dismissal, and input event capture before background widgets.
- Widgets or applications can push popups to the overlay layer without storing modal state or writing custom key/mouse interceptors in their `Draw`/`ConsumeKey`/`ConsumeMouse` loops.

## 3. Implementation & Verification Plan
/goal Provide an overlay/modal layer in Loom so widgets do not need local modal fields or custom event interceptors.
