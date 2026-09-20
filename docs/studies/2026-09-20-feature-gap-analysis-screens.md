---
title: Screens Example Feature-Gap Analysis
---
# Screens Example Feature-Gap Analysis

Date: 2026-09-20. Scope: `examples/screens/`.

## Findings

- `examples/screens/screens/screens.go`: `HandleKey` manually translates keys into
  `ResizeConfig.FullScreenBuffer`/`AltScreen` combinations. `Pane.SetScreenMode`
  exists, but the example still needs a higher-level toggle/cycle API that
  preserves the inline/full/alternate relationship.
- `examples/screens/screens/screens.go`: `screenName`, mode labels, and the
  `Screen:` status row duplicate framework knowledge already represented by
  `loom.ScreenMode` and `Pane.Screen()`. A framework formatter or screen-state
  view model would make diagnostics consistent.
- `examples/screens/screens/screens.go`: `detected` clones `ResizeConfig`,
  disables one detector at a time, and recomputes width/height reasons. This is
  presentation code for `Pane.AutoFullscreenReasons()` and should be exposed as
  a reusable detector result, including the winning reason(s).
- `examples/screens/screens/screens.go`: `NewApp` and `apply` directly manage
  `Pane.ResizeConfig`, `MaxCols`, wanted height, and background updates. A pane
  layout request could atomically update desired geometry, mode policy, and
  redraw state.
- `examples/screens/screens/screens.go`: `resizeWidth`, the height key cases,
  and the button `stepper` duplicate bounded geometry steppers. The framework
  has no reusable “wanted size with min/max/step” abstraction.
- `examples/screens/screens/screens.go`: `HandleMouse` converts coordinates,
  searches button rectangles, and maps clicks back to keys. A small button or
  action-control helper would remove this repeated example wiring.
- `examples/screens/screens/screens.go`: CLI defaults manually copy
  `DefaultResizeConfig()` into flags, then patch fields one by one. A
  framework-supported resize-policy builder/flag adapter could keep spec-backed
  defaults and runtime controls aligned.
- `examples/screens/screens/screens_pty_test.go`: the example verifies escape
  sequences (`?1049h`/`?1049l`) and stale-row cleanup indirectly. Expose a
  screen-transition event/state hook so applications and tests can observe
  mode transitions without parsing terminal output.

## Proposed loom APIs

```go
type ScreenPolicy struct {
    Mode       ScreenMode // Inline, Full, or Alt
    Auto       bool
    ByWidth    bool
    ByHeight   bool
    MarginCols int
    MarginRows int
    MinPercent int
    Alt        bool
    LeakGuard  bool
}

func (p *Pane) SetScreenPolicy(policy ScreenPolicy)
func (p *Pane) ToggleScreenMode() ScreenMode
func (p *Pane) CycleScreenMode() ScreenMode
func (p *Pane) ScreenStatus() ScreenStatus
```

`ScreenStatus` should contain the effective mode, requested mode, terminal and
wanted geometry, and detector reasons (`Width`, `Height`). This would cover the
example’s `screenName`, `detected`, and status-row plumbing.

```go
type SizeStepper struct {
    Value, Min, Max, Step int
}

func (s *SizeStepper) Add(delta int) int
func (p *Pane) SetWantedSize(width, height int)
```

Use `SetWantedSize` for bounded pane geometry and let the pane clamp against
terminal dimensions while retaining the wanted size across resizes.

```go
func (p *Pane) OnScreenTransition(func(from, to ScreenMode) error)
func (p *Pane) ScreenModeLabel(mode ScreenMode) string
```

The transition hook supports application diagnostics and PTY tests; the label
helper standardizes inline/primary/alternate terminology.

These are opportunities, not required changes to the example. Existing
`ResizeConfig.QuasiFullscreen`, `Pane.SetScreenMode`, and
`Pane.AutoFullscreenReasons` are the correct low-level foundation.
