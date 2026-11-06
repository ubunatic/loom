# 437 — canary probe for issue 095 observer tty relay

**Status**: Open — canary findings recorded; relay EINTR retry and child-side /dev/tty resolution still unverified
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Enhancement
**Related**: Issue 095

---

## 1. Problem & Motivation
Issue 095's human-observable PTY view mode depends on mechanisms that cannot be
unit-tested with a fake: attaching the observer's terminal from a test process,
raw mode plus signal-safe restore, a full-duplex relay between the child PTY and
the observer, and keeping chrome (scenario/prompt) rows while the child renders.
Per `docs/Canary.md`, probe each mechanism against the live environment before
building feature code on top.

## 2. Technical Specification / Findings
Canary lives in `scripts/canary095/main.go` (`make canary-095`, run inside a real
terminal; findings go to stdout and `/tmp/canary095.log`).

- **A1 — observer attach.** `/dev/tty` opens and enters raw mode from a Go process
  even when the process's stdout is not a TTY (`echo off=true, canonical off=true`).
  The test process's own stdout is captured by `go test`, so the observer must be
  `/dev/tty`, never stdout. A child run with `Setsid`+`Setctty` gets a controlling
  terminal distinct from the parent's (child `fd0 -> /dev/pts/37` while the parent's
  observer terminal is the tmux pane).
  *Not yet proven*: the child-side `/dev/tty` resolution — `os.Readlink("/dev/tty")`
  failed in the child, so the probe opens `/dev/tty` and resolves its own fd instead.
- **A2 — raw mode and restore.** Raw mode is set through `x/term`; restore must be
  signal-safe (`terminal.Restorer` + `CatchSignals` per `docs/TerminalSafety.md`), and
  the equality check compares termios via `unix.IoctlGetTermios`, not `term.State`
  pointers.
- **A3 — relay direction.** `unix.Select` returned `EINTR` and ended the relay loop on
  the first signal. The loop must retry on `EINTR`; the patch is in the canary but the
  observer→child key path has **not** been re-verified end to end.
- **A4 — chrome rows.** Chrome survives only when the observer renders from the cell
  model (`vt.VT.Repaint()`), not by passing bytes through: the pane showed the chrome
  row intact while the child frame rendered below it. Byte pass-through cannot keep
  chrome because the child owns cursor positioning from row 1.

## 3. Implementation & Verification Plan
Milestone 1 of issue 095 split into three independently buildable, unit-tested parts
in `internal/ptytest/view/` (no `ptytest.Session` dependency until milestone 2):

1. **M1.1 `frame.go`** — `Frame` = child `*vt.VT` + chrome lines; `Render()` emits the
   observer stream from the cell model. Test: styled child stream (SGR, cursor moves,
   box glyphs) keeps chrome, preserves cell styles, trims trailing blank rows, keeps
   alt-screen state. Sketch `docs/design/095-view-layout.ansi`, checked with
   `loom measure` and `loom check-box`.
2. **M1.2 `relay.go`** — event-driven select over the child master fd and observer fd,
   forwards bytes both ways, retries on `EINTR`, drains VT replies (DSR), no sleeps.
   Test: pipe pair stands in for the PTY; both directions asserted; `SIGUSR1` mid-select
   proves the loop survives; EOF/idle paths; `RaceFactor` budget.
3. **M1.3 `attach.go`** — open `/dev/tty` (or a supplied fd), save termios, enter raw,
   `terminal.Restorer` + `CatchSignals`, idempotent `Restore()`, close fds, skip when no
   controlling terminal is available. Test: PTY pair as stand-in observer terminal — raw
   flags set, restored on normal exit, restored on `SIGINT`, second restore is a no-op,
   skip path taken when `/dev/tty` is unavailable.

Resume point: rebuild `make canary-095` in a real terminal and re-verify A3 (key
injection) and A4 (child-side `/dev/tty`), then start M1.1. Live checks for the owner:
attach the tmux session, press keys and watch the child echo, `Ctrl-C` then compare
`stty -a` with the pre-run state, and check `ls /dev/pts` for leaked fds.
