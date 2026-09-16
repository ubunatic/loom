# 064 — Convert splash and monitor examples to hostable widgets

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `examples/splash/splash/splash.go`,
`examples/monitor/monitor/monitor.go`, `examples/monitor/monitor/watch.go`,
`collector/`, `internal/examplesreg/registry.go`,
[055](055-add-a-tab-panel-widget-for-pane-hosting.md),
[056](056-embed-real-applications-as-pty-hosted-widgets-tmux-screen-style.md),
[058](058-widget-declared-pane-requirements-panerequest.md),
[060](060-periodic-redraw-without-pane-ownership-ticker-interface-and-pane-invalidate.md),
[062](062-example-widget-factories-newwidget-for-split-and-tabs-hosted-loom-demo-mode-headless-bench-smoke.md)

---

## 1. Problem & Motivation

Third conversion wave of the "examples as just widgets" initiative: the two
examples with **live data**, which is what makes them depend on 060's redraw
work rather than just on plumbing.

Neither produces a `loom.Widget` in its headline (show-once) mode at all:

- `monitor`'s non-watch path renders rows and `fmt.Fprintln`s them to `out`
  (`examples/monitor/monitor/monitor.go:58-75`); only `runWatch`
  (`watch.go:70-139`) builds a real `*loom.Frame`.
- `splash`'s non-watch path does the same (`splash.go:145-184`); only
  `runWatch` builds a `*loom.SplashView` (`splash.go:186-235`).

Further blockers:

- **CLI parsing is entangled with execution.** Both build a cobra command whose
  `RunE` closure *is* the app (`splash.go:270-297`,
  `monitor.go:183-210`). Extracting a factory requires introducing an options
  struct first.
- **Pane ownership for redraw.** splash drives `RunWatch` at a 50 ms cadence;
  monitor drives it from a YAML-derived cadence (`watch.go:50,131-138`) *and*
  owns collector goroutines pushing into `collector.History`
  (`watch.go:186-204`), *and* sets pane height and `MaxCols`
  (`watch.go:99-138`).
- **splash ends itself.** Its loop terminates when `sc.Done()` fires
  (`splash.go:209-217`). Hosted, that must close its own tab, not the host pane.

## 2. Design — resolved decisions

### 2.1 Options struct per example

Introduce `splash.Options` / `monitor.Options` holding what the cobra flags
currently feed into the `RunE` closure. `Execute`/`Run` parse flags into the
struct; `NewWidget(args []string)` and a lower-level
`NewWidgetFromOptions(Options)` build the widget. Cobra stays in the standalone
entry point only. This is a prerequisite for both conversions and is the bulk
of the mechanical work.

### 2.2 Show-once mode stays a writer, hosted mode is the widget

Do **not** try to make the `fmt.Fprintln` path hostable. Show-once is a
legitimate non-TUI CLI mode; refactor it so both modes build the *same* widget
and show-once renders it headlessly via `loom.Render(w, cols, rows)`
(`screenshot.go:11`) and prints the resulting lines. That removes the duplicate
rendering path instead of preserving it, and gives the golden tests one source.

### 2.3 Redraw via `Ticker` (060)

- splash implements `Ticker` with its existing 80 ms spec tick
  (`spec/defaults.yaml` `splash.tick_interval`) / 50 ms cadence; its `collect`
  closure becomes `Tick(now)`.
- monitor implements `Ticker` with its YAML-derived cadence; its `collect`
  closure becomes `Tick(now)`. The collector goroutines keep running as they do
  today (they are the app's own data layer, not pane machinery) and call
  `Pane.Invalidate()` only where the tick cadence is demonstrably too coarse —
  otherwise the tick reads `collector.History` and no async wake-up is needed.
- Neither example calls `Pane.RunWatch` any more; `RunWatch` remains public API
  for direct pane users.

### 2.4 Lifecycle

splash's completion (`sc.Done()`) is reported through `Tabs.OnChildQuit`
(057): hosted, it closes its tab; standalone, the `Run` wrapper quits the pane
as it does today. Pane height and `MaxCols` move into `PaneRequest` (058).

### 2.5 Sequencing

splash first (simpler: one controller, no collectors), monitor second.

## 3. Verification & Acceptance

- `splash.NewWidget` / `monitor.NewWidget` exist and are registered in
  `examplesreg`; both `Run` paths are thin wrappers around the same widget.
- Show-once output for both examples is unchanged against existing golden
  tests, now produced via `loom.Render` rather than a second render path.
- Watch mode is unchanged standalone: `make watch-pty` passes (redraw, resize,
  toggles, terminal restoration), and existing monitor timing/provenance tests
  (012/013/027 coverage) still pass.
- Hosted in loom-demo: monitor and splash can be **two live tabs at once**
  (the case `RunWatch` structurally could not support), and an inactive tab
  receives no `Tick` (asserted by a counter in a test).
- splash completing closes its own tab and leaves the host running.
- No idle-redraw regression (cf. 026): an idle hosted tree repaints only on
  ticks its widgets asked for.
- `go test ./...`, `go test -race ./...` and `go vet ./...` pass;
  `make install` run afterwards.
