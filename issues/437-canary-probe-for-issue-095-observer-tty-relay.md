# 437 — canary probe for issue 095 observer tty relay

**Status**: Open
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
terminal; `scripts/canary095/probe.go` adds the `-probe` mode). Findings go to
stdout and `/tmp/canary095.log`.

All probes are green on Linux under tmux.

- **A1 — observer attach.** The observer writes to `/dev/tty` and this works even
  when stdout is redirected to a file, which is what `go test` does. So the observer
  render must never use stdout. A child run with `Setsid`+`Setctty` gets its own
  controlling terminal, distinct from the observer's (child `fd0 -> /dev/pts/37` while
  the observer pane is a different pts).
- **A1b — read path.** A fresh open of `/dev/tty` accepts writes but **never sees
  keystrokes**; keystrokes arrive on the process's own terminal fd (fd 0). Probe P2
  reported `fd 0 saw "x"` while the `/dev/tty` fd stayed silent. The relay must read
  from `os.Stdin` when it is a terminal and fall back to `/dev/tty` otherwise. This is
  what made the first canary run look hung.
- **A2 — raw mode and restore.** Raw mode on the observer is set through `x/term`;
  restore is signal-safe (`terminal.Restorer` + `CatchSignals` per
  `docs/TerminalSafety.md`) and must be idempotent. Verified: `stty -a` before the run
  and after the run is identical, and termios compared through
  `unix.IoctlGetTermios` matches. Compare termios structs, not `term.State` pointers.
- **A3 — relay direction.** Full duplex works: an observer keystroke reaches the child
  (`A3 child got key "x"`) and child output reaches the observer render. Two conditions:
  the child side must leave canonical mode (`ICANON` off), otherwise a keystroke is not
  delivered to the child until Enter; and `unix.Select` can return `EINTR` from Go's
  signal delivery, so the loop must retry on `EINTR` instead of ending.
- **A4 — chrome rows.** Chrome survives only when the observer renders from the cell
  model (`vt.VT.Repaint()`), not by passing bytes through: the chrome row stayed intact
  while the child frame rendered below it. Byte pass-through cannot keep chrome because
  the child owns cursor positioning from row 1.
- **A5 — Ctrl-C.** `term.MakeRaw` clears `ISIG`, so Ctrl-C arrived as byte `0x03`
  and the probe looked hung with no way to stop it — this is the bug the owner hit.
  Raw mode must keep `ISIG` (clear only `ICANON`/`ECHO`), plus a signal path that
  restores the terminal and ends the loop. Verified with `ISIG` kept: Ctrl-C interrupts
  the harness, `stty -a` before and after is identical, the signal path is the one used,
  and forwarding `0x03` afterwards still interrupts the test child
  (`FAIL: child: signal: interrupt`). Rule recorded in `docs/TerminalSafety.md`
  ("Raw mode must stay interruptible") with a matching checklist item.

## 3. Implementation & Verification Plan
Milestone 1 of issue 095 split into three independently buildable, unit-tested parts
in `internal/ptytest/view/` (no `ptytest.Session` dependency until milestone 2):

1. **M1.1 `frame.go`** — `Frame` = child `*vt.VT` + chrome lines; `Render()` emits the
   observer stream from the cell model. Test: styled child stream (SGR, cursor moves,
   box glyphs) keeps chrome, preserves cell styles, trims trailing blank rows, keeps
   alt-screen state. Sketch `docs/design/095-view-layout.ansi`, checked with
   `loom measure` and `loom check-box`.
2. **M1.2 `relay.go`** — event-driven loop over the child master fd and the observer
   input fd, forwards bytes both ways, retries on `EINTR`, intercepts `0x03`, drains VT
   replies (DSR), no sleeps. Test: pipe pair stands in for the PTY; both directions
   asserted; `SIGUSR1` mid-select proves the loop survives; EOF/idle paths;
   `RaceFactor` budget.
3. **M1.3 `attach.go`** — observer terminal: write side `/dev/tty`, read side `os.Stdin`
   when it is a terminal, save termios, enter raw, `terminal.Restorer` +
   `CatchSignals`, idempotent `Restore()`, close fds, skip when no controlling terminal
   is available. Test: PTY pair as stand-in observer terminal — raw flags set, restored
   on normal exit, restored on `SIGINT`, second restore is a no-op, skip path taken when
   `/dev/tty` is unavailable.

Live checks for the owner: run `make canary-095` in a real terminal, type a key (the
child should echo it back), compare `stty -a` before and after, press Ctrl-C (must stop
the app, restore the terminal, and interrupt the test child), and check `ls /dev/pts` for
leaked fds.
