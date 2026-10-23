# 371 — loom session: tabs (windows), ctrl-n opens a new tab

**Status**: Closed — Tabs in loom session: ctrl-n/prefix c new $SHELL tab, prefix n/p/1-9 switch, prefix x close, bottom bar hidden at one tab, DECSTBM protects bar, VT.Repaint on switch, last-tab exit code (1124380, dddb36a, cce3597). Host q1 green 2x; live tmux: switch harnez<->zsh<->mc, clear keeps bar, x hides bar, ctrl-q exit 0, no stale bar. Stale tab text after quit moved to 372
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

## Approved plan (host, 2026-10-23)

d371 plan approved: spec `session.bindings.new_tab`, `prefix_bindings.{new_tab,next_tab,previous_tab,
close_tab,select_tab}`; `VT.Repaint()` with replay round-trip test; one reader goroutine per PTY, only
the active tab writes out, one stdin dispatcher; rows-1 with the bar; last-tab exit code. Pre-work:

- The child can clear or scroll the whole outer screen (`ESC[2J`, `ESC[3J`, scroll past the last row):
  while the bar shows, set the outer scroll region to rows 1..rows-1 (DECSTBM) and redraw the bar
  (cursor save/restore) after each active write chunk; reset the region when the bar hides and on exit.
- Switching to a tab whose child uses the alt screen and back to one that does not must leave the outer
  terminal in the right screen; test both directions.
- `select_tab` uses the key's index in the list; no digit parsing in Go.
- On exit, restore the outer terminal: scroll region, alt screen off if loom turned it on, cursor visible.

## M1–M2 delivered (host review, 2026-10-23)

1124380 (VT.Repaint), dddb36a (tabs). Host q1 green, live in tmux 100x20: ctrl-n opens zsh tab, bar
shows, prefix n/1/2 switch, `clear` keeps the bar, mc in tab 2 survives switching, prefix x closes and
hides the bar, ctrl-q exit 0.

### M3 pre-work (required)

- After quit, a stale tab bar line (" 1: harnez  2: zsh") remains on the outer main screen below the
  shell prompt. The bar was drawn on the main screen while a child used the alt screen (or not erased
  on hide/exit). Erase the bar row when it hides and on exit, on the screen it was drawn on; PTY test:
  after exit the outer VT holds no bar text.
