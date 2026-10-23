# 372 — loom session: mouse tracking left on after exit, one safe terminal restore for all loom code

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: 369, 370, 371 (`loom session`), `pane_terminal.go` `Pane.close`, `pane_modes.go`

---

/goal After `loom session` exits (child exit, `ctrl-q`, prefix `q`/`x` on the last tab, SIGTERM/SIGHUP,
panic), mouse clicks in the outer shell print nothing; one shared, tested restore in a loom terminal
package resets every mode loom or a child may have set, and `Pane.close`, `loom session` and the other
raw-mode users call it; or stop and report when blocked on a user decision.

## Finding (user, 2026-10-23)

After exiting `loom session`, clicks in the shell print SGR mouse reports such as
`0;44;27M0;44;27m2;44;27M`: mouse tracking (`?1000h`/`?1002h`/`?1003h` + `?1006h`) is still on.

Likely cause: the child (e.g. `harnez usage`, `mc`) enabled mouse tracking and was killed by `ctrl-q`
before it could reset it, and `loom session` only restores the tty (`term.Restore`), the scroll region,
alt screen and cursor. The child's terminal modes pass straight through to the outer terminal, so loom
must reset them itself.

## Scope

- There is no loom `terminal` package yet; restore code lives in `pane_terminal.go` (`Pane.close`),
  `pane_modes.go` (mouse 1000/1002/1003/1006, bracketed paste 2004, kitty keyboard `>1u`/`<u`), and
  `cmd/loom/session.go`. Also check `cmd/loom-probe`, `cmd/loom-repaint-probe`,
  `examples/treemap`, `examples/ansiviewer` (they call `term.Restore` or set modes).
- Create (or extend) one package, e.g. `terminal`, with a safe restore that writes the full reset set
  (mouse modes 9/1000/1002/1003/1005/1006/1015 off, focus events 1004 off, bracketed paste off, kitty
  keyboard pop, alt screen off if entered, scroll region reset, SGR reset, cursor visible and default
  shape, auto-wrap on), drains input, restores the tty state, and is idempotent. Run it on every exit
  path, including signals and a deferred recover.
- Write the reset sequences in the spec if they count as loom content, per docs/Spec.md; protocol
  constants may stay in Go.

## Verification

- PTY test: a fixture child enables mouse SGR tracking, focus events and bracketed paste, then
  `loom session` is quit with `ctrl-q`; the outer VT/terminal shows all those modes off.
- Same for the child exiting by itself, the last tab closed with prefix `x`, and SIGTERM to loom.
- Host live check: `loom session -- mc`, `ctrl-q`, click in the shell: no output.

## Also (host, 2026-10-23, from 371 live check)

- With tabs, a non-active tab's main-screen text (zsh line "➜ live369 mc") stays on the outer main
  screen after `ctrl-q`, and the shell's next output overwrites it mid-line ("EXIT=0e369 mc"). The
  restore must leave the outer screen as the session found it (e.g. run tabs on the outer alt screen
  once more than one tab exists, or clear what loom drew) with the cursor at a clean line start.

## Also: direct tab select on alt digits (user, 2026-10-23)

- Add `session.bindings.select_tab: [alt-1 … alt-9]` (index in list = tab), swallowed like ctrl-n.
  Ctrl+digit is not usable (legacy encodings collide with NUL, Esc, ^\ … backspace). Check that
  `ESC 1` decodes as `alt-1`; PTY test: alt-2 switches to tab 2, alt-2 with one tab is swallowed.
