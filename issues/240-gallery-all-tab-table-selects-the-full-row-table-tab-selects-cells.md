# 240 — Gallery All tab Table selects the full row, Table tab selects cells

**Status**: In Progress
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: 232 (more widgets in the All tab)

---

/goal The Table in the gallery's "All" tab behaves like the Table tab (cell selection), with a test;
or stop and report when blocked on a user decision.

## 1. Problem & Motivation
In the "All" tab the table highlights the whole row in the "selected" color, while the single
"Table" tab allows selecting cells. The same widget should demo the same way in both places.

## 2. Technical Specification / Findings
- "All" builds its own table in `newAllDemo()` (`gallery/gallery.go` ~450) without `CellCursor`;
  the "Table" demo factory (`gallery/gallery.go` ~173) builds a separate one.
- Prefer one shared table constructor for both demos so they cannot drift. Check the other "All"
  widgets for the same duplication.

## 3. Implementation & Verification Plan
- Test that the All-tab table has cell selection enabled (left/right move the cell cursor).
- `make test-q1`, `make install`, PTY check of `loom widgets --show All`.

---

## Delivered

- The All tab builds Table, Stopwatch, Timer and KeyHelp from the standalone demo constructors (`newTableDemo`, `demos[...]`), so the All Table has the cell cursor, frozen column and controls; renders refreshed.
- Commits: `e931e73` (dev242); host made the All PTY table check wait for the cell cursor (`f6afcaf`).
- Host verification: `make test-q1` 0 FAIL; `make install`; PTY test clicks a Table cell in All and moves the cell cursor with left/right.
