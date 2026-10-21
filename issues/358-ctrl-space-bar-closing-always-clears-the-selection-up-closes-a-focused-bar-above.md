# 358 — Ctrl+Space bar: closing always clears the selection; Up closes a focused bar above

**Status**: Closed — Done in 3d75d0e, 0c249e4, 05c59f4, edf4c53: every bar close clears the selection (unselect_on_cancel default true; Select all keeps its selection); either vertical arrow closes a focused bar in both positions; submenus open only with Enter/Space; 14 old tests aligned; host q1 green; live tmux: bar below Down Down X and bar above Up Up X insert X without replacing text
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: UX
**Related**: 349 (Up/Down enter the bar), 353 (focus model), 356 (Left/Right collapse), 357

---

/goal Whenever the Ctrl+Space bar (and its submenus, incl. the ColorPicker) closes, the selection is
cleared; and the arrow that entered a focused bar closes it in both bar positions (Down below the
text, Up above it); specced, key and PTY tests, docs/F1 help; or stop and report when blocked on a
user decision or denied permission.

## Request (user, 2026-10-21)

Up/Down entering the bar regardless of its position is right and stays. Today, with the bar below:
Down focuses it, Down again closes it but keeps the selection, a third Down clears it.
1. Any close of the bar clears the selection (an action happened or the user cancelled), saving a
   key press.
2. With the bar above, Up on the focused bar does nothing today. It should close the bar, clear the
   selection and return to normal editing.

## Decisions (host)

- "Any close" means: Escape, arrow away, applying an action (style toggle, hotkey, colour pick,
  Frame, ≡ items), Ctrl+Space toggle, and mouse click outside. Select all (≡) keeps its new
  selection, since selecting is its action.
- The cursor stays where it is (selection end for a forward selection); text is not changed.
- The spec flag `unselect_on_cancel` (spec/defaults.yaml:327, false today) is the likely switch;
  flip its default to true if it covers all cases above, otherwise extend it. Keep it configurable.
- Bar above: Up on the focused bar closes it (mirror of Down on a bar below). Revised after M2
  (0c249e4): either vertical arrow closes a focused bar in both positions; simpler and consistent.

## M1/M2 review (host, 2026-10-21)

3d75d0e (close paths, default flip) and 0c249e4 (vertical arrows close) landed; host q1 red in
~14 tests (304, 317, popover/hamburger/box/Ctrl+Space). They predate 358 and encode the old
"selection survives a close" rule, which 358 changes on purpose.

### M3 pre-work

Update each failing test to the 358 rule, one by one: keep its original intent (isolation, style
applied, focus, undo) and only change the selection expectation after a close. If a failure is not
about the kept selection (e.g. italic not applied, wrong cursor, mouse leak), it is a bug: fix code.
- Submenus (user, 2026-10-21): open only with Enter or Space, never with Up/Down. On the focused bar
  the away-arrow closes it on every item, submenu items included (today Down on Frame/≡ does
  nothing; that becomes "close"). Inside an open submenu Up/Down move between its items; Escape
  leaves it.
