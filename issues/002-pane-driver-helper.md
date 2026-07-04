<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 002 — Ship an exported pane-driver helper

**Status:** Done in loom (2026-07-04) — consumer cutover tracked in 004

**Priority:** P2 — quality-of-life for consumers; not strictly release-blocking

## Problem

Driving a widget currently requires the open/run/close dance:

```go
pane, err := loom.New(w.ContentHeight())
if err != nil { return err }
err = pane.Run(w)
pane.Close()
item, ok := w.Selected()
```

Every consumer reinvents this. uzu already has a private `runPane` helper
(`uzu/main.go`) that also handles `Nav` navigation signals and guarantees the
pane is closed before any stdout write. That helper should live in loom, not be
copy-pasted into uman.

## Proposal

Export a driver from loom, e.g.:

```go
// RunPane opens a pane sized to w, runs w, always closes the pane, and returns
// the selection plus any navigation signal.
func RunPane(w Paneable) (Item, Nav, error)
```

This requires promoting uzu's `paneable` interface (`ContentHeight`, `Nav`,
`Selected`) into loom as an exported `Paneable` interface.

## Tasks

- [x] Define exported `Paneable` interface in loom (`driver.go`).
- [x] Add `RunPane` (`driver.go`). Signature landed as
  `RunPane(w Paneable) (item Item, ok bool, nav Nav, err error)` — the extra
  `ok` preserves `Selected()`'s abort/select distinction (a zero `Item` alone is
  ambiguous when an item legitimately has an empty `Name`).
- [x] Terminal restore on error/signal/panic: **loom owns it.** `Pane` installs
  a SIGINT/SIGTERM handler that restores and exits, and `Pane.Run` restores on
  panic before re-panicking (`pane.go`). `RunPane` always `Close()`s the pane
  before returning. No caller wiring required.
- [x] Update README example to use `RunPane`.
- [ ] uzu/uman switch their local helper to `loom.RunPane` (relates to 004 and
  uman issue 005) — happens in those repos after the tag.

## Resolved questions

- Should loom own the signal/panic terminal-restore handlers? **Yes — it already
  does**, via `Pane.installSignalHandler` and the `recover` in `Pane.Run`. The
  earlier lean toward a caller-wired `Restore()` is unnecessary; the extraction
  already carried uzu's handlers into `Pane`.
