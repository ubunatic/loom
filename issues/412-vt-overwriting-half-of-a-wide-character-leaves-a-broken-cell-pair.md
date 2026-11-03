# 412 — vt: overwriting half of a wide character leaves a broken cell pair

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: [405](405-session-drops-outer-replies-to-child-queries-the-vt-does-not-answer-decrqm-kitty-u-xtversion.md)

---

## 1. Problem & Motivation
`VT.putCluster` (`vt/vt.go`, the `v.cells[v.y][v.x] = Cell{...}` write) does not clear the other half of a wide character it overwrites. In a 6-column VT (host probe, 2026-11-03):

- `中` then `ESC[1;2H x`: cell 0 still holds `中` (width 2) and cell 1 holds `x`, so the row renders `中x`, three columns in two cells.
- `中` then `ESC[1;1H x`: cell 0 holds `x` and cell 1 is still a continuation cell, so the row renders `x` and loses a column.

Programs inside `loom session` that redraw over CJK text or emoji (editors, `ls` in columns) can misalign the rest of the row.

Found by Jules ideas run `vt-parser` (GitHub #83, Jules ID 3059214202567251601); the host reproduced both cases.

## 2. Technical Specification / Findings
Like xterm: before writing at `x`, if the cell is a continuation cell, blank the base cell at `x-1`; if the cell being replaced is a wide base, blank its continuation at `x+1`. Apply the same at the far edge of a newly written wide character.

## 3. Implementation & Verification Plan
/goal Writing over either half of a wide character leaves no orphan base or continuation cell; table tests cover both halves, a wide-over-wide overwrite and the right edge. Stop and report when blocked on a user decision.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| implement | jules-implement-412 | 8254930771683443876 | GH #86 | adopted: fix in `putCluster`; host added a straddle case, a right-edge case and a cell-level orphan check (Screen hides left-half orphans); all 7 cases run, 5 fail without the fix; host make test-q1 green |
