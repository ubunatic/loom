# 357 — ColorPicker: Enter applies the highlighted swatch

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: UX
**Related**: 346 (ColorPicker), 347 (FG/BG mode row), 350 (∅ reset cell)

---

/goal In the ColorPicker, Enter applies the highlighted swatch in one press (what Space then Enter
does today), specced in spec/defaults.yaml, with key and PTY tests and docs/F1 help updated; or stop
and report when blocked on a user decision or denied permission.

## Request (user, 2026-10-21)

On the color picker, Enter should apply the currently highlighted item, as Space + Enter does now.

## Notes

- Check live code first: what Space does today (select/preview only?) and what Enter does alone.
- Space keeps its current meaning unless it becomes redundant; the ∅ reset cell (350) and the FG/BG
  mode row (347) must behave the same way under Enter.

## Pre-work (host, from 358 review)

358 changed RichTextEdit semantics (`unselect_on_cancel` now defaults to true; any bar close clears
the selection; either vertical arrow closes a focused bar) but added no row to the breaking-changes
table in `docs/Upgrading.md`. Add that row in this ticket's commit.
