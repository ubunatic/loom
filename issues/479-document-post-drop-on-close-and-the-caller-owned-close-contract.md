# 479 — Document Post drop-on-Close and the caller-owned Close contract

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Documentation
**Related**: 458, 320, 372, 478

---

## 1. Problem & Motivation
Two doc comments promise more than the code does:

- `Post` (`pane_run.go:45`) says it "never drops callbacks; queued callbacks run in FIFO order". `Close`
  discards pending callbacks (`pane_terminal.go:196`) and `runPosted` abandons its batch after a close.
  458 batch 9 canaries B1 and B2 (`TestB9BSlowCloseExternal`, `TestB9BCloseInCallback`): 1 of 10 and
  4 of 10 callbacks ran. Dropping on Close is reasonable; the comment needs the shutdown exception.
- `driver.go:27` says "Terminal restore on SIGINT/SIGTERM and panic is handled by Pane itself". After
  SIGINT, `Pane.Run` returns nil with the alt screen, hidden cursor and raw mode still set (canary C5);
  the caller's `Close` restores them (C6). `New`, `RunWatch`, the examples and `RunPane` (`driver.go:29`)
  all follow a caller-owned Close contract, so only the comment is wrong. Panic restore inside `Run`
  works (C1, C2).

Evidence: `docs/studies/tuios/report-testing.md`.

## 2. Technical Specification / Findings
- `Post`: accepted callbacks run in FIFO order while the pane is open; `Close` discards the rest, and
  `Post` after `Close` is a no-op.
- `driver.go` and `Pane.Run`: `Run` restores the terminal on panic; on return for any other reason
  (error, signal, quit) the caller must call `Close`, which `RunPane` does.
- If 478 changes the queue policy, keep this text in step.

## 3. Implementation & Verification Plan
/goal Both comments (and any doc in `docs/` that repeats them) state the real contract; stop and report
when blocked on a decision or denied permission; on a loop or no progress, clear the goal and write a
failure report.

- Re-check live code first; `harnez find docs "Post Close SIGINT restore"` for copies.
- `make lint`.
