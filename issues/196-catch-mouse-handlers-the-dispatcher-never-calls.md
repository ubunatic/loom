# 196 — Catch mouse handlers the dispatcher never calls

**Status**: Draft
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: 112, docs/Widgets.md §Event Handling

---

## 1. Problem & Motivation
`DispatchMouseEvent` accepts `ConsumeMouse(e) EventResult` and `ConsumeMouseEvent(e) EventResult`, but, unlike keys, has no legacy `(quit, consumed bool)` form. A widget with the tuple signature compiles, looks right, and silently receives no clicks. This cost two developer rounds on 112.

## 2. Technical Specification / Findings
Known instance: `media.Widget.ConsumeMouse` (tuple form), now bridged by `ConsumeMouseEvent`. Choose one:
- a test that reflects over exported widget types and fails on a tuple-form `ConsumeMouse` without an `EventResult` adapter, or
- dispatcher support for the tuple form, mirroring `KeyConsumer`.

## 3. Implementation & Verification Plan
Prefer the test (it keeps the dispatcher small). Verify it fails when the media adapter is removed.

/goal Add the guard and verify it catches the media case, or stop and report if the design choice needs the user.
