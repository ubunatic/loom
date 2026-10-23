# 371 — loom session: tabs (windows), ctrl-n opens a new tab

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 369 (`loom session`), 370 (prefix ctrl-b), `Tabs` widget

---

/goal `loom session` holds several child PTYs as tabs ("windows", tmux style): `ctrl-n` opens a new tab
running `$SHELL`, prefix keys switch and close tabs, a one-line tab bar shows them, screenshots capture
the active tab; keys live in the spec; PTY tests cover open, switch, close and exit of the last tab; or
stop and report when blocked on a user decision.

## Decisions so far (user, 2026-10-23)

- `ctrl-n` (direct, swallowed) opens a new tab; shell "next history" on ctrl-n is rarely used.
- Prefix is `ctrl-b` (370), so tmux habits apply.

## Proposed (host, to confirm before work)

- Each tab: own PTY, own `vt.VT`; only the active tab writes to the outer terminal; switching repaints
  from the VT (needs a full-screen redraw from `VT` cells, not raw byte replay).
- Keys: prefix `n`/`p` next/previous tab, prefix `1`–`9` go to tab, prefix `c` also new tab, prefix
  `x` close tab (with confirm), `ctrl-q` quits all.
- Tab bar: one row at top or bottom, using the loom `Tabs` widget look; the child PTY gets rows-1. Hidden
  while only one tab is open, so the single-tab case stays a pure pass-through.
- Closing the last tab, or its child exiting, ends the session with that child's exit code.

## Decided (host, 2026-10-23, user set /goal "implement tabs")

- Tab bar at the bottom, hidden while one tab is open; the active tab's PTY then gets all rows, and is
  resized (SIGWINCH) to rows-1 when the bar appears.
- New tab runs `$SHELL` (fallback `/bin/sh`).
- Inactive tabs keep running and update their VT only.
- Out of scope: splits (panes), detach/attach.
- Prefix `x` closes without confirm in this ticket (the child's own exit is the confirm); a confirm is a
  follow-up if needed.

## Milestones

- M1: VT full-screen repaint (`VT.Repaint()` or similar: clear, cells with SGR, cursor position, alt
  screen state) with unit tests.
- M2: session holds tabs; `ctrl-n`, prefix `c`/`n`/`p`/`1`-`9`/`x`, all keys in spec; tab bar; last-tab
  exit code; PTY tests for open, switch (screen shows the other tab's content), close, last exit.
