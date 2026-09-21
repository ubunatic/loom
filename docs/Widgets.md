---
title: Loom Widgets and Framework Primitives
weight: 35
---

# Loom Widgets and Framework Primitives

Loom provides composable, dependency-light widget primitives designed for inline terminal UIs.
Widgets implement `loom.Widget` (`Draw`, `HandleKey`, `HandleMouse`) and integrate cleanly with `loom.Frame`, `loom.Pane`, and the layered compositor.

---

## 1. Split Layout (`loom.Split`)

`loom.Split` manages two child widgets partitioned horizontally or vertically with proportional ratio control, minimum bounds constraints, and customizable dividers.

```go
split := loom.NewSplit(leftWidget, rightWidget)
split.Orientation = loom.Horizontal // or loom.Vertical
split.Ratio = 0.40                  // 40% left, 60% right
split.MinFirst = 20                 // Minimum width/height for first pane
split.MinSecond = 30                // Minimum width/height for second pane
split.Divider = loom.DividerStyleThin
```

### Key Capabilities & Invariants
- **Ratio Control**: Dynamically adjust `Ratio` (0.0 to 1.0) via `SetRatio(r)`. Left/right or top/bottom arrow keys with `Ctrl` adjust ratios interactively.
- **Mouse Dragging**: Dragging the divider line smoothly recalculates the split ratio within constrained min bounds.
- **Focus Hierarchy (`FocusContainer`)**: `Split` implements `FocusContainer`. Pressing `Tab` / `Shift-Tab` traverses child focusable elements cleanly through arbitrary nested split trees.
- **Mouse Coordinate Translation**: Mouse click and drag events are translated relative to each child pane's local coordinates before dispatch.

---

## 2. Tabs Widget (`loom.Tabs`)

`loom.Tabs` provides a tabbed container with header navigation, active styling, and nested content panels.

```go
tabs := loom.NewTabs(
    loom.Tab{Title: "Status", Child: statusView},
    loom.Tab{Title: "Logs", Child: logsView},
)
```

### Dynamic Lifecycle & Declarative Navigation
- **Dynamic Mutation**: Modify tabs at runtime using `Add(Tab)`, `Insert(index, Tab)`, `Remove(index)`, `SetTabs(...Tab)`, and `Select(index)`. Active index and focus are safely preserved or clamped.
- **Declarative Key Bindings (`TabsKeys`)**:
  ```go
  tabs.Keys = loom.TabsKeys{
      Previous: []string{"h", "left"},
      Next:     []string{"l", "right"},
      Cycle:    "ctrl-t",
      Select:   []string{"1", "2", "3", "4"}, // direct numeric selection
  }
  ```
- **Fallback Compatibility**: Existing `SwitchKey` configurations remain fully supported.

---

## 3. Metrics & Monitoring Primitives

Loom provides concurrency-safe metric history storage and bound visualization widgets.

### Thread-Safe Metric Store (`loom.MetricStore`)
`MetricStore` maintains bounded rolling time-series values for named metrics:

```go
store := loom.NewMetricStore(loom.Retention(60))
store.Publish("cpu_usage", 42.5, time.Now())

// Immutable snapshot for safe rendering in UI loops
snap := store.Snapshot()
val := snap.Latest("cpu_usage")
series := snap.Series("cpu_usage")
```

### Bound Widgets (`Gauge` and `Sparkline`)
```go
// Direct bar gauge bound to metric value
gauge := &loom.Gauge{
    Store: store,
    Name:  "cpu_usage",
    Min:   0,
    Max:   100,
    Width: 10,
}

// Rolling sparkline bound to time-series history
sparkline := &loom.Sparkline{
    Store: store,
    Name:  "cpu_usage",
    Min:   0,
    Max:   100,
    Width: 20,
}
```

---

## 4. Filesystem & OS Primitives (`loom.Directory` and `OpenFile`)

Loom extracts common terminal file-browsing and launch operations into clean, portable helpers:

- **Structured Directory Scanner (`ReadDirectory`)**: Reads directory entries, detects symlinks/directories, sorts entries with `..` parent navigation, and preserves path mapping.
- **Terminal-Safe Path Escaping (`DisplayPath`, `QuoteUnprintable`)**: Escapes unprintable control codes and formats paths safely for terminal display columns without corrupting layouts.
- **Injectable Platform File Opener (`OpenFile`, `FileOpener`)**: Dispatches file launch commands (`xdg-open`, `gio`, `open`, `start`) detached from the terminal process, with mockable injection for deterministic unit tests.

---

## 5. Startup & Splash Transition Runner (`Pane.RunStartup`)

`RunStartup` orchestrates application startup loading screens, asynchronous task completion, final frame hold delays, and deterministic handover to the main application widget without ad-hoc `time.Sleep` loops:

```go
cfg := loom.StartupConfig{
    Splash:     splashController,
    View:       splashView,
    Next:       mainAppWidget,
    Cadence:    loom.Cadence{CollectInterval: 50 * time.Millisecond},
    Transition: loom.ImmediateTransition(),
}

err := pane.RunStartup(ctx, cfg)
```

- **One-Shot Writer Previews (`loom.RenderTo`)**: Renders static widget snapshots directly to an `io.Writer` (e.g. for non-interactive CLI `--version` or `--help` banners).

## 6. Primitives Added by the 2026-09 Lean Sprints

Details live in the closed tickets and their `docs/progress/<ticket>/` frames.

| Primitive | Ticket | Use |
|---|---|---|
| `loom.ParseANSI(s) []Cell`, `Canvas.WriteANSI(x, y, text)` | 034 | Draw SGR-styled text (recordings, `docs/data/*.ansi`) into a canvas. |
| `Themeable` (`ApplyTheme(ThemeColors)`) | 061 | Hosts restyle Frame, Tabs, Stack, Grid, Choice, Table and filebrowser through one call. |
| `NewWidget` factories in `examples/split`, `examples/tabs`, `internal/examplesreg` | 062 | Hosted vs standalone examples; `loom-demo --hosted`, `loom-bench` smoke. |
| `Frame` `FillHeight` (stacked fill) | 051 | A frame box expands to the available content height. |
| `Canvas.DrawBorder`, `Canvas.DrawBox`, `spec/box.yaml` | 037 | Spec-driven box glyphs; do not duplicate glyphs in Go. |
| `Ticker` (`TickInterval`, `Tick`), `Pane.Invalidate()` | 060 | Periodic redraw without pane ownership; invalidate is goroutine safe. The pane must not reset its tick timer on every event. |
| `Choice.MouseTextOnly` | 093 | Mouse selects only on rendered item text. |
| Scrollbar drag in `View` and `Split` | 092 | Drags stay with the widget that started them. |
| [Key defaults](KeyDefaults.md) and the decoder audit | 088 | Which keys are decoded and which are terminal limitations (Ctrl-I, Ctrl-J, Ctrl-M). |

Mouse coordinates: events reaching widgets are 0-based, PTY SGR mouse reports are 1-based.
In tests locate screen text by runes or display width, never byte offsets.
