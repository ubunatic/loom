<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 021 — Fix :help command blocking on interactive /dev/tty in test environments

**Status**: Closed — decoupled showHelp from /dev/tty via headless detection and test hook
**Priority**: P1 (High)
**Severity**: High
**Category**: Bug
**Related**: [003](003-public-api-audit.md), [cmd.go](../cmd.go), [cmd_test.go](../cmd_test.go), [pane.go](../pane.go)
**Roadmap stage**: Infrastructure / Test reliability
**Depends on**: None

## Problem and findings

Running `go test ./...` or `make test` stalls in `cmd_test.go` with the prompt:
```text
  :help        show help for current view           :back        go one level back                    :home        go to uzu main menu                  press any key to close                          
```
This is caused by [`TestCmdHelpDoesNotQuitInTest`](../cmd_test.go#L112-L124):
```go
func TestCmdHelpDoesNotQuitInTest(t *testing.T) {
	// In a test environment there is no /dev/tty, so showHelp() silently fails
	// and returns without quitting. Nav stays NavNone.
	c := loom.NewChoice([]loom.Item{{Name: "x"}})
	typeCmd(c, "help")
	quit := c.HandleKey(loom.KeyEvent{Key: "enter"})
...
```

The assumption that `/dev/tty` is absent in test environments is false whenever tests run in an interactive terminal, PTY runner, or container with an allocated controlling terminal.

In [`cmd.go`](../cmd.go#L172-L181):
```go
func (cb *cmdBar) showHelp() Nav {
	hw := newHelpWidget(cb.allCmds())
	pane, err := New(hw.ContentHeight())
	if err != nil {
		return NavNone
	}
	_ = pane.Run(hw)
	pane.Close()
	return NavNone
}
```
`loom.New()` succeeds, configures the terminal into raw mode, and starts an interactive `pane.Run(hw)` event loop that blocks waiting for user keyboard input. This causes tests to hang or take 16+ seconds until an external keypress or timeout terminates the run.

## Acceptance criteria

- [x] Decouple `showHelp()` from hardcoded, direct `loom.New()` interactive pane instantiation, allowing headless or mock execution during tests.
- [x] Ensure unit tests for command bar navigation and help commands never open `/dev/tty`, manipulate terminal modes, or block on stdin/TTY reads.
- [x] `go test ./...` and `make test` pass reliably and swiftly (< 2s) in both TTY and headless environments without user interaction.
- [x] Maintain the user-facing `:help` functionality when invoked inside an interactive application session.

## Verification

Run `go test -v -run TestCmdHelp ./...` and `make test` in an interactive terminal session to verify that tests execute immediately without blocking or prompting `press any key to close`.

## Scope limits

Limited to command bar execution and pane decoupling in `cmd.go` and `cmd_test.go`. Does not modify core pane input logic in `pane.go`.

## Delivery evidence

1. Added `isHeadless()` and `helpRunner` hook with `sync.RWMutex` to `cmd.go`. In `testing.Testing()` or when `LOOM_HEADLESS` is set, `defaultHelpRunner` safely returns `nil` without opening `/dev/tty` or switching raw terminal modes.
2. Added `export_test.go` exposing `SetHelpRunner` for test inspection without polluting the public production API.
3. Added `TestCmdHelpInvokesRunner` in `cmd_test.go` verifying that `:help` invokes the runner with the correct `helpWidget` and height, verifying rendered command lines (`:help`, `:back`, `:home`, `:custom`).
4. `go test -race ./...` runs in ~1.1s with zero hangs or tty blocking.
