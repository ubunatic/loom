# Flaky Test304EditPTYPopoverBackdropIsolation under -race (GitHub #21)

Status: Open
Priority: P2
Category: Bug

## Problem
The cloud agent's flake hunt (`-race -count=5`, first dispatch) reported
`Test304EditPTYPopoverBackdropIsolation` intermittent: GitHub issue ubunatic/loom#21.
q-loom's PTY stabilisation (a1b4372, ticket 337) changed edit_301/302/303 tests but not
`cmd/loom/edit_304_test.go`.

## Fix
Apply the same pattern as 337: wait on screen state via ptytest/loomtest, scale timeouts with
`raceFactor`, no wall-clock sleeps.

## Acceptance
- `go test -race -count=20 -run Test304 ./cmd/loom/` passes in a 2-CPU container.
- Close GitHub #21 with the commit hash.
