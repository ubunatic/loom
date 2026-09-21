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

---

## Resolved Design Additions (host decisions, binding for the sprint)

- Follow section 2 as written. No `Host`/`SetHost`.
- Testable core: a pure walker over the widget tree that (a) returns the shortest positive `TickInterval` (0 if none)
  and (b) `tickTree(root, now)` calling `Tick` on each `Ticker` that is due. Per-widget last-tick times are kept by the
  pane in a small map keyed by the widget value (pointer identity), so a child ticks no more often than it asked even
  though the pane timer runs at the shortest interval. Tests drive it with a fake clock (explicit `now` values); no
  test sleeps for more than a few milliseconds.
- `Pane.Invalidate()`: a `cap(1)` channel created where the pane is constructed; make it safe when the pane was
  built without that constructor (lazy init guarded by `sync.Once`, or a documented zero-value-safe path). A call
  must never block and never be lost while the loop runs (coalescing is fine).
- The pane's tick timer must stop when no widget is a `Ticker` (no timer, no wakeups, no repaint).
- Concurrency proof: `go test -race` on the new tests, N=100 goroutines calling `Invalidate` before the loop runs,
  during the loop, and after it stopped; no goroutine leak (check `runtime.NumGoroutine` before and after, with a
  short settle).

## Milestones (lean sprint, developer: luna:low)

Host reviews only diffs, test output and evidence frames; this ticket is the only channel. Root-package tests needing
`/dev/tty` fail before this work; ignore them. Commit each milestone (message ends '(issue 060 MX)'), stage only your
files, never docs/README.md, no stray binaries in the repo root; if `.git/index.lock` blocks the commit, stage your
files and say so (the host commits). Evidence: frames produced by code, gated on env `LOOM_EVIDENCE=1`, written to
repo-root `docs/progress/060/` (find the root by walking up to `go.mod`); run only this ticket's evidence test; view
frames ANSI-stripped before finishing; labels on their own rows. gofmt. No dead code.

### M1 - Ticker, forwarding, tree walker
- `Ticker` in `widget.go`; `Tabs` (active child only), `Stack`, `Grid`, `Frame` (all children) forward; walker and
  `tickTree` with fake-clock tests: shortest interval, per-child cadence, inactive tab never ticked, tab switch moves
  ticks, non-Ticker tree yields no interval.
- Evidence: `M1-tick-0.ansi`, `M1-tick-1.ansi`, `M1-tick-2.ansi` (a small counter widget inside Tabs rendered after 0, 1,
  2 driven ticks, plus `M1-tabs-switched.ansi` showing the other tab's counter after a switch and further ticks).

### M2 - Pane integration, Invalidate, guards
- `Pane.run` uses the timer and the walker; `Pane.Invalidate`; `RunWatch` keeps its public behavior; idle-redraw guard
  (a non-Ticker widget causes no extra repaint); concurrency tests above under `-race`.
- Run `go test -race` for the root package tests you added and `go test ./...` (root `/dev/tty` failures are pre-existing).
- Evidence: `M2-pane-ticks.ansi` (final frame of a real `Pane` run headless or in the PTY helper showing the counter
  advanced by ticks) and `M2-invalidate.ansi` (frame after an `Invalidate` from another goroutine).

### M1-M2 Review (host)
Code committed by the host (index.lock). Build, vet, `go test .` and `go test -race` on the new tests pass. Accepted:
`Ticker`, forwarding, per-widget cadence map, lazily initialized coalescing `Invalidate`. Defects:
- Timer starvation: `Pane.run` stops and resets the tick timer on EVERY loop iteration, so any event storm (mouse
  motion, `Invalidate`, key repeat) restarts the countdown and ticks never fire. Reset only when the shortest interval
  changed, and re-arm after a fire; keep the timer running across unrelated events.
- The acceptance list is only partly covered (4 tests): missing are Stack/Grid/Frame forwarding with mixed intervals,
  a real `Pane` run that ticks (integration), the idle-redraw guard (non-Ticker tree: no timer, no repaint),
  Invalidate under 100 goroutines before, during and after the loop with a goroutine-leak check, `RunWatch` unchanged,
  and the event-storm test for the starvation bug above (write it first, watch it fail, then fix).
- Missing: all evidence frames.

### M3 - Pre-Work / Required Refinements
1. Failing test first for the starvation bug (a Ticker with a 20 ms interval, an event source hammering `Invalidate`
   or input every 1 ms for 200 ms: at least 5 ticks must happen), then fix.
2. Add every missing test from the list above, under `-race`.
3. Evidence (gated on `LOOM_EVIDENCE=1`, repo-root `docs/progress/060/`; a counter widget inside Tabs):
   `M1-tick-0.ansi`, `M1-tick-1.ansi`, `M1-tick-2.ansi` (rendered after 0/1/2 driven ticks with the fake clock),
   `M1-tabs-switched.ansi`, `M2-pane-ticks.ansi` (final frame of a real `Pane` headless run after real ticks),
   `M2-invalidate.ansi` (frame after an `Invalidate` from another goroutine). Labels on their own rows; view ANSI-stripped.
4. Commit '(issue 060 M3)'; if git is blocked say so and leave files staged.
