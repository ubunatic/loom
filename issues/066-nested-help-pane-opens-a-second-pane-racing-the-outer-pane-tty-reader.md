# 066 — Nested help pane opens a second Pane racing the outer pane tty reader

**Status**: Closed — single-pane ownership guard plus inline help popup; manually verified :h/CR show-dismiss cycle in the filebrowser; split-layout root-overlay gap tracked in 069
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: `cmd.go`, `pane.go`, `popup.go`,
[053](053-pane-run-key-decoding-drops-multi-key-reads-and-splits-escape-sequences-across-tty-reads.md),
[055](055-add-a-tab-panel-widget-for-pane-hosting.md)

---

## 1. Problem & Motivation

Surfaced while reviewing hosting readiness, but it is an **existing bug
independent of that initiative**: `cmdBar.showHelp` (`cmd.go:196-205`) calls
`defaultHelpRunner` (`cmd.go:180-190`), which opens a *second* `Pane` via
`New(height)` while the outer pane is still running.

The inner `New` opens `/dev/tty`, calls `term.MakeRaw`, issues a DSR cursor
query and polls the tty for the reply (`pane.go:84-93,153-190`) — while the
outer pane's reader goroutine is polling the **same** fd (`pane.go:396-441`).
The DSR reply can be consumed by the outer reader (and decoded as garbage
input), or the outer reader can swallow the user's keystrokes meant for the
help pane. On `Close`, the inner pane also calls `drainInput()` (TCIFLUSH) and
`term.Restore` on a tty the outer pane still owns
(`pane.go:613-654`).

This is only masked today by `isHeadless()` (`cmd.go:191-194`), which short
circuits the whole thing under `testing.Testing()` or `LOOM_HEADLESS` — i.e.
**every** test. So the racy path is exactly the untested one. It is closely
related to 053 (input-path correctness) and it violates the single-owner
invariant `pane.go:67-70` documents.

## 2. Design — resolved decisions

Loom's model is one pane owning the tty. Two options, in preference order:

1. **Render help inside the existing pane** as a `Popup` (`popup.go` already
   exists and is the intended overlay primitive). `showHelp` swaps in / layers
   the help widget on the running pane's root and restores on dismiss. No second
   raw-mode session, no second reader, works under hosting, and it is testable
   headlessly with `loom.Render`. This is the resolved choice.
2. Fallback if 1 proves impractical (the `cmdBar` does not have a handle on the
   pane): make pane ownership explicit and serialized — a package-level guard
   that refuses (or suspends the outer reader before) a nested `New` while
   another pane is live, so the failure is an error rather than a race.

Either way, keep `helpRunner` as the injection point (tests override it) but
stop making `isHeadless()` the only thing preventing the bug.

## 3. Verification & Acceptance

- A test exercising the real (non-headless) help path — via a PTY harness, as
  `make watch-pty` does — shows the help view, dismisses it, and returns to the
  outer UI with input still working. This must run without `LOOM_HEADLESS`.
- No second `loom.New`/`term.MakeRaw` occurs while a pane is running (grep +
  test assertion, or the guard from option 2 returning an error).
- Help rendering is coverable by a headless `loom.Render` golden test.
- Terminal state is correctly restored after dismissing help (no lost echo, no
  stuck raw mode) — checked in the PTY test.
- `go test ./...`, `go test -race ./...` and `go vet ./...` pass.

## 4. Implementation Notes

`Pane.New` now claims the process-wide terminal owner before opening `/dev/tty`
and rejects a nested pane with an explicit error. `Pane.Close` releases the
claim exactly once, including after repeated close calls. The default help path
now renders an inline `Popup` from the active `Choice` or `Table`, so it does
not need a second pane at all. The injected help runner remains available for
headless and custom tests.

Manually verified against `go run ./examples/filebrowser`: `:h<CR>` swaps the
`filter>` prompt for command mode and opens the inline help popup, `<CR>`
dismisses it and returns control to the file list with input still working,
terminal state intact. The single-pane ownership guard makes the original
race structurally impossible regardless, since the default path no longer
calls `Pane.New` a second time at all. The split-layout limitation (help
confined to its child pane rather than the full root) is tracked separately
in [069](069-render-help-as-a-root-level-modal-overlay.md).
