---
title: Winch Feature Gap Analysis
---
# Winch Feature Gap Analysis

## Scope

Review of `examples/winch/winch/winch.go` and its PTY tests for behavior that
belongs in Loom rather than in a diagnostic app.

## Already framework-owned

- Resize-event coalescing is implemented in `Pane.Run` (`pane.go`) using the
  buffered `Pane.winch` channel and `ResizeConfig.Coalesce`.
- Burst measurement and adaptive guard policy are already reusable APIs in
  `resize.go`: `WinchMeter`, `AdaptiveGuardN`, `Pane.WinchRate`, and
  `Pane.EffectiveGuardN`.
- Resize mode switching/configuration is exposed through
  `Pane.SetResizeMode`, `Pane.ToggleResizeMode`, `Pane.ResetResizeModes`, and
  `ResizeConfig` (`resize.go`).
- Render timing and rolling rate counters are exposed by `RenderMetrics`
  (`pane.go`), including `RedrawAvg`, `RedrawMax`, `LoomFPS`, and
  `BytesPerFrame`.

## Gaps and opportunities

- **Diagnostic snapshot API.** `drawStressPanel` directly combines
  `WidthGuardActive`, `AutoFullscreenReasons`, `Screen`, `EffectiveGuardN`,
  `WinchRate`, and `RenderMetrics` into display strings. Add a stable snapshot
  type so diagnostic UIs do not depend on internal state assembly:

  ```go
  type ResizeDiagnostics struct {
      Active bool
      Events int
      Rate float64
      GuardN int
      GuardSource GuardSource
      Screen ScreenMode
      FullscreenByWidth, FullscreenByHeight bool
  }

  func (p *Pane) ResizeDiagnostics() ResizeDiagnostics
  ```

- **Resize/layout mode helper.** Winch hard-codes the `r.W >= 72` breakpoint,
  `modesW := 40`, stacked-panel height arithmetic, and the status-row reserve
  in `App.Draw`. These are generic responsive-pane concerns. Consider a small
  width-aware layout helper:

  ```go
  type LayoutMode int // Compact, Wide
  func ResponsiveMode(width, wideAt int) LayoutMode
  func SplitVertical(r Rect, leftWidth, gap int) (left, right Rect, ok bool)
  ```

- **Burst lifecycle callback.** The app can only poll `WidthGuardActive` and
  infer settling from displayed values. A callback or event stream would let
  applications react without polling:

  ```go
  type ResizeEvent struct { Width, Height int; Burst bool; Rate float64 }
  func (p *Pane) OnResize(fn func(ResizeEvent))
  ```

- **Metrics presentation helpers.** `ms`, the one-second render summary, and
  the `fps`/bytes-per-frame formatting in `drawStressPanel` are likely useful
  to any profiling/demo app. Keep `RenderMetrics` as the data API, but add
  narrowly scoped formatting or a metrics snapshot with a documented window:

  ```go
  func (m RenderMetrics) Snapshot() RenderSnapshot
  ```

## Recommendation

Prioritize `ResizeDiagnostics()` and the responsive layout helpers. They remove
Winch's direct knowledge of Pane internals and make future resize demos smaller.
Do not duplicate coalescing, burst measurement, adaptive guard, or counters in
the example; those are already correctly framework-owned.
