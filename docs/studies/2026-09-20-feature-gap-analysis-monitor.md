---
title: Monitor Example Feature Gap Analysis
---

# Monitor Example Feature Gap Analysis

The monitor is a useful integration example, but several framework-shaped
concerns are implemented locally.

## Gaps and opportunities

- **System stat collectors.** `examples/monitor/monitor/watch.go` parses
  `/proc/stat` with `parseProcStat` and derives CPU percentage with
  `cpuPercentage`. The framework could provide Linux collectors and normalized
  metrics:

  ```go
  collector.CPU(ctx) (Metric, error)
  collector.Memory(ctx) (Metric, error)
  collector.GPU(ctx, device string) (Metric, error)
  ```

  A portable collector registry could expose `collector.System("cpu")` while
  keeping platform-specific implementations behind the API.

- **Metric history and sampling.** `state.go` owns bounded slices,
  `SampleHardware`, `SampleUsage`, cloning, and snapshot publication. This is
  generic monitor plumbing, not app behavior. Add a concurrency-safe series
  store with retention/capacity and named metric snapshots:

  ```go
  history := loom.NewMetricStore(loom.Retention(32))
  history.Publish("cpu", value, at)
  snapshot := history.Snapshot()
  ```

  It should support independent sampling rates and ensure redraws never mutate
  producer data.

- **Gauge and sparkline widgets.** `monitor.go` manually calls
  `graph.RenderBar` in `applySnapshot`, then formats percentages and brackets.
  `timeline` and `splitTimeline` similarly wrap `graph.RenderSparkline`.
  Existing `graph.RenderBar` and `graph.RenderSparkline` are good primitives,
  but Loom lacks widgets that bind a metric value/history to a cell region:

  ```go
  loom.Gauge{Value: metric("ram"), Min: 0, Max: 100, Width: 4}
  loom.Sparkline{Series: metric("cpu"), Width: 10, Min: 0, Max: 100}
  loom.MultiSeries{Series: []Series{"vram", "gtt"}, Width: 4}
  ```

  These could also be declared in YAML, removing row-index and column-index
  mutation from `applySnapshot`.

- **Multi-metric chart composition.** The `load` box has a special
  `vram/gtt` name and `splitTimeline(left, right)` branch in
  `monitor.go`. A chart model should declare multiple named series, labels,
  ranges, and layout rather than require application-specific string matching:

  ```go
  loom.Chart{Series: []loom.SeriesRef{{Name: "vram"}, {Name: "gtt"}},
      Range: loom.FixedRange(0, 100)}
  ```

- **Polling and lifecycle orchestration.** `startSources`, `sourceRuntime`,
  `Close`, error polling, and goroutine ownership in `watch.go` duplicate a
  general collector scheduler. `Pane.RunWatch` already owns redraw cadence,
  but collection remains manually wired. A framework API could combine both:

  ```go
  loom.RunMonitor(ctx, loom.Monitor{
      Collectors: sources, CollectEvery: time.Second,
      RedrawEvery: 50 * time.Millisecond,
      Update: func(*loom.Metrics, time.Time) error { ... },
  })
  ```

  It should coalesce missed ticks, cancel all workers, surface the first error,
  and preserve the example's rule that resize/input redraw without sampling.

- **Spec-driven monitors.** `watch.go` decodes `watch.yaml`, validates
  cadence, builds `collector.Spec`, creates histories, templates the title, and
  connects collector IDs to presentation behavior. This suggests a declarative
  monitor schema with `sources`, retention, cadence, metric transforms, and
  widget bindings. For example:

  ```yaml
  monitors:
    - metric: cpu
      source: system.cpu
      view: load.cpu
      widget: sparkline
  ```

  A `loom.BuildMonitor(io.Reader)` API could return a validated monitor model
  alongside `BuildWidget`, leaving application code only for custom collectors
  and transforms.

## Scope boundary

The example should continue to own domain-specific labels, simulated fallback
data, CLI policy, and custom metric transforms. Loom should own reusable
collection, history, scheduling, metric-to-widget binding, and declarative
validation. The current `collector` package, graph renderers, and
`Pane.RunWatch` are foundations; the missing layer is their integrated,
spec-aware monitor runtime.
