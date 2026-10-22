# 361 — ColorPicker: Up/Down keep the column when switching rows

**Status**: Open — M2 0ebabb1 reopened: dim markers must sit on the current colour when nothing is applied
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: UX
**Related**: 359 (two-row picker), 360 (markers)

---

/goal In the two-row ColorPicker, Up/Down move the cursor to the same column in the other row, and
Enter no longer applies a colour the user only passed over; specced if configurable, key and PTY
tests, docs/F1 help; or stop and report when blocked on a user decision or denied permission.

## Request (user, 2026-10-22)

With the cursor on FG cell 4, Down puts the cursor on BG cell 4 (and Up from BG back to FG cell 4),
whether or not anything was applied. Moving the cursor is not a pick, so this is safe.

## Decisions (host)

- Today each row keeps its own cursor and Enter applies both rows' cursors. With a shared column that
  would apply a colour the user never chose (FG red via Space, Down, Enter gives red on red).
  New rule: each row has an applied value and the cursor is shared. Space applies the cursor's cell
  to the active row. Enter applies the cursor's cell to the active row and closes; the other row
  keeps what Space applied earlier (or stays unchanged).
- Markers: the active row's marker follows the cursor (selection colour); the inactive row's dim
  marker shows that row's applied value, if any.
- Preview shows the applied values plus the cursor's cell on the active row.
- Mouse click on a cell: unchanged (applies that cell to its row, moves the cursor there).
- Update docs/Upgrading.md (Enter semantics from 359 change).

## Addendum (user, 2026-10-22)

- Both rows show a dim marker at their applied ("locked-in") colour; the active row additionally shows
  the cursor marker in the selection colour. When cursor and applied value are on the same cell, the
  selection-colour marker wins. With nothing applied, a row shows no dim marker.
- Correction (user): dim markers are always visible on both rows. With nothing applied, a row's dim
  marker sits on its current colour (the selected text's existing FG/BG, or `∅` when it has none).
  Only the cursor marker overlays it on the same cell.
