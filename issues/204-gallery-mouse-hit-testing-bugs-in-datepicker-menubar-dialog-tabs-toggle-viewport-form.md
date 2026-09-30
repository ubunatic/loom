# 204 — Gallery mouse hit-testing bugs in DatePicker, MenuBar, Dialog, Tabs, Toggle, Viewport, Form

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: 195, 197-202 gallery

---

## 1. Problem & Motivation
User feedback on `loom widgets`: DatePicker: row is right but x is off (click on Sunday selects Tuesday). MenuBar ("File  Edit  Help"): clicks do nothing except on "elp"; a click on the "e" of Help opens File. Dialog: button clicks ignored. Tabs demo: clicking "Overview"/"Details" does nothing. Toggle: click does nothing (space works). Viewport: mouse drag scroll does not work (wheel and keys do). Form: hover changes input focus; only a click should. Correct already: Choice, FilePicker, Paginator, Tree, Media.

## 2. Technical Specification / Findings
Events reaching widgets are 0-based and child-local (docs/Widgets.md §8); a mouse handler must return `EventResult` or it is never called. Suspect widgets that measure their own offsets from stale rects or in bytes instead of display width (MenuBar), and ones whose demos never forward mouse.

## 3. Implementation & Verification Plan
Proof: each item gets a regression test that fails before the fix, driven through the real `loom widgets --show <Name>` binary in a PTY (`gallery/gallery_test.go`, `internal/ptytest`), plus a `docs/progress/<Widget>.ansi` capture where the look changes. For each widget, click the target text located by rune column in the PTY screen.

/goal Every listed widget reacts correctly to clicks at the right position, proven by PTY tests; or stop and report when blocked on a user decision.
