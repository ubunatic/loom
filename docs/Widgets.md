---
title: Loom Widgets and Framework Primitives
weight: 35
---

# Loom Widgets and Framework Primitives

Loom provides composable, dependency-light widget primitives designed for inline terminal UIs.
Widgets implement `loom.Widget` (`Draw`, `ConsumeKey`, `ConsumeMouse`) and integrate cleanly with `loom.Frame`, `loom.Pane`, and the layered compositor.

### Wrapping widgets

A wrapper can implement `Unwrap() loom.Widget` to expose the widget it decorates to
the framework. `Pane` follows unwrap chains when it discovers optional lifecycle
hooks, including ticking, invalidation, and tick reset callbacks. This lets a
presentation wrapper focus on drawing and input without forwarding each optional
interface. `loom.UnwrapWidget` returns the innermost widget when a host needs to
look up another optional interface, such as `loom.Themeable`.

```go
type bordered struct{ child loom.Widget }

func (b *bordered) Unwrap() loom.Widget { return b.child }
```

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

### Axes Chart (`loom.Chart`)
`Chart` plots ordered numeric samples as braille line series or grouped block bars. It scales automatically to the data, or uses `Min`, `Max`, and `RangeSet` for a fixed vertical range. The Y axis uses five evenly spaced numeric labels, X ticks mark sample positions, and a legend appears only when there are multiple series.

```go
chart := &loom.Chart{Series: []loom.ChartSeries{
    {Name: "Requests", Values: []float64{12, 18, 14, 26}},
    {Name: "Errors", Values: []float64{2, 4, 3, 6}},
}}
chart.Mode = loom.ChartGroupedBar // default is ChartLine
```

---

## 4. Filesystem & OS Primitives (`loom.Directory` and `OpenFile`)

Loom extracts common terminal file-browsing and launch operations into clean, portable helpers:

- **Structured Directory Scanner (`ReadDirectory`)**: Reads directory entries, detects symlinks/directories, sorts entries with `..` parent navigation, and preserves path mapping.
- **Terminal-Safe Path Escaping (`DisplayPath`, `QuoteUnprintable`)**: Escapes unprintable control codes and formats paths safely for terminal display columns without corrupting layouts.
- **Injectable Platform File Opener (`OpenFile`, `FileOpener`)**: Dispatches file launch commands (`xdg-open`, `gio`, `open`, `start`) detached from the terminal process, with mockable injection for deterministic unit tests.

### Reusable File Picker (`loom.FilePicker`)
`FilePicker` browses directories using `ReadDirectory`, filters files with `filepath.Match` patterns, and calls `OnSelect` with the chosen path. File mode navigates directories and selects regular files; directory mode selects directories and filters out files. Esc calls `OnCancel`; backspace navigates to the parent directory when the filter is empty.

```go
picker, err := loom.NewFilePicker(".", loom.FilePickerOptions{
    Mode: loom.FilePickerFiles,
    Patterns: []string{"*.go", "*.md"},
    OnSelect: func(path string) { fmt.Println("Selected", path) },
    OnCancel: func() { fmt.Println("Cancelled") },
})
if err != nil { return err }
frame.Boxes[0].Child = picker
```

Use `List()` to customize the embedded `Choice`, `Directory()` to inspect the current directory, and `ApplyTheme` to apply host theme colors. The widget shows the current path above its list and supports mouse selection as well as keyboard navigation.

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

`NavigationPane` consumes ESC and backspace when they navigate to a parent or
request a root quit. Other keys follow normal event bubbling: the focused pane
may consume them, and its containing frame handles only the keys left unconsumed.

The pane retains one `Choice` instance while opening directories: `Choice.SetItems` replaces the rows and resets its filter, while `SelectIndex` restores a requested row. Hosts that customize list behavior can use `List()` without replacing the pane's child widget; the frame and pane then keep a consistent child identity.

The pane draws its Choice inside the rectangle supplied by its host. Both widgets receive 0-based, child-local mouse coordinates; `NavigationPane` forwards the result from `Choice` directly. Containers translate events to child-local coordinates, so callers should pass those events directly and must not subtract an additional cell.

---

## 5. Forms (`loom.Form`)

`Form` lays out existing input widgets with aligned labels, required markers,
help text, keyboard focus traversal, a validation summary, and submit/cancel
callbacks. Add `FormAction.OnActivate` for custom footer actions.

```go
name := loom.NewTextInput("")
form := loom.NewForm([]loom.FormField{{Label: "Name", Widget: name, Required: true}})
form.Actions = []loom.FormAction{{Label: "Save"}, {Label: "Cancel", Cancel: true}}
form.OnSubmit = func(values map[string]any) { fmt.Println(values["Name"]) }
```

Tab/Down and Shift-Tab/Up move between fields and actions. Enter validates
fields, focuses the first error, and submits valid values. Esc invokes
`OnCancel`; validation messages render together above the fields.

## 6. Startup & Splash Transition Runner (`Pane.RunStartup`)

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

## 7. Primitives Added by the 2026-09 Lean Sprints

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

## Menus (`loom.MenuBar`, `loom.Menu`)

`MenuBar` lays out titled menus in one row and draws the active menu as a bordered dropdown beneath its title. Dropdown items can show shortcuts, disabled state, separators represented by a `Label` of `---`, and check marks through a `*bool` `Checked` field.

```go
bar := loom.NewMenuBar(loom.Menu{
    Title: "File", Mnemonic: 'F', Items: []loom.MenuItem{
        {Label: "Open", Shortcut: "Ctrl+O", Action: openFile},
        {Label: "Autosave", Checked: &autosave},
    },
})
```

F10 opens the bar; `Alt+<mnemonic>` opens a matching menu. Left/Right moves between titles, Up/Down moves through items, Enter/Space invokes the selected action, Escape closes the dropdown, and a second Escape unfocuses the bar. Mouse hover changes the active title or highlighted item; a title click toggles its dropdown and an outside click dismisses it. Item shortcut labels are matched through `KeyMap`. `Submenu` is reserved for nested menus and is not opened by this one-level widget.

## 8. Root Event Loop Contract: EventResult and Quit Invariants

Loom uses the small value struct `loom.EventResult` to cleanly distinguish whether an event was consumed from whether the application should terminate:

```go
type EventResult struct {
    Consumed bool
    Done     bool // the widget's interaction completed; the host keeps running
    Quit     bool
}
```

Constructors and helpers (always returned by value):
- `loom.Handled()` / `loom.Consumed()`: `EventResult{Consumed: true, Quit: false}`
- `loom.Ignored()` / `loom.Unhandled()`: `EventResult{Consumed: false, Quit: false}`
- `loom.Quit()` / `loom.QuitResult()`: `EventResult{Consumed: true, Quit: true}`
- `loom.DoneResult()`: `EventResult{Consumed: true, Done: true}` (issue 223)

### Event Handling
`Widget` has one input contract: `ConsumeKey(e KeyEvent) EventResult` and
`ConsumeMouse(e MouseEvent) EventResult`. Key events bubble: each container
offers the key to its focused child first and handles it only when the child
returns unconsumed. A focused widget consumes an arrow only when it can move in
that direction, so a parent such as `Grid` can use the key at the child's edge.
Composite widgets preserve both `Consumed` and `Quit` while forwarding results.
`DispatchKeyEvent` and `DispatchMouseEvent` call the corresponding widget method
directly. `Pane` applies fallback quit keys only when a key result is unconsumed.
Navigation keys (`arrows`, `home`, `end`, `pgup`, `pgdn`, `delete`, `tab`,
`backspace`) never trigger fallback quit.

Before v0.2.15 widgets returned `bool`, where `true` meant quit; see
[Upgrading](Upgrading.md) for the mapping to `EventResult`.

### Widget Implementation Contract
- Widgets handling user input (e.g. navigation, typing, selection) return `loom.Handled()` when they consume an input without quitting and `loom.Ignored()` when they did not handle it.
- Explicit quit shortcuts (such as `F10` and `Ctrl-Q`) return `loom.QuitResult()`.
- Confirming a widget (Enter or click on a Choice item, a Table row, a Confirm button) returns `loom.DoneResult()`, never Quit: an embedded widget must not end its host app. Standalone prompt runners in `driver.go` map Done to exit; hosts never mask a child's `Quit` (issues 212, 223). Cancel keys (Esc, Ctrl-C) may still quit a standalone prompt.
- A host that draws its own rows (footer, status bar) passes children only the remaining rect; drawing a child over those rows leaks the child's styles, e.g. the selection colour after a theme switch (issue 220).
- `Pane` enables bracketed paste for the run and restores the terminal mode on exit. Widgets that implement `PasteConsumer` receive one `PasteEvent` per paste; other widgets ignore it. `TextInput` replaces pasted line breaks with spaces, while `TextArea` preserves them.
- `TextInput.Mask` optionally replaces each displayed value rune with the configured rune (for example, `•`) and adjusts the caret to the mask's display width. A zero mask keeps normal text display. `Value()` continues to return the entered text to the host. `TextInput` and `TextArea` can copy via an optional `Keys` `KeyMap` binding named `copy`; `SetSelection(start, end)` selects a half-open rune range, and an empty selection copies the whole value. `Pane` sends the value through OSC 52.

## 8. Mouse Coordinate Invariants

Mouse coordinates: events reaching widgets are 0-based, PTY SGR mouse reports are 1-based.
Containers (`Split`, `Stack`, `Pane`, …) already translate events to the child's local
0-based cells (059). Never subtract 1 again in widget code: loomoji's grid did (`e.Y-1`)
and highlighted the row above the pointer (108). Verify hit-testing black-box with a
[hover probe](HoverTesting.md).
In tests locate screen text by runes or display width, never byte offsets.

## Rich Text Editing (`loom.RichDocument`, `loom.RichTextEdit`)

`RichDocument` stores logical lines as styled spans. `RichTextEdit` renders those
spans into a clipped, scrollable viewport and supports rune-based cursor
navigation, shift-arrow selection, mouse drag selection, and inline bold,
italic, and underline shortcuts (`Ctrl+B`, `Ctrl+I`, `Ctrl+U`). A non-empty
selection shows a floating formatting popover by default; set `ShowPopover` to
false to hide it. Its B/I/U/S buttons toggle styles on the selection. Click
`#FG` or `#BG` to open a 16-color ANSI swatch row and apply a color across the
selected spans. Mouse coordinates are child-local, so the editor can handle
selection drags and popover clicks directly. `RichSpan` can also carry link,
code, or `RichPill` metadata; pill spans are atomic when deleted. The widget
gallery demonstrates selection formatting and color picking alongside mention
pills.

Editing shortcuts: `Ctrl+B/I/U` style the selection, or the word under the cursor
when nothing is selected (no word: only the typing style changes). `Ctrl+Space`
selects the word and shows the popover. The clipboard is internal and keeps
styled spans: `Ctrl+C`/`Ctrl+Insert` copy the selection or word, `Ctrl+X`/
`Shift+Delete` cut, `Ctrl+V`/`Shift+Insert` paste. `Ctrl+Z`/`Ctrl+Y` undo;
`Ctrl+R`, `Ctrl+Shift+Y`, `Ctrl+Shift+Z` redo (100 steps; a typing run up to a
word boundary, a style change, a cut or a paste is one step). Double-click
selects a word and triple-click a line. `Ctrl+I` arrives only from terminals
that speak the kitty/CSI-u protocol; on legacy terminals byte 0x09 stays Tab,
and `Ctrl+Shift+Y/Z` likewise need CSI-u.

```go
doc := &loom.RichDocument{Lines: []loom.RichLine{{Spans: []loom.RichSpan{
    {Text: "Hello ", Style: loom.Style{Bold: true}},
    {Text: "@ada", PillData: &loom.RichPill{Kind: "mention", ID: "ada"}},
}}}}
editor := loom.NewRichTextEdit(doc)
frame.Boxes[0].Child = editor
```

## 9. Syntax Highlighting and Navigation Engine (`syntax/`, `TextArea`)

Loom provides a UI-neutral syntax highlighting and structural navigation framework (`ubunatic.com/loom/syntax`):

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

### Controls (112)
A two-row footer (key hints, then `▶ Play [-] [+]`, zoom level and an activity ProgressBar) shrinks or disappears on short widgets. Space toggles playback; `+`/`-`/wheel zoom by 1.25× within 1×–8×; arrows pan 12%; drag pans; `0` resets. Pan clamps to the media bounds. The ProgressBar shows activity only: the video API exposes no duration.





## 13. Library Widget Catalog

Every catalog widget has a live demo in the `gallery/` package: `loom widgets --show <Name>...` (195). `spec/widgets.yaml` is the catalog source of truth, kept in strict name order; `widgets_catalog_test.go` fails when an exported widget is missing from it.

`PaintCanvas` draws braille dots with the left mouse button. Press to start a stroke,
drag to connect points, and press `c` or Ctrl-L to clear the drawing.

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

## 14. Date Picker (`loom.DatePicker`)

`DatePicker` binds a date-only `*time.Time`, with optional inclusive `Min` and
`Max` bounds. Its month grid starts on Monday by default; set `WeekStart` to a
different `time.Weekday` to change the first column. Left/Right moves one day,
Up/Down moves one week, PgUp/PgDn moves one month, and Ctrl+Left/Ctrl+Right
moves one year. Enter selects the cursor date. Type `YYYY-MM-DD` and press
Enter to select a date directly. The current day is bold with `*`; the selected
day is marked with `>`, and the cursor is underlined. The injected `Now`
function controls the current-day marker; `NewDatePickerWithClock` also makes an
empty bound value initialize deterministically.

```go
selected := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)
picker := loom.NewDatePicker(&selected)
picker.Min = time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
picker.OnSelect = func(date time.Time) { fmt.Println(date.Format("2006-01-02")) }
```

## 15. Search Bar (`loom.SearchBar`)

`SearchBar` provides an input line with prompt glyph (`> `), typed query tracking, placeholder hint support, cursor management, match count / status label controls (`Controls`), and command mode (`:` / `/`).

```go
sb := loom.NewSearchBar()
sb.Prompt = "> "
sb.Placeholder = "type to filter"
sb.Controls = "4/12"
sb.OnChange = func(query string) {
    // filter items
}
sb.OnSubmit = func(query string) {
    // confirm selection
}
```

### Key Capabilities & Invariants
- **Standalone & Embeddable**: `SearchBar` implements `loom.Widget` (`Draw`, `ConsumeKey`, `ConsumeMouse`). Search-capable container widgets like `Choice` and `Table` embed `SearchBar` directly and expose it via `.SearchBar()`.
- **Command Mode Integration**: Supports `AddCmd(loom.Cmd)` and automatically enters command mode on `:` or `/` keypresses when commands are registered.
- **Surface & Theme Blending**: Uses `SearchBarStyle` (`Container`, `Prompt`, `Query`, `Placeholder`, `Controls`) and paints surfaces seamlessly over parent grid/widget focus surfaces.

---

## 16. Focus Management, Editor Activation & Cursor Routing

Loom handles keyboard focus, inline editing lifecycles, and terminal cursor routing across arbitrarily nested container hierarchies (`Grid`, `Tabs`, `Split`, `Stack`, `Form`, `Frame`):

### The `Focusable` Contract
Widgets that maintain interactive or selectable states implement `loom.Focusable`:
```go
type Focusable interface {
	Widget
	Focused() bool
	SetFocus(bool)
}
```

- **Focus Propagation**: Composite containers forward focus to their active child during traversal (e.g. `Grid.setFocus`, `Tabs.Select`, `Split.setFocusedChild`).
- **Inline Editor Teardown on Focus Loss**: When focus leaves an editing widget (e.g. `NumberInput`), `SetFocus(false)` automatically cancels/ends active inline editing (`endEdit()`). This prevents orphaned input editors, stale cursors, and accidental key interception.
- **Embedded vs Standalone State**: Embedded editor primitives like `TextInput` and `TextArea` take an explicit `focused bool` parameter in `Draw(c *Canvas, r Rect, focused bool)`, delegating focus lifecycle ownership to their host widget.

### Terminal Cursor Placement & Sub-Canvas Routing
- **Focused Placement**: Widgets set `Canvas.CursorX` and `Canvas.CursorY` during `Draw` only when active/focused. Unfocused widgets and static views never touch canvas cursor coordinates.
- **Compositor Propagation (`SubCanvas` & `Blit`)**: `SubCanvas` and `Blit` preserve and translate cursor positions relative to their destination bounds, ensuring cursor coordinates accurately reach the terminal when rendering through `Viewport`, `Split`, `Frame` (`paintClipped`), or nested layers.
- **Hardware Cursor Flush**: `Canvas.FlushWithConfig` emits `\x1b[?25h` with terminal row/col addressing only when `CursorX >= 0 && CursorY >= 0`, and emits `\x1b[?25l` (hidden cursor) otherwise.
