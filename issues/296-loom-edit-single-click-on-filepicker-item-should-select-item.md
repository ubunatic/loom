# 296 — loom edit: Single click on FilePicker item should select item

**Status**: Closed — implemented and verified in 0bff53b
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: [172](172-add-a-reusable-filepicker-widget-on-top-of-the-directory-model.md), [277](277-standardize-click-only-focus-and-cursor-invariants-across-compound-widgets.md), [282](282-add-f2-side-panel-with-file-browser-to-loom-edit.md), [290](290-keep-the-loom-edit-file-browser-usable-after-cancel-and-mouse-activation.md)

---

## 1. Problem & Motivation
In the `loom edit` file browser, a single click on an entry often does not move the selection to it. Single click should select; double click or Enter should open.

## 2. Technical Specification / Findings
- `filepicker.go:62-64` configures its `Choice` with `DoubleClickToActivate = true` and `MouseTextOnly = true`.
- `choice.go:588-597`: with `MouseTextOnly`, a press outside the text run of the row (to the right of a short file name) returns `Handled()` before `c.sel = fi` (`:602`). A click on the row is swallowed without selecting. This is the most likely cause.
- `filepicker.go:296-314` `ConsumeMouse` rejects `e.Y < 1` and translates by `listRect - lastRect`. Check that the row offset is right with a header present and with no header (browser mode in `loom edit`).
- `cmd/loom/edit.go:818-826` translates into browser-local coordinates and changes focus on press. Make sure no second offset is applied (0-based child-local rule, docs/Widgets.md).
- Fix in the library: in `Choice`, `MouseTextOnly` should limit hover highlighting only. A left press anywhere on an item row selects it; activation stays governed by `DoubleClickToActivate`/`SelectOnlyOnClick`. No FilePicker- or app-specific offset patches.

## 3. Implementation & Verification Plan
- PTY repro: click a short name past its text, and click the first and last visible rows.
- Fix the `Choice` press path (and the FilePicker offset if it turns out wrong). Document the semantics in docs/Widgets.md.
- Tests: `choice_test.go` (press beyond text selects, does not activate; double click activates); `filepicker_test.go` row-offset table; PTY click in `loom edit`.
- `make test-q1`, `make install`.

/goal Make a single left click anywhere on a Choice/FilePicker row select it (double click or Enter activates) by fixing the library press path and offsets, verified with unit and loom edit PTY tests, or stop and report when blocked on a user decision or denied permission.
