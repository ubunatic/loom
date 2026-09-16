# 056 — Embed real applications as PTY-hosted widgets (tmux/screen-style)

**Status**: Open
**Priority**: P3 (Low) — aspirational, not scheduled
**Severity**: Minor
**Category**: Feature
**Related**: `widget.go`, `pane.go`, `canvas.go`, `examples/tabs`, [055](055-add-a-tab-panel-widget-for-pane-hosting.md)

---

## 1. Problem & Motivation

Following up on 055 (`Tabs` widget): the natural next ask is hosting loom's
own example applications (or arbitrary external programs) as tab panes in a
single demo app, one tab per app. That is *not* a small extension of 055 —
each `examples/*` `Run(args)` today bundles its own `Pane` creation, arg/flag
parsing, and `pane.Run(widget)` call into one function (see
`examples/filebrowser/filebrowser/filebrowser.go:34-41`,
`examples/split/split/split.go:51-57`); there is no bare `Widget` to hand to
`Tabs`. Two example apps (`monitor`, `treemap`) also run periodic redraw
goroutines tied to their own pane.

Two possible directions surfaced in discussion on 2026-09-16:

1. **Split each example into `NewWidget(...) (loom.Widget, error)` +
   thin `Run` wrapper.** Mechanical but real refactor across 5 examples;
   keeps everything in-process (no exec), so periodic-redraw goroutines just
   need to target a shared pane instead of their own.
2. **Host arbitrary real applications (not just our own examples) as
   subprocesses, tmux/screen-style**: allocate a PTY per child process
   (e.g. via `creack/pty`), run a VT100/ANSI parser over its output into an
   in-memory screen grid, blit that grid into a `Canvas` each `Draw`, forward
   key/mouse events back to the PTY as raw bytes, and propagate resize via
   `TIOCSWINSZ`. This is the actually-aspirational half of this ticket: it
   generalizes far beyond loom's own examples to "any terminal program as a
   loom widget."

## 2. Known constraints / open questions

- **Dependency policy conflict**: `docs/Go.md` says avoid deps; PTY
  allocation needs real syscalls (`openpty`/`pgrp`/`winsize`) that are not
  reasonably hand-rolled, so direction 2 requires accepting an external PTY
  dependency at minimum. A VT100/ANSI parser is a second, larger dependency
  surface unless deliberately scoped to a minimal subset.
- **Architecture reversal for direction 2**: `cmd/loom-demo/main.go`'s own
  doc comment states it imports each example's library package directly
  "rather than duplicating any example logic or execing subprocesses." PTY
  hosting is inherently exec-based — a second, subprocess-based hosting mode
  alongside the current in-process one, not a drop-in replacement.
- **Full-screen guest semantics**: alt-screen switching, cursor
  visibility/shape, and mouse-protocol passthrough matter for real guests
  (`vim`, `htop`, `tmux` itself) and are meaningfully more work than capturing
  simple line-oriented output.
- **Build-vs-embed alternative worth scouting first**: rather than
  hand-rolling a VT100 parser, look for an existing Go-based terminal
  multiplexer/emulator library (tmux-like) that could be embedded as a
  library dependency directly, instead of building PTY + VT parsing from
  scratch. If one exists with an acceptable license and API, it likely
  dominates a from-scratch implementation for direction 2. No such library
  has been identified or evaluated yet — this is a research task, not a
  known solution.

## 3. Suggested first step (not scheduled)

A minimal PTY-widget canary: one hardcoded child process, no `Tabs`
integration, proving the render/input loop (PTY spawn -> ANSI parse -> Canvas
blit -> key/mouse forwarding -> resize) before committing to either
direction above or investing in a library survey.

## 4. Verification & Acceptance

Not defined yet — this ticket is a placeholder to capture the idea and the
constraints found during initial discussion, not a scoped, schedulable unit
of work. Acceptance criteria should be written once a direction (1, 2, or a
library-embedding approach) is chosen and someone commits to scheduling it.
