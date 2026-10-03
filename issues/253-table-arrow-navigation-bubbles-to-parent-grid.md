# 253 — Table arrow navigation bubbles to parent Grid

**Status**: Open
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
- Return a consumed result for arrows used by Table navigation. Determine whether no-op arrows (column edges, Left/Right in row mode, empty or single-row tables) should be consumed or bubble; do not assume a policy or confuse row wrapping with an unhandled boundary key. Stop and report if this requires a user decision.

## 3. Implementation & Verification Plan
- Add Table result and Grid-with-Table regression tests: Up/Down (including wrapping) and enabled cell-cursor Left/Right change selection without moving Grid focus. Cover the agreed no-op policy and preserve Shift-arrow Grid navigation.
- Keep the fix in the library; change routing/interfaces only if needed. Event routing work starts on `codex:sol:med` per AGENTS.md. Run `make test-q1`, `make install`, and check `loom widgets --show All` before closing.
