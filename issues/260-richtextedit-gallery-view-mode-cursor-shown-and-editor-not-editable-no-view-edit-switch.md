# 260 — RichTextEdit gallery: view-mode cursor shown and editor not editable, no View/Edit switch

**Status**: Closed — view mode no longer places a caret; gallery editor focused and editable; F7/clickable View/Edit toggle (3d5224c, 2dd10bd). Host reran make test-q1 green and verified live in tmux: typing reaches editor, caret stays in editor, F7 switches to View and blocks typing
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: issue 258 (commits f7d666f, 3e3cdc6), `richtextedit.go`, `gallery/richtextedit.go`

---

/goal Make the RichTextEdit gallery demo editable again with a visible View/Edit switch, and stop drawing an edit cursor in view mode; verify with tests and a live `loom widgets --show RichTextEdit` run, or stop and report when blocked on a user decision.

## Observed (user, live gallery after 258)

1. In the view-mode line, a block cursor sits on the "V" of "View mode", as if it were an edit caret.
2. The editor can no longer be edited; keys seem to reach the read-only view, and there is no visible way to switch between View and Edit mode.

## Expected

- View mode draws no edit caret (selection highlight may stay, in line with the 258 decision that selection/copy work in view mode).
- The demo's main editor stays editable and keeps keyboard focus by default.
- A visible View/Edit toggle (button, key hint, or both) switches the demo editor between modes, and the hint bar shows it.

## Notes

- Likely cause: the gallery demo added the view editor as a second focusable/key-receiving widget, or focus/routing lands on it. Fix it in the library or widget if the caret or focus behavior is wrong there, not with a gallery-only workaround ("fix the library, not the caller").
- Event-routing work: per AGENTS.md, routing tickets start on `codex:sol:med`.
