# 349 — loom edit: Down should enter the Ctrl+Space popover also when it opens above the selection

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: UX
**Related**: 342 (Ctrl+Space toggle), 348 (popover above selection)

---

/goal Make Ctrl+Space then Down focus the popover bar wherever it opens (above or below the
selection), spec it in spec/defaults.yaml and the F1 help, test both placements, run `make test-q1`
and `make install`, or stop and report when blocked on a user decision or denied permission.

## Problem (user, 2026-10-21)

Users press Ctrl+Space, Down to enter the menu out of habit. Today only the arrow pointing towards
the bar focuses it (F1 help: "arrow towards toolbar focuses"; `arrows_cancel_unfocused_menu: true`),
so when the bar opens above the selection, Down cancels the menu instead.

## Notes

- Decide with the user whether Up should still focus a bar above (both arrows enter), or only Down.
  Proposed: both Up and Down enter the bar; Left/Right keep current behaviour.
- Event routing change: start on `codex:sol:med`.
