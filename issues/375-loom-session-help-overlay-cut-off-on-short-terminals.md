# 375 — loom session: help overlay cut off on short terminals

**Status**: Closed — Help fits short terminals in columns with footer visible, PTY test at 12 rows (896787b). Host q1 green; live 80x12 help shows all bindings
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: 373 (bar and prefix `?` help)

---

/goal Prefix `?` help in `loom session` is fully reachable on any terminal height of at least 5 rows:
it fits, or it scrolls/pages with keys from the spec, and never loses its first lines; a PTY test covers a
short terminal; or stop and report when blocked on a user decision.

## Finding (host, 2026-10-23, 373 live check)

In tmux 100x12, prefix `?` showed the list starting at `M  Toggle mouse`; the title and the bindings
before it (direct keys, help) were scrolled off the top. The overlay writes more lines than the screen
has rows.

## Options

- Two columns when the width allows, otherwise page with any key / show "more" in the footer.
- The footer ("Press any key to return") must stay visible.
