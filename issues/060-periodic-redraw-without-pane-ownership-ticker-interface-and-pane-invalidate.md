# 060 — Periodic redraw without pane ownership: Ticker interface and Pane.Invalidate

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Feature
**Related**: `pane.go`, `widget.go`, `tabs.go`, `yaml.go`,
`examples/monitor/monitor/watch.go`, `examples/splash/splash/splash.go`,
[055](055-add-a-tab-panel-widget-for-pane-hosting.md),
[056](056-embed-real-applications-as-pty-hosted-widgets-tmux-screen-style.md),
[064](064-convert-splash-and-monitor-examples-to-hostable-widgets.md)

---

## 1. Problem & Motivation

Third library-side prerequisite of the "examples as just widgets" initiative.
Today the only way a widget gets periodically redrawn is for its app to own
the pane:

- `Pane.run` paints only when `dirty` (`pane.go:466-484`), set by key/mouse
  input, SIGWINCH, and the `frames` channel.
- The only producer for `frames` is `Pane.RunWatch(ctx, root, cadence, collect)`
  (`pane.go:349-367`) — **one** cadence and **one** collect callback, pane-global.
- `monitor` drives it from its own YAML cadence and separately runs collector
  goroutines feeding `collector.History`
  (`examples/monitor/monitor/watch.go:50,131-138,186-204`).
- `splash`'s watch path runs at a 50 ms cadence
  (`examples/splash/splash/splash.go:186-235`), ending when `sc.Done()` fires.
- `treemap` bypasses the pane entirely with its own `time.Ticker` around
  `RawScreen.Draw` (`examples/treemap/treemap/treemap.go:112-123`) — see
  [065](065-ansi-styled-rows-widget-and-treemap-conversion.md).
- There is **no** `Invalidate`/`RequestRedraw` anywhere in the library
  (verified by grep: zero hits).

Consequences for hosting: a non-pane-owning widget can neither request a
repaint nor register a tick, and `RunWatch` structurally supports exactly one
live widget per pane — monitor and splash could not be two live tabs. There is
also no visibility signal, so an inactive tab would keep working for nothing.

## 2. Design — resolved decisions

The repo owner's instinct ("every app has its own render loop; just ask
everyone to give us a function to call for rendering, and hook it into one
loop") **converges with the review's recommended option** — a pull-based tick
driven by the single pane event loop. Adopt it, in two small pieces.

### 2.1 `Ticker` — the pull-based render/update callback

```go
// Ticker is an optional interface for widgets that need periodic updates.
// Pane calls Tick from the event-loop goroutine at (at most) the shortest
// TickInterval requested by the widget tree, then repaints.
// TickInterval() <= 0 disables ticking for this widget.
type Ticker interface {
    Widget
    TickInterval() time.Duration
    Tick(now time.Time)
}
```

`Pane.run` walks the root for the shortest requested interval and fires `Tick`
from the event-loop goroutine — the same threading guarantee `RunWatch`'s
`collect` already documents (`pane.go:346-348`), so migrating the existing
collect closures is a rename, not a concurrency redesign.

`Tabs` implements `Ticker` by forwarding **only to the active child**
(`active()` already exists, `tabs.go:96-101`). That solves inactive-tab waste
for free, with no separate visibility API. `Stack`/`Grid`/`Frame` forward to
all children (all are visible).

This is the owner's "one loop, everyone hands us a function", expressed as an
interface rather than a registered `func` so that (a) it composes through the
existing composite-delegation pattern used by `Focusable`/`ContentHeighter`,
(b) the widget needs no reference to its host, and (c) there is nothing to
unregister when a widget goes away.

### 2.2 `Pane.Invalidate()` — the async escape hatch

```go
// Invalidate requests a repaint. Safe to call from any goroutine; never blocks.
func (p *Pane) Invalidate()
```

Non-blocking send on a new `cap(1) chan struct{}` handled as one extra case in
`run`'s `select` (`pane.go:495`), exactly like `<-frames`. Needed for widgets
with genuinely async data sources — monitor's collector goroutines already
exist and push into `collector.History` off the event loop; polling them on a
tick works, but a direct wake-up avoids cadence/latency guessing.

`RunWatch` stays as-is for the single-root case; it is reimplemented on top of
the tick path if that is clean, otherwise left untouched (it is public API).

### 2.3 Rejected: `Host`/`SetHost(Host)` injection

Passing an `interface{ Invalidate() }` to widgets via an optional
`SetHost(Host)` forwarded by composites is more general but invites widgets to
own goroutines and lifecycle, which is what makes the examples unhostable in
the first place. It also extends a precedent that is already slightly smelly —
`Router` holding a `*Pane` (`yaml.go:208`, wired at `yaml.go:527-529`). Explicitly
rejected; `docs/Go.md`'s no-speculative-generality rule applies. Widgets that
truly need async wake-ups get the `*Pane` handed to them by the code that
constructs *and* runs them, not by a framework-wide injection path.

## 3. Verification & Acceptance

- `Ticker` in `widget.go`; `Pane.run` computes the shortest interval across the
  widget tree and calls `Tick` from the event-loop goroutine before repainting.
- `Pane.Invalidate()` exists, is documented as goroutine-safe and
  non-blocking, and does not deadlock when the loop is busy (test: N
  concurrent calls, loop not running yet, no goroutine leak).
- `Tabs` forwards ticks to the active child only — test asserts an inactive
  child's `Tick` is never called, and that switching tabs moves the ticks.
- `Stack`/`Grid`/`Frame` forward ticks to all children; a tree with mixed
  intervals ticks at the shortest one, each child no more often than it asked.
- A widget not implementing `Ticker` causes no ticking and no extra repaints
  (idle-redraw regression guard, cf. 026).
- `RunWatch`'s public behavior is unchanged; existing monitor/splash watch
  tests and `make watch-pty` still pass.
- `go test ./...`, `go test -race ./...` and `go vet ./...` pass.
