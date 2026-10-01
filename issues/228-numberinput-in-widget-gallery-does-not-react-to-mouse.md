# 228 — NumberInput in widget gallery does not react to mouse

**Status**: Closed — implemented mouse scroll and click stepping for NumberInput with full test coverage
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

### Milestone 1 (Delivered: c8f5c0c5)
- Implemented `NumberInput.ConsumeMouse` for scroll wheel up/down stepping and click stepping.
- Added unit tests in `numberinput_test.go` and gallery integration test in `gallery/gallery_test.go`.

### Milestone 2: Pre-Work & Refinements
- Run `gofmt -w gallery/gallery_test.go` to fix table formatting.
- Refine stepper click hit-testing: clicks on the left arrow (`◂` at start) step down, clicks on the right arrow (`▸` at end) step up. If clicked on the interior number text, enter inline editing or ignore.
- Add unit tests for `lastRect` bounds rejection after `Draw` and for unformatted/default `NumberInput` layouts (`◂ 5 ▸`).
- Run `make test-q1` and commit with message ending in `(issue 228 M2)`.

/goal Implement mouse event handling for NumberInput, verify with unit and gallery tests, or stop and report when blocked on a user decision or denied permission.
