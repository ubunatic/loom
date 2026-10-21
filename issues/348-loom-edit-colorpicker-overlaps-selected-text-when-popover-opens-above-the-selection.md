# 348 — loom edit: ColorPicker overlaps selected text when popover opens above the selection

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 347 (one-row ColorPicker, measured height), 346 (🎨 popover integration)

---

/goal When the Ctrl+Space popover opens above the selection, the opened ColorPicker (and any other
submenu) must not cover the selected text; add a regression test, run `make test-q1` and
`make install`, or stop and report when blocked on a user decision or denied permission.

## Problem (user, 2026-10-21)

The popover bar can open above the selection. The ColorPicker then still extends downward (2–3
rows incl. marker and preview) and draws over the selected text, so the effect of a pick is hidden.

## Notes

- Expected: submenus grow away from the selection (upward when the bar is above it), or the bar is
  placed so bar + submenu fit on the chosen side. Check how the Box submenu behaves too.
- Picker height comes from `ColorPicker.Measure` since 347 M2 (richtextedit_popover_draw.go).
- Verify live in tmux with a selection near the bottom of the view (bar above) and near the top.
