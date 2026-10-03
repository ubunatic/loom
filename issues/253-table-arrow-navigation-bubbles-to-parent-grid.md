# 253 — Table arrow navigation bubbles to parent Grid

**Status**: In Progress
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: 209 (EventResult contract), 242 (Dialog arrow keys bubble to Grid)

---

/goal Table navigation arrows are consumed at library level without also moving parent Grid focus, with regression tests; or stop and report when blocked on a user decision.

## 1. Problem & Motivation
In the widget gallery's All tab, Up/Down change the Table's row selection and also move focus to another Grid cell. A cell-cursor Table has the same problem with Left/Right when they move its selected column. Fix the library's event-consumption contract, without per-widget gallery wrappers or caller workarounds.

## 2. Technical Specification / Findings
- `Table.ConsumeKey` (`table.go`) falls through to `Ignored()` after arrow navigation. Up/Down wrap at row ends; Left/Right move only with `CellCursor` enabled and stop at column edges. The All-tab table currently has `CellCursor` disabled (`gallery/gallery.go`); enabling it is separate issue #240.
- `Grid.ConsumeKey` (`grid.go`) already dispatches ordinary arrows to the focused child first and navigates only if the result is not consumed. Shift-arrows bypass the child for explicit Grid navigation.
- Consume arrows when they change Table selection. No-op arrows bubble to Grid: Left/Right in row mode, at cell-cursor column edges, or Up/Down with fewer than two rows. Up/Down wrap and are consumed when at least two rows are available.

## 3. Implementation & Verification Plan
- Implemented in the library and covered with Table and Grid regression tests, including Shift-arrow navigation and no-op boundaries. The `sol` reviewer approved the diff.
- The latest `make test-q1` run failed in the untouched Pane PTY test and tests in the concurrent repaint-probe package. Re-run the quota suite and check `loom widgets --show All` before closing. Event routing work starts on `codex:sol:med` per AGENTS.md.
