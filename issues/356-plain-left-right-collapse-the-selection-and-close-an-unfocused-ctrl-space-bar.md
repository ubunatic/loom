# 356 — Plain Left/Right collapse the selection and close an unfocused Ctrl+Space bar

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: UX
**Related**: 353 (focus model), 349 (Up/Down enter the bar), 344 (selection movements), 355

---

/goal With text selected and the Ctrl+Space bar shown but unfocused, plain Left/Right act as normal
editor keys (collapse the selection, the bar closes as a side effect) instead of only dismissing the
bar; tested with key tests and a PTY test; or stop and report when blocked on a user decision or
denied permission.

## Request (user, 2026-10-21, confirmed)

Today a plain Left/Right (no Shift, no Ctrl) on an unfocused bar only dismisses it, like Escape;
the selection stays. Standard editor behavior is wanted instead.

## Decisions

- Plain Left puts the cursor at the selection start, plain Right at its end (no extra step), and the
  selection is removed; the bar closes because the selection is gone.
- Shift/Ctrl+Shift selection keys and other selection movements (344) are unchanged.
- A focused bar (after Ctrl+Space or Up/Down) keeps Left/Right for moving between items.
- Up/Down on an unfocused bar still enter it (349).
- Without a bar, the same collapse rule applies to any selection (one editor rule, not a bar case).
