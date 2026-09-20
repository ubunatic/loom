---
title: Splash Example Feature Gap Analysis
---

# Splash Example Feature Gap Analysis

Date: 2026-09-20. Scope: `examples/splash/` versus the existing Loom splash APIs.

## Findings

The example is mostly a consumer of Loom’s intended primitives, but it still
hard-codes application-level startup orchestration in
[`examples/splash/splash/splash.go`](../../examples/splash/splash/splash.go):

- **Startup transition runner**: `runWatch` manually creates `loom.New(10)`,
  starts `SplashController`, builds `loom.Cadence`, copies
  `sc.Snapshot()` into `view.ApplySnapshot`, and calls `pane.RunWatch`.
  This is reusable framework lifecycle code, not splash-specific UI.
- **Completion hand-off delay**: the goroutine after `sc.Done()` sleeps for
  `100 * time.Millisecond` before cancelling the pane. This is a fragile,
  duplicated transition policy and can race the final render.
- **Animation timing**: `runWatch` hard-codes `TickInterval: 80ms`,
  `HoldDuration: 600ms`, and a `50ms` collect/redraw cadence. The controller
  already owns tick and hold defaults, but the complete animation profile is
  not a framework abstraction.
- **Default initialization tasks**: `defaultTasks` hard-codes provider names,
  symbols, and simulated durations. The task model is in Loom, but a standard
  startup-task registry / adapter is missing.
- **Terminal sizing and one-shot rendering**: `runShowOnce` duplicates terminal
  width/height probing (`terminalWidth`, `terminalHeight`), fallback dimensions,
  clipping, and ANSI reset stripping. A framework render-to-writer helper could
  provide consistent non-interactive output.
- **Modal intro layer**: `SplashView` is installed as the pane root, so the
  example has no reusable “show this startup layer over the eventual root and
  reveal the root on completion/skip” primitive. The current code only cancels
  the splash loop; it does not model the destination view.

## Proposed Loom APIs

Rough signatures for the highest-value gaps:

```go
type StartupConfig struct {
    Splash    *SplashController
    View      *SplashView
    Next      Widget
    Cadence   Cadence
    Transition Transition
}

type Transition interface {
    Enter(ctx context.Context, from, to Widget) error
}

func (p *Pane) RunStartup(ctx context.Context, cfg StartupConfig) error
```

`RunStartup` should own snapshot application, redraw cadence, completion/skip
ordering, and terminal cleanup. A built-in short final-frame hold or explicit
transition duration would replace the example’s `time.Sleep(100ms)`.

```go
type StartupTaskRegistry struct { /* named ProviderTask set */ }
func NewStartup(tasks ...ProviderTask) *SplashController
func (sc *SplashController) Wait(ctx context.Context) error
```

This would standardize task registration and completion waiting without making
applications manually coordinate `Done`, cancellation, and worker lifetime.

```go
func RenderTo(w io.Writer, view Widget, size Size, opts RenderOptions) error
func TerminalSize(w io.Writer, fallback Size) Size
```

These helpers would consolidate `runShowOnce`’s sizing, clipping, and ANSI-safe
one-shot output path.

## Priority

1. Add `Pane.RunStartup` (or an equivalent startup-layer runner) with a
   deterministic completion transition.
2. Add a composable modal/overlay startup layer that reveals a destination
   widget on completion or dismissal.
3. Add shared terminal-size and writer-render helpers for CLI splash previews.
4. Consider a task registry only if multiple applications need the same
   provider initialization conventions.
