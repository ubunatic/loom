# 113 — Fix filebrowser arrow up/down jumping multiple items

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [105](105-unify-filebrowser-navigation-pane-across-examples.md) (shared `NavigationPane`),
[063](063-convert-filebrowser-example-to-a-hostable-widget.md) (hostable widget, c0612e0)

## Goal

One arrow up/down press in `examples/filebrowser` moves the selection by exactly one item,
standalone and hosted.

## Observed

User smoke test after 105/063 (2026-09-24): up/down jumps several items per press. The rest of the
105 checklist (ESC back, parent reselect, `/` filter, click select) looks fine.

## Notes

- Suspect: one key reaching the list more than once, e.g. via frame focus dispatch and
  `browser.HandleKey`/`ConsumeKey`/`NavigationPane.HandleKey` together. Verify, don't assume.
- Also check `examples/ansiviewer` (same pane since 105 M3).
- Reproduction test first: send one "down" through the real key path (framed/hosted and
  standalone, ideally a PTY test) and assert the selection moves by one. It must fail before the fix.
