<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 002 — Ship an exported pane-driver helper

**Status:** Open

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

- [ ] Define exported `Paneable` interface in loom.
- [ ] Add `RunPane` (port `uzu/main.go:runPane`).
- [ ] Ensure terminal restore on error/signal/panic is handled or clearly
  documented as the caller's responsibility.
- [ ] Update README example to use `RunPane`.
- [ ] uzu/uman switch their local helper to `loom.RunPane` (relates to 004 and
  uman issue 005).

## Open questions

- Should loom own the signal/panic terminal-restore handlers, or leave that to
  the consuming binary? (uzu currently owns them in `main.go`.) Leaning: loom
  offers a `Restore()` / deferred cleanup, consumer wires the signal handler.
