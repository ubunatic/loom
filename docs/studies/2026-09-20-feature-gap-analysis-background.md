# Background Example Feature-Gap Analysis

Scope: `examples/background/` compared with Loom’s current pane, widget, frame,
and background APIs.

## Already framework-owned

- `loom.Pane.Run` in `pane.go` owns the event loop and redraw scheduling.
- `loom.AnimatedBackground`, `loom.BackgroundCadence`, and the ticker in
  `pane.go` own animated-background timing and cancellation. `ReduceMotion`
  also exists.
- `Canvas.ComposeBackground` and `PaintDecoration` in `canvas.go` provide the
  compositor hook and protected-cell behavior.
- `loom.NewAstraBackground` in `background.go` is already the preset; the
  example does not duplicate Astra rendering.

## Gaps and opportunities

- **Full-screen pane sizing:** `background.Run` calls `loom.New(1000)` and
  relies on clamping, then sets `pane.MaxCols = 0`. This is an app workaround
  for “use the current terminal size.” Add an explicit constructor or option:

  ```go
  pane, err := loom.NewScreen(loom.ScreenOptions{FullWidth: true})
  // or: loom.New(loom.FullScreen)
  ```

- **Widget composition with a `Frame`:** `widget` in
  `examples/background/background/background.go` manually stores `*loom.Frame`,
  draws it at a calculated rectangle, and delegates `HandleKey`/`HandleMouse`.
  A reusable layout wrapper would make this integration declarative:

  ```go
  root := loom.Overlay(statusWidget, loom.FrameWidget(frame, loom.Rect{Y: 7}))
  ```

  More generally, expose a first-class `Frame`/`Widget` adapter that handles
  child bounds and event forwarding.

- **Metrics presentation:** the example formats `RenderMetrics` directly into
  a status line (`widget.Draw`). Loom exposes collection but not a compact
  diagnostic widget or formatter. Add an opt-in helper:

  ```go
  loom.NewMetricsWidget(metrics, loom.MetricsView{FPS: true, Redraw: true})
  // or: metrics.Format(loom.MetricsFormatCompact)
  ```

- **Background presets/configuration:** `NewAstraBackground` is useful, but
  `AstraBackground` has no public preset/options for density, palette, seed, or
  lifecycle. Add a spec-backed constructor while retaining the preset:

  ```go
  loom.NewAstraBackgroundWith(loom.AstraOptions{Seed: 1, Density: 0.18})
  ```

- **Declarative background attachment:** assigning `pane.Background` is simple,
  but selecting a preset, reduced-motion policy, and cadence remains scattered
  across app setup. A `PaneOptions{Background: ..., ReduceMotion: ...}` or
  `pane.SetBackground(...)` API could centralize validation and lifecycle.

## Conclusion

The example does not hard-code a missing animation loop or compositor. The main
framework opportunities are ergonomic: a full-screen constructor, a reusable
frame/widget composition adapter, and optional metrics/preset configuration.
