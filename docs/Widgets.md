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

### Shared File Navigation Pane (`examples/filebrowser/filebrowser.NavigationPane`)
The filebrowser example package provides a reusable `NavigationPane` for directory lists. It is a widget that hosts can place directly in a `Frame` box; it owns a stable `loom.Choice` internally and updates its items when the directory changes.

```go
nav, err := filebrowser.NewNavigationPane(dir, filebrowser.NavigationPaneOptions{
    OnSelection: func(entry loom.FileEntry) { /* update preview or metadata */ },
    OnActivate:  func(entry loom.FileEntry) { /* activate a selected file */ },
    OnOpen:      func(directory loom.Directory) { /* react to navigation */ },
    OnQuit:      func() { /* handle ESC at the configured root */ },
})
if err != nil {
    return err
}
frame.Boxes[0].Child = nav
```

`OnSelection` runs when the current entry changes, including after directory loading. `OnActivate` handles confirmed non-directory entries. `OnOpen` runs after a directory is opened, and `OnQuit` handles ESC at the navigation root. Hosts can inspect `Directory()` and `Selected()`, style the list through `List()`, and call `SetRoot("")` when ESC should navigate to the filesystem parent without requesting a quit.

`ConsumeKey` reports whether ESC or backspace belongs to the navigation contract. Hosted widgets can forward these keys to `ConsumeKey` before a frame handles other keys, preserving parent navigation and root quit behavior across focus boundaries. Other keys should pass through the normal frame routing so filtering, selection, and frame actions continue to work.

The pane retains one `Choice` instance while opening directories: `Choice.SetItems` replaces the rows and resets its filter, while `SelectIndex` restores a requested row. Hosts that customize list behavior can use `List()` without replacing the pane's child widget; the frame and pane then keep a consistent child identity.

The pane draws its Choice inside the rectangle supplied by its host. `Choice.HandleMouse` expects coordinates relative to that drawn rectangle, while `NavigationPane.HandleMouse` receives child-local coordinates and bridges them to the last Choice draw rectangle. Containers already translate mouse events to child-local coordinates, so callers should pass those events directly and must not subtract an additional cell.

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

## 7. Root Event Loop Contract: Quit vs Consumption Invariants

In Loom's application event loop contract:
- Returning `true` from a root widget's `HandleKey` or `HandleMouse` signals a **request to quit the application event loop**, not merely that the event was consumed.
- Child widgets and application containers (`TextEditApp`, custom layouts) must return `false` after handling ordinary keystrokes, navigation, or mouse clicks.
- Returning `true` upon handling a keystroke (such as typing a character or clicking a pane) will cause the application to immediately terminate.
- Reserved exit keys (`F10`, `ctrl-q`) should be intercepted at the application root and explicitly return `true` (or delegate to `Frame.HandleKey(e)` with a quit action), ensuring global quit capability across all focused children.

## 8. Mouse Coordinate Invariants

Mouse coordinates: events reaching widgets are 0-based, PTY SGR mouse reports are 1-based.
Containers (`Split`, `Stack`, `Pane`, …) already translate events to the child's local
0-based cells (059). Never subtract 1 again in widget code: loomoji's grid did (`e.Y-1`)
and highlighted the row above the pointer (108). Verify hit-testing black-box with a
[hover probe](HoverTesting.md).
In tests locate screen text by runes or display width, never byte offsets.

