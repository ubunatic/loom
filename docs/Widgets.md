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
| 2D Panning in `loom.View` (`OffsetX`, `OffsetY`, `Pan`) | 138 | 2D offset panning with ANSI-aware style-preserving horizontal column clipping. |
| Lazy Media Loading (`media.Widget`) | 140 | 50 ms threshold non-blocking background render with dim loading indicator. |
| `loom.ProgressBar` (`Set`, `Done`, `Reset`), `graph.BracketedBarStep` | 167 | Bracketed progress bar; set `Total`, `ShowPercent`, and `ShowCount` for labels. `Indeterminate` enables a bouncing pulse driven by the pane ticker. `StyleFill` and `StyleEmpty` style cells separately, and `ApplyTheme` supplies theme colors. `Set` invalidates only when a visible half-cell changes. `Done` swaps to the specced `:` pattern like the splash. |
| `loom.Spinner` (`Start`, `Stop`) | 186 | Labeled braille activity indicator driven by the pane ticker; stopped spinners return a zero cadence and stop requesting ticks. |
| `loom.Paginator` (`Page`, `Pages`, `OnChange`) | 188 | Zero-based page navigation with clickable dots or a numeric indicator; the callback lets a host update a Choice or Table offset. |
| `loom.Viewport` (`ScrollX`, `ScrollY`) | 190 | Scroll any child that reports its content size through `Measurer`, `ContentWidther`, or `ContentHeighter`; arrow keys, PgUp/PgDn, Home/End, and the mouse wheel move the visible window. |
| `loom.Tree` (`TreeNode`, `OnActivate`) | 168 | Navigate nested nodes with arrows or `hjkl`, expand/collapse branches, select rows with the mouse, and scroll long visible trees. |
| `loom.Timer` and `loom.Stopwatch` (`Start`, `Stop`, `Reset`) | 187 | Countdown and elapsed time widgets tick once per displayed second. Inject `Now` for deterministic clocks, set `Formatter` for custom text, and use `Timer.OnDone` for one-shot completion. |

## 7. Root Event Loop Contract: EventResult, EventConsumer, and Quit Invariants

Loom uses the small value struct `loom.EventResult` to cleanly distinguish whether an event was consumed from whether the application should terminate:

```go
type EventResult struct {
    Consumed bool
    Quit     bool
}
```

Constructors and helpers (always returned by value):
- `loom.Handled()` / `loom.Consumed()`: `EventResult{Consumed: true, Quit: false}`
- `loom.Ignored()` / `loom.Unhandled()`: `EventResult{Consumed: false, Quit: false}`
- `loom.Quit()` / `loom.QuitResult()`: `EventResult{Consumed: true, Quit: true}`

### Event Handling Hierarchy
When an event arrives at `Pane` (or composite containers `Frame`, `Tabs`, `Stack`, `Grid`, `Split`), `DispatchKeyEvent` and `DispatchMouseEvent` resolve the event in precedence order:
1. `EventConsumer` (`ConsumeKey(e KeyEvent) EventResult`) / `MouseConsumer` (`ConsumeMouse(e MouseEvent) EventResult`).
2. Legacy `KeyConsumer` (`ConsumeKey(e KeyEvent) (quit, consumed bool)`).
3. Historical `Widget.HandleKey(e KeyEvent) bool` / `Widget.HandleMouse(e MouseEvent) bool`, where returning `true` signals a request to **quit the application**.
4. Unhandled fallback quit keys (e.g. `Ctrl-C`, `Ctrl-Q`, `Esc`, `q` for non-text widgets). Navigation keys (`arrows`, `home`, `end`, `pgup`, `pgdn`, `delete`, `tab`, `backspace`) never trigger fallback quit.

### Widget Implementation Contract
- Widgets handling user input (e.g. navigation, typing, selection) should implement `EventConsumer` with `ConsumeKey(e KeyEvent) EventResult` and return `loom.Handled()` on consumed inputs.
- If implementing historical `HandleKey(e KeyEvent) bool`, return `false` on ordinary keystrokes and `true` ONLY when requesting application termination.
- Explicit quit shortcuts (such as `F10` and `Ctrl-Q`) return `loom.QuitResult()`.
- `Pane` enables bracketed paste for the run and restores the terminal mode on exit. Widgets that implement `PasteConsumer` receive one `PasteEvent` per paste; other widgets ignore it. `TextInput` replaces pasted line breaks with spaces, while `TextArea` preserves them.
- `TextInput.Mask` optionally replaces each displayed value rune with the configured rune (for example, `•`) and adjusts the caret to the mask's display width. A zero mask keeps normal text display. `Value()` continues to return the entered text to the host; the masked display does not reveal it, and TextInput has no copy action.

## 8. Mouse Coordinate Invariants

Mouse coordinates: events reaching widgets are 0-based, PTY SGR mouse reports are 1-based.
Containers (`Split`, `Stack`, `Pane`, …) already translate events to the child's local
0-based cells (059). Never subtract 1 again in widget code: loomoji's grid did (`e.Y-1`)
and highlighted the row above the pointer (108). Verify hit-testing black-box with a
[hover probe](HoverTesting.md).
In tests locate screen text by runes or display width, never byte offsets.

## 9. Syntax Highlighting and Navigation Engine (`syntax/`, `TextArea`)

Loom provides a UI-neutral syntax highlighting and structural navigation framework (`codeberg.org/ubunatic/loom/syntax`):

### Core Architecture & Separation of Concerns
- **UI-Neutral Coordinates**: `syntax.Point` (0-based line and column), `syntax.Edit` (delta ranges and replacement text), and `syntax.Span` (named token captures like `"keyword"`, `"string"`, `"comment"`). The syntax engine has zero terminal or UI dependencies.
- **Engine Contract (`syntax.Engine`)**:
  - `SetSource(src []byte, path string)`: Full buffer initialization.
  - `ApplyEdit(e Edit)`: Incremental synchronization for responsive typing.
  - `HighlightViewport(startLine, endLine int) []Span`: Viewport-bounded token queries preventing full-buffer highlighting overhead.
- **Theme & Styling (`syntax.ThemeMap`, `syntax.StyleResolver`)**: Maps capture IDs to `loom.Style` (`Fg`, `Bg`, `Attrs`).
- **Single-Pass Viewport Rendering in `TextArea`**: `loom.TextArea` queries `HighlightViewport` during `Draw(rect, canvas)` and directly formats canvas cells without intermediate multi-pass color overlays.

### Structural Code Navigation & Folding (`syntax.Navigator`, `syntax.OutlineProvider`)
- **Symbol Outline & Breadcrumbs**: `syntax.Symbol` represents hierarchical code elements (`KindPackage`, `KindFunction`, `KindType`, `KindMethod`, etc.).
- **Scope Detection (`ScopeAt(point)`)**: Detects enclosing function/type hierarchy for real-time status bar breadcrumbs (`pkg > Type > Method()`).
- **Code Folding**: `FoldRanges()` computes foldable block boundaries; `TextArea.SetFolded(line, true)` hides collapsed lines while maintaining accurate cursor and line gutter mapping.

### Engines Available
- **`syntax.LexicalEngine`**: Fast, pure-Go regexp/scanner engine with syntax rules for Go, Markdown, JSON, and YAML.
- **`syntax.NullEngine`**: Zero-allocation no-op engine for plain text.
- **Wasm Tree-Sitter Integration**: Extensible through WebAssembly (`wazero`) runtime bindings.

---

## 10. ANSI Graphic Cell Buffer and Editor (`loom.AnsiBuffer`, `loom.AnsiEditor`)

Loom provides 2D spatial ANSI cell grid modeling and interactive overtype/insert graphic editing:

- **`loom.AnsiBuffer`**: An in-memory 2D styled cell grid (`AnsiCell` storing `Rune`, `FG`, `BG`, `Bold`, `Dim`, `Underline`, `Invert`):
  - **SGR/CSI Parser & Serializer**: `ParseAnsiBuffer`, `LoadAnsiBuffer`, `SerializeAnsiBuffer`, and `SaveAnsiBuffer` for lossless ANSI art roundtrips.
  - **2D Editing Primitives**: `Put`, `PutChar` (overtype & insert modes), `Delete`, `Backspace`, `Copy`, `Cut`, `Paste`, `NextWord`, `PrevWord`, `NextObjectRow`, `PrevObjectRow`.
- **`loom.AnsiEditor`**: Standard widget implementing `loom.Widget`, `loom.EventConsumer`, `loom.MouseConsumer`, `loom.Focusable`, and `loom.PaneRequester`:
  - Interactive arrow/word/object navigation and character painting.
  - Automatic viewport panning and focused cursor styling.
  - Mouse click-to-focus/navigate and wheel scrolling.

---

## 11. 2D Scrollable Viewport (`loom.View`)

`loom.View` renders lines of plain or ANSI-styled text into a constrained rectangle with full 2D offset controls:

```go
view := loom.NewView(lines)
view.SetOffset(colOffset, lineOffset)
view.Pan(dx, dy) // moves offset with non-negative clamping
```

### Key Capabilities & Invariants
- **2D Offset & Panning**: `OffsetX` sets the first visible terminal display column; `OffsetY` (alias for `Scroll`) sets the first visible line index.
- **ANSI-Aware Style-Preserving Slicing**: When lines contain SGR escape codes (e.g. truecolor or 256-color text), horizontal column offsets cleanly clip preceding characters without dropping active color styles or breaking multi-byte/wide Unicode runes (`drawViewLine`).
- **Interactive Navigation**: Supports two-axis keyboard panning (`left`/`right`/`up`/`down`, `h`/`l`/`k`/`j`, `pgup`/`pgdn`, `home`/`end`) and mouse wheel / trackbar dragging.

---

## 12. Media and Image Widget (`media.Widget`)

`media.Widget` renders static images and streaming video preview frames into halfblock/terminal canvas grids:

```go
imgWidget, err := media.LoadImage("asset.png", media.ModeHalfblock)
videoWidget, err := media.NewVideo("preview.mp4", media.ModeHalfblock, 24)
```

### Lazy Background Loading
- **50 ms Threshold (`renderWithThreshold`)**: Media decodes and renders exceeding 50 ms are automatically backgrounded asynchronously.
- **Non-Blocking UI Loop**: The widget immediately displays a dim `"loading"` indicator during slow operations and populates the rendered grid into the cache upon completion, keeping event loops and parent layouts fully responsive.





## 13. Library Widget Catalog

### Standalone Form Controls

`NumberInput` and `Toggle` can be used outside a `Settings` list. Both bind to
the caller's value through a pointer:

```go
size := 5.0
number := loom.NewNumberInput(&size, 0, 10)
number.Step = 0.5

enabled := false
toggle := loom.NewToggle(&enabled)
```

`NumberInput` steps with Left/Right and edits with Enter; Enter commits a valid
value and Esc cancels. `Toggle` changes on Enter or Space. `Settings` uses the
same control implementations for its boolean and bounded-number rows.

Run `loom widgets` for the complete library widget catalog, short examples,
categories, and source references. Use `loom widgets <name>` to show one entry.
The catalog is maintained in `spec/widgets.yaml`; its completeness test checks
exported `loom.Widget` implementations in library packages.

The test also counts embedded controls such as `TextInput` and `TextArea`,
which draw with a focus flag and are hosted by a parent widget.

Check the catalog before telling a consumer an input widget is missing:
`Settings` also offers toggle (`KindBool`), text (`KindString`), choice
(`KindChoice`), and bounded-number (`KindNumber`) rows. The file browser `NavigationPane` lives in
`examples/filebrowser` and is not catalogued until a library `FilePicker`
exists (issue 172). Forms, dates, and menus are issues 169, 175, and 170.
