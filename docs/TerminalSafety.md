# Terminal Safety: Auto-Wrap, Cursor State, and x/term

<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

This document exists because a specific class of terminal-UI bug has hit
this codebase more than once, looks completely different each time it
happens (scrambled colored boxes, clobbered shell scrollback, garbled
watch-mode output), and is easy for a coding agent to reintroduce even
after "fixing" it once. Read this before writing anything that writes
ANSI-styled, multi-row content directly to a terminal outside `loom.Pane`.

## The trap: terminal auto-wrap corrupts the whole screen, not one line

Every terminal has a real, current column count. If a program prints a
row wider than that count, the terminal's own auto-wrap (DECAWM) wraps
the overflow onto the next physical screen line — this is normal,
correct terminal behavior, not a bug in the terminal. The bug is on the
program's side: for a **multi-row grid** where each row's absolute
screen position is load-bearing (a treemap, a table, anything drawn with
sequential `\n`-separated writes), one wrapped row eats the start of the
next row's line, and the program's own explicit newline then advances
*again* — so every row after the bad one lands one physical line lower
than intended. The result looks like the whole screen scrambled, not
like one line got cut off, and it gets worse with every subsequent row.

This is exactly the failure mode Loom hit building `examples/treemap`
(see [`docs/studies/2026-09-14-terminal-safety-hardening-and-treemap-theme-iteration.md`](studies/2026-09-14-terminal-safety-hardening-and-treemap-theme-iteration.md)):
a too-wide `--width` on an 80-column PTY produced visibly torn, offset
colored boxes with no error, no panic, no test failure — only visible in
a real terminal.

**Go cannot catch this at compile time.** It is pure runtime string
content interacting with a runtime terminal property (`x/term.GetSize`).
There is no substitute for actually doing the defenses below.

## The fix has three independent layers — use all three

None of these alone is sufficient; each catches what the others miss.

1. **Clamp dimensions to the live terminal size before rendering**
   (`examples/treemap/main.go`'s `resolveDimensions`/`clampDimensions`).
   Query the real terminal size (`x/term.GetSize`) on every redraw — not
   once at startup — since the terminal can be resized mid-`--watch`.
   This fixes the common case, but does nothing if the clamp itself has
   a bug or something bypasses it.
2. **Clip every row in software at write time** (`loom.ClipRow`,
   `rawscreen.go`) — ANSI-aware, preserves color/style in the kept
   prefix, appends a reset if it cut mid-style. This is the authoritative
   safety net: it doesn't matter what width a renderer (e.g.
   `graph.RenderTreemap`) was told to produce, because renderers never
   know the terminal's *actual* live width — only the final write loop
   does, so that is where this belongs, not inside a renderer.
3. **Disable the terminal's own auto-wrap** (`DECAWM`, `\x1b[?7l` /
   `\x1b[?7h`) for the duration of a redraw loop, as a second, terminal-
   level line of defense: even a row that somehow reached the terminal
   still too wide (a bug in layer 2, a race) can no longer wrap onto the
   next physical line and cascade — worst case it clips at the terminal's
   own right margin instead.

Layers 2 and 3 are bundled in `loom.RawScreen` (below); layer 1 is a
caller's own responsibility — see `graph.RenderTreemap`'s "CALLER
RESPONSIBILITY" doc comment for the contract renderers make to their
callers about this.

## `loom.RawScreen` vs. `loom.WriteRows` — these are NOT interchangeable

Package `loom` (root) exposes two terminal-output primitives in
`rawscreen.go`, for callers who need inline per-cell ANSI color/style
that `loom.View` would strip (`View` collapses every line to one
`Style`). **Picking the wrong one is itself a bug**, not just a missed
optimization:

- **`loom.RawScreen`** (`OpenRawScreen`/`.Draw`/`.Close`) is for a
  **persistent, full-screen redraw loop** — a `--watch`-style command
  redrawing in place on an interval, the same convention `watch(1)`/
  `htop` use. It always draws relative to the *physical top-left of the
  terminal* (row 1), taking over the full visible screen on every
  `Draw`. This is deliberately **not** anchored to wherever the cursor
  happened to be when it was opened.
- **`loom.WriteRows`** is for a **one-shot "print once and exit"**
  path. It writes plain, sequential, newline-terminated lines at
  wherever the cursor already is, like any ordinary command's output —
  never touching cursor position.

Using `RawScreen` for a one-shot print is a real, previously-shipped
regression: absolute row-1 positioning has no idea what the shell already
had on screen above it, so it silently overwrites prior scrollback (5
lines of shell history become 1). This was caught only by a PTY test that
populated scrollback *before* invoking the one-shot command — a test
against a blank terminal cannot see it, because there's nothing above row
1 to clobber. If you're adding a new command with both a one-shot and a
`--watch` mode, they need **different** primitives, not the same one
configured differently.

## `x/term` coverage: use it for everything it covers

`golang.org/x/term` is the correct, portable layer for: entering/
restoring raw mode (`MakeRaw`/`Restore`), size queries (`GetSize`), and
terminal detection (`IsTerminal`). `pane.go`'s `termSize` used to
hand-roll `syscall.Syscall(SYS_IOCTL, TIOCGWINSZ, ...)` directly instead
— pure duplication of what `x/term.GetSize` already does, less safely
(needed `unsafe.Pointer`) and less portably. Fixed; see
[`docs/studies/2026-09-14-x-term-coverage-gap-in-pane-termsize.md`](studies/2026-09-14-x-term-coverage-gap-in-pane-termsize.md).

`x/term` does **not** cover everything terminal-related — it has no
flush/drain primitive (`pane.go`'s `drainInput`/`TCFLSH` is legitimately
hand-rolled, `x/term` has nothing to replace it with) and no polling API
(`golang.org/x/sys/unix.Poll` is the correct, and only, layer for making
a blocking `Read` cancelable). Before hand-rolling a new raw syscall
against a tty fd, check whether `x/term` already covers it — but don't
assume every terminal-adjacent thing belongs there either.

## Checklist for any new raw-ANSI terminal writer

- [ ] Does it need to redraw the same region repeatedly (`--watch`), or
      print once and exit? Pick `loom.RawScreen` or `loom.WriteRows`
      accordingly — never both from the same code path, never the
      watch-style one for a one-shot print.
- [ ] Is every dimension re-queried from the live terminal (`x/term.GetSize`)
      on every redraw, not cached from startup?
- [ ] Does every renderer call clamp its `Width`/`Height` to that live
      size before rendering (not after)?
- [ ] Does the actual write path clip each row (`loom.ClipRow`) at write
      time, independent of what the renderer was told to produce?
- [ ] If you're testing this, does the test populate the terminal with
      prior content (a PTY, not a blank buffer) before checking that
      prior content survives? A blank-terminal test cannot catch either
      the auto-wrap-cascade bug or the one-shot-clobbers-scrollback bug.
