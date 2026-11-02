# 407 — loom view: ctrl-f is find, never page down; drop ctrl-b paging; document key policy

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 368 (partially reverted, 2ff965c)

---

/goal Make ctrl-f mean find in loom view, remove ctrl-f/ctrl-b paging, and document the key policy; or stop and report when blocked on a user decision or denied permission.

## 1. Problem
Owner rule (2026-11-02): **ctrl-f is always find.** pgup/pgdn and home/end scroll the page; home/end only when no text input is focused.

368 (`2ff965c`) bound `view.page_down` to `ctrl-f` and `view.page_up` to `ctrl-b` (`spec/defaults.yaml:359-360`). Found in a manual test in foot:
- Outside a session, ctrl-f pages down in `loom view`, which breaks the rule.
- Inside `loom session`, ctrl-f opens the session's scrollback "Find:" (`spec/defaults.yaml:617`), so the view binding is unreachable anyway.
- `ctrl-b` is the session prefix (`spec/defaults.yaml:612`), so it clashes the same way.

## 2. Plan
1. Remove `ctrl-f` from `view.page_down` and `ctrl-b` from `view.page_up`. Keep space, `b`, pgup/pgdn, home/end and g/G.
2. Add a find action to `loom view` bound to `ctrl-f` in the spec. If find in view is too big, file it separately and leave ctrl-f unbound here.
3. Document the key policy (ctrl-f = find; pgup/pgdn scroll; home/end scroll unless a text input is focused) in `docs/Widgets.md` next to the Tab policy.
4. Update the 368 PTY test and add one showing that ctrl-f does not page.

Re-check the current bindings and recent commits before starting.
