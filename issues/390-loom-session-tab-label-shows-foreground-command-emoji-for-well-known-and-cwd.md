# 390 — loom session: tab label shows foreground command (emoji for well-known) and cwd

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 371 (session tabs), 373 (session bar), 383 (attach/detach)

---

/goal In `loom session`, each tab label shows the tab's current foreground command instead of the
static spawn name (`1: zsh`). Well-known commands show as an emoji plus the cwd, unknown ones as
the command name plus the cwd. Labels update when the foreground process or the directory changes,
the bar still fits narrow widths, and the emoji table lives in the spec. Stop and report when
blocked on a user decision (label format, platform scope); on a loop or no progress, clear the goal
and write a failure report saying why.

## 1. Problem & Motivation

The session bar renders `" %d: %s "` with `tab.title`, which is fixed at spawn to
`filepath.Base(cmdName)` (`internal/session/session.go:204`, `:242`, `:717`). Every shell tab reads
`1: zsh`, `2: zsh`, so tabs running `vim`, `htop` or `harnez` in different projects can't be told apart.

## 2. Technical Specification / Findings

- Foreground process: `tcgetpgrp` on the tab's PTY master gives the foreground process group. On
  Linux, read the name from `/proc/<pgid>/comm` (or `cmdline`) and the cwd from the
  `/proc/<pgid>/cwd` link. Unverified: macOS (no `/proc`; would need `proc_pidinfo`/`lsof`).
  Decide the scope: Linux only with a fallback to today's static title, or both platforms.
- Optional source: OSC 7 (cwd URI) and OSC 0/2 (title) sent by the child, if the VT parses them.
  Check whether `vt` handles these before adding polling.
- Label: `<emoji> <cwd>` for well-known commands, `<cmd> <cwd>` otherwise. Shorten the cwd (`~`
  for home, last path element or truncated at a spec width limit), using display width, not bytes
  (docs/Go.md). Open: does the tab number stay (`1: 🐚 ~/projects/loom`)? It keeps prefix `1`..`9`
  and alt-digit switching discoverable, so the default is to keep it.
- The emoji table (for example shells, editors, `git`, `go`, `make`, `htop`/`btop`, `ssh`, `man`,
  `less`) and the cwd width live in `spec/defaults.yaml` under `session`, with schema, and no Go
  copy (docs/Spec.md). Validate emoji width with `loom measure`.
- Refresh: poll at a modest spec interval or on output and focus events; never redraw when the
  label is unchanged. Tab width must be recomputed (`:204`).

## 3. Implementation & Verification Plan

- A foreground-info helper behind an interface with a fake for unit tests: label formatting, emoji
  lookup, cwd shortening, width cut-off and fallback when `/proc` is unavailable.
- PTY test in the full tier: start a shell tab, `cd` and run `sleep`, and assert that the label
  changes.
- Live check in tmux at 100 and 40 columns (`/tui-check`), then `make test-q1` and `make install`.
