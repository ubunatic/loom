# Flaky TestAllDialogAndTableNavigationPTY under -race: mouse did not close dialog (GitHub #22)

Status: Closed — Fixed in 5f1d8be: gallery dialog/table PTY test uses fresh-output screen waits instead of sleep loops; 19/20 -> 20/20 under -race -count=20 on 2 CPUs (host rerun 20/20)
Priority: P2
Category: Bug

## Problem
Gallery `TestAllDialogAndTableNavigationPTY` failed with "mouse did not close dialog" on main
b7579e6 under `-race` (docs-only commit), and passed on rerun (run 39151189778). It also failed in
race run 39027257269 (8.64s). GitHub issue ubunatic/loom#22. Not covered by 337 (a1b4372).

## Fix
Wait for the dialog-open screen state before sending the mouse click and for the closed state
after it (ptytest/loomtest screen waits), timeouts scaled by `raceFactor`; no sleeps. Same pattern
as 337; fix together with 340.

## Acceptance
- `go test -race -count=20 -run TestAllDialogAndTableNavigationPTY` passes in a 2-CPU container.
- Close GitHub #22 with the commit hash.
