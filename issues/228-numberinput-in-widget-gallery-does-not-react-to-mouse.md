# 228 — NumberInput in widget gallery does not react to mouse

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: `numberinput.go`, `gallery/gallery.go`, 204

---

## 1. Problem & Motivation

In the `loom widgets` gallery (and standalone `NumberInput`), `NumberInput` does not respond to mouse interactions because `NumberInput.ConsumeMouse` currently returns `loom.Ignored()`. Users clicking or scrolling with the mouse over the NumberInput widget cannot increment, decrement, or adjust values.

## 2. Technical Specification / Findings

- `NumberInput.ConsumeMouse` in `numberinput.go` is a no-op returning `Ignored()`.
- Mouse actions such as `MouseScrollUp` / `MouseScrollDown` or clicks on steppers/buttons should adjust the value according to `Step` and bounds.
- In `gallery/gallery.go`, ensure the NumberInput widget demo is wired and formatted to receive mouse events smoothly.

## 3. Implementation & Verification Plan

- Implement `ConsumeMouse` in `numberinput.go` to handle mouse scroll wheel and click actions.
- Add unit tests in `numberinput_test.go` and gallery integration test in `gallery/gallery_test.go`.
- Verify with `make test-q1`.

/goal Implement mouse event handling for NumberInput, verify with unit and gallery tests, or stop and report when blocked on a user decision or denied permission.
