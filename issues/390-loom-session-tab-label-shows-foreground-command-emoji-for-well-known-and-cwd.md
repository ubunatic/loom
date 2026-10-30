# 390 — loom session: tab label shows foreground command (emoji for well-known) and cwd

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 371 (session tabs), 373 (session bar), 383 (attach/detach)

---

/goal In `loom session`, each tab label shows the tab's current foreground command instead of the
static spawn name (`1: zsh`). The label is a superscript tab number, then an emoji for a well-known
command or the command name otherwise, then the cwd basename (`¹🐚 loom`). Linux only. Labels update when the foreground process or the directory changes,
the bar still fits narrow widths, and the emoji table lives in the spec. Stop and report when
blocked on a user decision; on a loop or no progress, clear the goal
and write a failure report saying why.

## 1. Problem & Motivation

The session bar renders `" %d: %s "` with `tab.title`, which is fixed at spawn to
`filepath.Base(cmdName)` (`internal/session/session.go:204`, `:242`, `:717`). Every shell tab reads
`1: zsh`, `2: zsh`, so tabs running `vim`, `htop` or `harnez` in different projects can't be told apart.

## 2. Technical Specification / Findings

- Foreground process: `tcgetpgrp` on the tab's PTY master gives the foreground process group. On
  Linux, read the name from `/proc/<pgid>/comm` (or `cmdline`) and the cwd from the
  `/proc/<pgid>/cwd` link. Linux only (user decision 2026-10-30). Other platforms keep today's
  static title.
- Optional source: OSC 7 (cwd URI) and OSC 0/2 (title) sent by the child, if the VT parses them.
  Check whether `vt` handles these before adding polling.
- Label (user decision 2026-10-30): a superscript tab number, then the emoji, then the cwd
  basename, for example `¹🐚 loom`. Unknown commands show the name instead of the emoji
  (`²make loom`). The home directory shows as `~`. Tabs 10 and above use superscript digits too
  (`¹⁰`). Measure with display width, not bytes (docs/Go.md), and cut overlong basenames at a spec
  width limit.
- The emoji table (for example shells, editors, `git`, `go`, `make`, `htop`/`btop`, `ssh`, `man`,
  `less`) and the basename width limit live in `spec/defaults.yaml` under `session`, with schema, and no Go
  copy (docs/Spec.md). Validate emoji width with `loom measure`.
- Refresh: poll at a modest spec interval or on output and focus events; never redraw when the
  label is unchanged. Tab width must be recomputed (`:204`).

## 3. Implementation & Verification Plan

- A foreground-info helper behind an interface with a fake for unit tests: label formatting, emoji
  lookup, cwd shortening, width cut-off and fallback when `/proc` is unavailable.
- PTY test in the full tier: start a shell tab, `cd` and run `sleep`, and assert that the label
  changes.
- Live check in tmux at 100 and 40 columns (`/tui-check`), then `make test-q1` and `make install`.
