---
title: Loom Widgets and Framework Primitives
weight: 35
---

# Loom Widgets and Framework Primitives

Loom provides composable, dependency-light widget primitives designed for inline terminal UIs.
Widgets implement `loom.Widget` (`Draw`, `ConsumeKey`, `ConsumeMouse`) and integrate cleanly with `loom.Frame`, `loom.Pane`, and the layered compositor.

The widget that places the visible cursor can set `Canvas.CursorShape`; clipped
containers carry that request with the cursor position. `Pane` applies DECSCUSR
bar/block changes only when the requested shape changes and restores the
terminal default when it closes. `RichTextEdit` requests a bar cursor while
typing and a block cursor in box drawing mode. Its optional
`GhostCursorEnabled` setting defaults to `false` in `spec/defaults.yaml`; when
enabled, navigation can move into empty columns and rows without changing the
document until typing or drawing materializes that space.

`RichTextEdit.Highlights` accepts display-only `RichTextHighlight` ranges in document rune coordinates. Highlight styles overlay styled spans without changing editing selection, formatting popovers, dirty state or serialized document content. A highlight with `Reveal: true` is kept in view while the editor is unfocused, so a host can show an off-screen search match without moving the cursor; focus returns the view to the cursor. Highlights are display-only and never change selection, cursor, popovers, dirty state or content. Assign an empty/nil slice to clear highlights; callers keep ranges current after edits.

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

### Grid layout (`loom.Grid`)

`Grid` divides its width into `Cols` columns. Rows share available height by
default. Set `FitRows = true` to give each row the largest preferred height
reported by its children through `Measurer.Measure(cellWidth)`. Heights are at
least one terminal line; widgets without `Measurer` receive a one-line fallback. `Dialog` measures its border, body and buttons; `Chart` asks for a five-line plot plus axis, labels and, with several series, a legend line. A wrapper widget must forward `Measure`, because Grid does not look through `Unwrap`.
Spare height stays below the fitted rows. Overflow is clipped to the Grid area;
there is no scrolling, and keyboard navigation can focus an offscreen child.

`ChildRect(i)` returns the full cell bounds from the most recent draw, including
clipped portions. Mouse routing uses those bounds and passes 0-based child-local
coordinates only for points inside the visible Grid area. Arrow keys move focus
between columns and rows; Shift-arrow keys bypass the child's key handler.
`GridBorderInner` adds one-cell separators, and `GridBorderFull` also borders the
assigned area's outside edge, including any spare height. Border cells do not
route mouse events to children. The gallery **All** demo enables fitted rows and
inner borders (`loom widgets --show All`).

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

`FilePickerSave` adds a filename field and calls `OnSave` (or `OnSelect`) with
the destination path. A divider separates the directory search/list from the
filename field. Tab and Shift+Tab switch focus and cursor between the filename
field and directory search; set `FocusFileName` to start with the filename field
active. An `OnSave` error leaves the picker open for correction and retry.

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
- **HTML Screenshots (`loom.RenderHTML`, `loom.WriteHTML`, `loom.WriteHTMLRows`)**: Writes a frame (or ANSI rows from `loom.Render`) as one self-contained HTML document you can open offline in a browser. The output is a single `<pre>` monospace grid with inline styles only: no external assets and no scripts. Each run of equal-style cells is one `<span>` with its foreground, background (256-colour and 16-colour values resolved to hex), bold, italic and underline. Terminal-default colours use fixed values (`#d4d4d4` on `#1e1e1e`). Dim and strikethrough are not rendered.

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
| `Pane.Post(fn)` | 320 | Run `fn` on the event loop from any goroutine, then repaint. Use `Post` to change widget state from a producer (streams, job logs); `Invalidate` only requests a repaint and must not be paired with off-loop mutation. Never blocks or drops; FIFO; a burst draws once; no-op after `Close`. |
| `Pane.Suspend(fn)` | 319 | Hand the terminal to `fn` (e.g. run `$EDITOR` on `/dev/tty`) from the event loop: the input reader stops without losing bytes, alt screen/mouse/paste/raw mode are left, then restored with a fresh size and full repaint; restores on error or panic; if `fn` calls `Close`, Suspend does not resume and `Run` returns. `fn` must not open a nested Pane: the single-pane guard still holds. See `examples/suspend`. |
| `Choice.MouseTextOnly` | 093 | Hover and drag track only rendered item text; a left press anywhere on a row selects it (296), activation follows `DoubleClickToActivate`/`SelectOnlyOnClick`. |
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

When the host permits menu routing, F10 opens the bar and `Alt+<mnemonic>` opens a matching menu. A `Pane` handles F10 as global Quit by default. Left/Right moves between titles, Up/Down moves through items, Enter/Space invokes the selected action, Escape closes the dropdown, and a second Escape unfocuses the bar. Mouse hover changes the active title or highlighted item; a title click toggles its dropdown and an outside click dismisses it. Item shortcut labels are matched through `KeyMap`. `Submenu` is reserved for nested menus and is not opened by this one-level widget.

## Box Title Segments (`loom.Box.TitleSegments`, `loom.HotkeyTitle`)

`Box.TitleSegments` is an ordered list of `loom.TitleSegment{Text, Style}` runs drawn in place of `Box.Title`. Each run keeps its own style, so one title can hold a hotkey letter, a bold key badge and dim metadata. `Border.TitlePrefix` and `Border.TitleSuffix` still frame the runs in `Style.Title`. Clipping is by display width across run boundaries: a wide rune or grapheme cluster that does not fit is dropped whole.

`loom.HotkeyTitle(title, key, hotkey, text)` builds the runs for a hotkey. If `key` matches a letter of the title's first word (case-insensitive), that letter is drawn in `hotkey`, e.g. `menu` with `m` gives `m` + `enu`. Otherwise it returns a key badge `[k]` in `hotkey` followed by the full title in `text`, e.g. `[d] Fetch diagnostics`. The helper is render-only and adds no key routing.

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

Tab policy: Tab belongs to containers (Split, Form, FilePicker, the gallery's
demo switch), which use it to move focus. A widget must not consume Tab unless
Tab is its documented action: Dialog moves between buttons with left/right only,
while Table uses Tab for sort and shows it as `[tab:…]` (open question in issue 277).
`ColorPicker` leaves Tab to its containing Frame, Split, Tabs or other focus
container. Up and Down move between its foreground and background rows.

Find and paging policy (issue 407): Ctrl+F always means find and is never
bound to paging. PgUp/PgDn scroll by a page, and so do Shift+PgUp/PgDn in
read-only views (in text input they extend the selection by a page); Home/End scroll to the top and
bottom unless a text input is focused, where they move the caret. Ctrl+B is
the `loom session` prefix and is not bound to paging either. `loom.View` has
no find yet, so it leaves Ctrl+F unconsumed for its host.

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
navigation, shift-arrow selection, `Home`/`Ctrl+A` and `End`/`Ctrl+E` (line
start/end of the logical line; the editor does not soft-wrap), their Shift
variants (`Shift+Home`, `Shift+End`, `Ctrl+Shift+A`, `Ctrl+Shift+E`, the last two
on CSI-u terminals) that extend the selection, Ctrl+Shift+Left/Right word
selection, Ctrl+Home/End document navigation and Ctrl+Shift+Home/End selection,
Shift+PgUp/PgDn page selection, and Select All (`Ctrl+Alt+A`, `Alt+A`, or
`Super+A`; Ctrl+A remains line start). Page movement uses the visible editor
height. Alt bindings on macOS require Option-as-Meta in the terminal; Super+A
requires the terminal to forward the Super/Cmd key. Mouse drag selection and inline bold,
italic, and underline shortcuts (`Ctrl+B`, `Ctrl+I`, `Ctrl+U`). A non-empty
selection shows a floating formatting popover by default; set `ShowPopover` to
false to hide it. Its B/I/U/S buttons toggle styles on the selection. The bar
uses `▏` separators and offers 🎨, Frame, Link and the ≡ submenu. The b/B, i/I,
u/U and s/S hotkeys toggle formatting; f/F opens Frame. L/l activates the
disabled Link entry without changing the document. The popover captures mouse
input above any host overlay, so a host that paints its own overlay over the
editor (such as a search panel) sets `DeferPopover`: `Draw` then skips the
popover and the host calls `DrawPopover(canvas, rect)` with the same rectangle
after its overlays. Capture, hit areas and keys are unchanged, and the default
`Draw` paints the popover itself; `DrawPopover` does nothing for a default
editor and while the help or save popup is open. Ctrl+Space opens the bar
focused on its first item; Up/Down enters an auto-shown bar before its letters
act as hotkeys. Letters remain normal text while the selection popover is
unfocused. The 🎨
button opens the reusable `ColorPicker` with the current FG/BG pair and a
preview; swatches update the selected spans or the typing style when there is no
selection. Mouse coordinates are child-local, so the editor can handle
selection drags and popover clicks directly. `RichSpan` can also carry link,
code, or `RichPill` metadata; pill spans are atomic when deleted. The widget
gallery demonstrates selection formatting and color picking alongside mention
pills. Popover submenus open on the side away from the selection when space
allows; constrained layouts may clamp them only when selected cells remain
visible, otherwise the submenu stays closed. The formatting popover's Box
dropdown wraps a single- or multi-line selection in a plain (sharp) or rounded
box, defaulting to plain; rows are padded
to the widest display width and the selection's inline styles and metadata
remain on the enclosed text. On a selected box, the same dropdown restyles its
corners in place while retaining dimensions and interior content. Unselected
prefix and suffix text stays on its own line so partial-line boxes keep their
columns aligned. Ctrl+D wraps a selection using the default box style; without a
selection it toggles box drawing mode. The popover's Draw button still enters
box mode. Arrow keys draw a connected path; Enter or Esc finishes drawing.
When a stroke starts on an existing box glyph, its foreground color is carried
to each glyph changed by that stroke. With a selection and an unfocused toolbar,
Tab/Ctrl+I toggles italics. The arrow towards the toolbar (Up above the selection,
Down below it) focuses it; Up and Down both focus it regardless of placement.
Left/Right dismiss it and move the editor cursor.
With a selection, plain Left collapses to its start and plain Right to its end,
closing the unfocused bar; the same editor behavior applies when the bar is
hidden. Shift+Left/Right and Ctrl+Shift+Left/Right keep extending selections.
Once focused on the main menu bar, the vertical arrow that entered it closes it
(Down below the text, Up above it); the other vertical arrow also closes it.
Left/Right navigate actions, and Enter or Space activates the focused action.
Submenus open with Enter or Space; inside them, Tab/Shift+Tab and arrows move
through choices, and Enter or Space applies one. Esc closes the submenu before
the bar. When the bar closes with Esc, an arrow,
Ctrl+Space, an action, a color pick, or an outside click, the selection clears
and the cursor stays in place. The ≡ Select all action keeps its new selection.
Another Ctrl+Space opens the popover at the cursor. Set
`PopoverDefaults.UnselectOnCancel` to false to retain selections when the bar closes.
Less common editor actions belong in the single ≡ hamburger submenu, not as
additional top-level toolbar buttons; it currently contains Select all.
When the bar is closed, Ctrl+Space selects a box or word, or opens a popover at
the cursor on whitespace. The
focused action is highlighted; Ctrl+Space starts at B, while Up/Down entry
follows uniformly active bold, italic, underline, or box selection, otherwise B.
Drawing past a line end
pads that row with spaces; drawing below the document appends rows and pads them
to the current display column. Up/Down can place the caret beyond the end of a
line or below the final line without editing; typing or drawing there pads the
document, while moving away leaves it unchanged. Ctrl+Space on blank space opens
a cursor-anchored popover so Draw can start a box stroke there. One mode session
is one undo step.

Set `ShowFileBar` to reserve the final editor row for a File menu and document
name/save status; it is off by default. The `loom widgets` gallery shows `F1 Help`
and responsive file shortcut hints below the editor, with F8/F9/F10 gallery
controls on the last row. F1 opens a list of editor, formatting, file,
selection, and picker bindings. Alt+F or clicking File opens the menu, `^S`
saves to `FilePath`, and Save as is available only in the File menu. Help scrolls
with arrow keys or the mouse wheel; clicking its backdrop closes it. F10 and
Ctrl+Q quit the gallery. Hosts check `ModalOpen()` and route input to the editor
first while a popup is open, including mouse events outside the editor bounds.
The save picker opens as a bounded, bordered popup with directory search active;
clicking outside closes it. `Save` and `SaveAs` use
`RichDocument.ToANSI` unless `SerializeDocument` supplies an application format.
Save errors remain available in `LastSaveError` and appear in the file bar.
The File menu contains document actions only; Ctrl+Space continues to open the
selection formatting popover. While the embedded ColorPicker is open, Up and
Down move between foreground and background rows at the same column; Tab follows the host's normal
focus traversal. Enter applies the active row and closes the picker, while Esc closes
without further application and Ctrl+Space closes the entire formatting popover.

View mode: `ViewMode` (or `NewRichTextView`) makes the editor read-only. Movement,
selection, mouse selection and copy (`Ctrl+C`/`Ctrl+Insert`) work; typing, `Ctrl+B/I/U`,
cut, paste, undo/redo, `Ctrl+Space` and the popover are off, and those keys return
`Ignored` so the app can use them. View mode does not place an edit caret, even
when focused. The gallery editor starts editable and uses F7 to switch between
View and Edit; its hotkey row shows the destination mode. The File bar never
shows an F7 hint. A span with `Link` set draws with the spec link
style (`rich_text_edit.link_fg`, `link_underline`) in both modes, focused or not;
the URL stays on `RichSpan.Link` (OSC 8 output is issue 259).

Editing shortcuts: `Ctrl+B/I/U` style the selection, or the word under the cursor
when nothing is selected (no word: only the typing style changes). `Ctrl+Space`
selects the smallest closed sharp or rounded box whose perimeter contains the
cursor cell, then shows the popover; reciprocal arms must connect every edge,
and equal-size boxes are resolved by top, left, bottom, then right position.
Interior text is unrestricted and side text outside the rectangle stays outside
the selection. On plain text, box interiors, or broken perimeters it selects the
word under the cursor as before; on whitespace it opens the popover at the
cursor without selecting text. The clipboard is internal and keeps
styled spans: `Ctrl+C`/`Ctrl+Insert` copy the selection or word, `Ctrl+X`/
`Shift+Delete` cut, `Ctrl+V`/`Shift+Insert` paste. `Ctrl+Z`/`Ctrl+Y` undo;
`Ctrl+R`, `Ctrl+Shift+Y`, `Ctrl+Shift+Z` redo (100 steps; a typing run up to a
word boundary, a style change, a cut or a paste is one step). Double-click
selects a word and triple-click a line. `Ctrl+I` arrives only from terminals
that speak the kitty/CSI-u protocol; on legacy terminals byte 0x09 stays Tab,
and `Ctrl+Shift+Y/Z` likewise need CSI-u.

VTE terminals such as Tilix and GNOME Terminal reserve Shift+Home, Shift+End,
Shift+PageUp, and Shift+PageDown for scrollback while an application uses the
primary screen. These keys reach Loom on the alternate screen. VTE does not
support CSI-u, so `Ctrl+Shift+A/E` cannot replace Shift+Home/End there. The
`loom widgets --show RichTextEdit` and combined `loom widgets --show` galleries
use the alternate screen; standalone non-editor demos remain inline.

`RichTextEdit.HotkeyBar()` returns a reusable `HintBar` with `F1 Help`, `^S Save`,
and `F7 View/Edit`. `HotkeyHint(width)` remains available as
plain text. The gallery draws the structured editor hints and its own
`F8 BG`, `F9 Theme`, and `F10 Quit` controls using the same cap styles.

`HintBar` accepts `HintEntry` values with a display `Key`, `Label`, decoded
`Binding`, and `Action func() EventResult`. Its keyboard binding and a left
press on either cap or label run that action; hover and release are inert.
Route mouse events relative to the bar's draw rectangle. It drops optional
`Detail` names before removing whole pairs from the end; `DropPriority` lets
an entry drop first. Colors come from
`ThemeColors.HintBarStyle()`; `ApplyTheme` updates the bar's styles.

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
  - **SGR/CSI Parser & Serializer**: `ParseAnsiBuffer`, `LoadAnsiBuffer`, `SerializeAnsiBuffer`, and `SaveAnsiBuffer` for lossless ANSI art roundtrips. `CSI K` (modes 0/1/2) fills the erased cells with the current background (BCE). The screen width is the widest printed column (at least the minimum), not the size of the cursor-addressed pre-scan, which can reach hundreds of columns. `ValidateAnsiBox` (ansibox.go) is line-based and skips the edge comparison at nested box corners.
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
- **Interactive Navigation**: Supports two-axis keyboard panning (`left`/`right`/`up`/`down`, `h`/`l`/`k`/`j`, `pgup`/`pgdn` and Shift+PgUp/PgDn, Space/`b`, `home`/`end`; Ctrl+F and Ctrl+B are left to the host, see the find and paging policy) and mouse wheel / trackbar dragging.

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

### Color Picker (`loom.ColorPicker`)

`ColorPicker` exposes its current `FG`, `BG`, text `Dim` attribute, active row as `Target`, and the
shared cursor's `Selected` palette index. `FG` and `BG` hold the applied
values independently of cursor movement. `ResetSelected` focuses the
terminal-default cell while keeping
`Selected` valid for `Palette`. `PreviewStyle()` combines the applied values
with the cursor cell on the active row.
`OnChange` reports each channel applied through Space, a swatch click or Enter;
`OnDimChange` reports changes from the Faint control. Faint is available after
the FG palette row, where Right from its last swatch focuses it; Left returns
to that swatch and Down moves to the final BG swatch. Space, Enter, or a left
click toggles the ANSI faint text attribute. It never changes either color.
`OnClose` runs after Enter, Esc or Ctrl+Space, with `DoneResult` so an embedded
host keeps running. Palette, edge glyphs, preview dimensions and key bindings
come from `spec/defaults.yaml`. The panel uses the rich text toolbar background
(`rich_text_edit.toolbar_bg`); its edge glyphs (`▂` top, `🮂` bottom by
default; eighths `▁`/`▔` as fallback for fonts without `🮂`) use that color as
their foreground on the editor background, so they extend the panel by a
quarter cell and clear the text under it. Swatches remain `Aa`
cells with one letter and one padding space per palette position. Row 3 stays
panel background through columns 0–8, then draws upper-edge glyphs from column
9; this closes the gap between the preview rug and the BG marker. Both rows always show a dim marker at their current applied color. When the
picker opens, this starts from the selected text's existing FG/BG; terminal
default colors use `∅`. The active row also marks the cursor in the theme's
`selection` colour, which takes priority when it overlaps its applied marker. `selection` is the editor selection and focused bar
item colour; `selected_bg` remains the list-row selection colour.

```go
picker := loom.NewColorPicker(loom.ColorIndex(7), loom.ColorIndex(0))
picker.OnChange = func(target loom.ColorPickerTarget, color loom.Color) {
    // Store the selected foreground or background.
}
picker.OnClose = func() { /* finish the picker session */ }
frame := &loom.Frame{Boxes: []loom.Box{{Child: picker}}}
```

The picker uses 36 columns by default, including one panel-background padding
column after the final swatch. Row 0 is a `▂` lower edge with the FG
selection highlighted. Row 1 starts with `∅ ` and shows all 16 `Aa` samples as
foreground colors on the panel background; row 2 shows the reset cell and
palette as background colors. Each row's swatches remain independent of the
other selected color. Up/Down changes rows while keeping the same cursor
column, including reset; Left/Right moves within a row, and Home selects reset. Space or a left click applies the active row and leaves the
picker open. Enter applies only the active row and closes it; the other row
keeps its applied value. For example, FG 1, Space, Down, Right, Enter produces
FG 1 on BG 2. Esc closes without further application. Tab follows the
containing focus container, and right-click has no picker action.

The `🮂` upper edge marks the BG selection beneath its swatch, including palette
15 at columns 33–34 beyond the 9-column preview rug, with panel padding at
column 35. Row 4 contains the Faint toggle. The preview label sits on
the preview FG/BG pair on row 5, with its 9-column rug and the bottom edge on
row 6. Edge cells
paint the panel background to clear text beneath the popup. Use `▁` and `▔` as
the lower and upper edge glyphs in fonts without `🮂` (Symbols for Legacy
Computing, Unicode 13). These glyphs and their padding come from
`spec/defaults.yaml`. Set `color_picker.show_preview` to false to remove the
preview rug.

ANSI SGR 2 represents faint text, not a subdued background. Loom's `Style.Dim`
therefore affects text foreground rendering only; `ThemeColors.NormalBG` is a
theme surface color and cannot represent a dimmed version of an arbitrary
background. Loom has no semantic dim-background style attribute, so a subdued
background option cannot be offered without inventing a color blend.

The `loom widgets ColorPicker` gallery keeps one interactive picker and shows
three static comparisons below it: Default uses terminal-reset foreground and
background, Faint text adds the ANSI dim text attribute, and Theme surface (BG)
uses terminal-reset foreground with the current theme's `NormalBG`. The last
sample illustrates a theme surface only; it is not a semantic dim-background
color, and themes whose `NormalBG` is terminal reset may render it like Default.
The samples sit side by side when there is room and stack at narrower widths.

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

---

## 17. Modal Overlay Architecture, Temporary Mouse Grab & Backdrop Isolation

Loom provides a library-level modal contract for overlays (`Popup`, `Dialog`, `Menu`, `RichTextEdit` format popovers):

### The `ModalTarget` and `ModalMouseTarget` Contract
Widgets hosting an active modal overlay implement `ModalTarget`:
```go
type ModalTarget interface {
	ActiveModal() (Widget, Rect, bool)
}
```
- **Exclusive Mouse Routing**: When `ActiveModal()` returns `ok == true`, `Pane` and standard composite containers (`Frame`, `Split`, `Stack`, `Grid`, `Tabs`, `Viewport`, `Form`) automatically forward mouse events to the modal target before performing standard hit-testing, focus updates, or background click handling.
- **Backdrop Click Isolation**:
  - Clicking outside the bounding box of a `Popup` dismisses the popup and returns `loom.Handled()`.
  - Clicking outside a `Dialog` invokes its `Cancel` button action and returns `loom.Handled()`.
  - In `RichTextEdit`, clicking outside an open popover dismisses the popover without leaking the mouse press, drag, or release into document text, selection logic, or cursor movement.
- **Retained Gesture Handling**: Gestures (press-drag-release) started on an overlay or backdrop are retained until release, guaranteeing trailing release/drag events never bleed into underlying widgets.
- **Modeless Popups/Dialogs**: Embedded or static previews in demo galleries can opt out using `Modeless: true` to prevent unintended modal capture.
- **Pane-Level Drag Capture (issue 449)**: A scrollbar drag in `View` or `Viewport` claims the pointer until release. While the claim holds, `Pane` routes drag and release events to the claiming widget, even when the pointer leaves its bounds or returns. Containers (`Stack`, `Grid`, `Split`, `Tabs`, `Frame`, `Form`, `Box`, `Viewport`) forward the claim query to their children, so a drag survives leaving the widget. Coordinates reaching the claimant are 0-based and child-local. A press or a hover ends the claim, as does `Esc`, which also restores the offset from before the press. Widgets that never claim keep hit-test routing. The claim contract is unexported, so only loom widgets claim captures.
- **Middle-Button Pan (issue 452)**: A middle press on content in `View` or `Viewport` starts a scrub that maps the pointer row onto the whole scroll range: the second row of the widget is the first line and the second-to-last row is the end (so the pointer need not leave the widget; a view under 4 rows uses its first and last row), and the press row keeps the offset the press found, so nothing jumps. Dragging up scrolls toward the top, dragging down toward the end; rows outside the widget clamp to the ends. The pan claims the pointer through the same pane capture, so it continues outside the widget. A `Viewport` child sees the press first; a child that consumes it keeps it and no pan starts. A middle press on the scrollbar column does not pan, and a left press never does. Release, a press or a hover ends the pan where it is. Focus loss and `Esc` end it and restore the offset from before the press. The pan is vertical only and whole-line; wheel, keys and the thumb keep their half-row positions (450).
- **Wheel Acceleration (issue 451)**: `View` and `Viewport` need no change for fast scrolling: `Pane` repeats a wheel report during a fast burst (see `docs/TuiInput.md`). A slow notch moves as it did before (in foot, one notch is about 3 reports, so 3 lines), and a fast burst of the same number of reports moves further. The curve (`pane.wheel` in `spec/defaults.yaml`) is unverified.
- **Half-Row Scrollbar Thumbs (issue 450)**: `View` and `Viewport` place the thumb in half-row units, two per track row, so a short bar has a distinct picture for every offset. With a themed track and thumb (`ScrollbarStyle` with a background on each, and different colours) and colour output, each cell is a space on the track or thumb background, or a lower half block (`▄`) where the thumb starts or ends inside the row. Without colour, or when the track and thumb share a background (every shipped theme today), the bar uses whole-cell `scrollbar.foreground_char`/`background_char` glyphs, with the thumb rounded to the nearest cell edge. A drag keeps the half-row phase of the press. A track click on a `Viewport` centers the thumb. `View` track clicks keep their proportional mapping. `Esc` still restores the offset from before the press.

### Temporary Mouse Grab in Mouse-Off Mode (`"m"`)
- **Base Mode**: In `"m"` mode, terminal mouse tracking is fully disabled (`\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l`), allowing unobstructed native terminal drag-to-select and copy.
- **Temporary Engagement**: When any interactive modal overlay opens, `Pane` automatically engages temporary mouse tracking (`\x1b[?1000h\x1b[?1006h` / `\x1b[?1003h\x1b[?1006h`) so users can click buttons, select menu options, or scroll.
- **Automatic Restoration**: When all overlays close (or the backdrop is clicked), `Pane` automatically restores base mouse-off mode. If persistent global grab (`"M"`) was already active, mouse tracking remains permanently enabled.

## 18. Plain-Text Entry (`loom.TextEntry`)

`loom.TextEntry` is a multi-line text input for chat-style lines. The host embeds it, gives it a width and height through `Draw` or `Measure`, and receives plain text through `Value`, `OnChange`, and `OnSubmit`. Optional bold formatting is enabled with `EnableBold`; `MarkdownValue` serializes that formatting for hosts that submit Markdown. There is no folding, and tabs become four spaces.

```go
entry := loom.NewTextEntry(loom.TextEntryOptions{
    Placeholder: "Message the team…",
    Prompt:      "> ",
    MaxHeight:   4,
    OnSubmit: func(value string) loom.EventResult {
        send(value)
        return loom.Consumed()
    },
})
entry.SetFocus(true)
```

### Options and defaults

| Field | Default | Meaning |
|---|---|---|
| `Placeholder` | empty | Dimmed text shown while the entry is empty. Only its first line is drawn. |
| `Prompt` | empty | Prefix on the first row. Continuation rows are indented by its width. `SetPrompt` changes it at runtime (issue 455); the text re-wraps at the next `Measure` or `Draw`. |
| `Height`, `MinHeight`, `MaxHeight` | 1, 1, automatic | Text rows shown while short, the fewest rows, and the most before the entry scrolls. Automatic is 30% of the app height (the canvas height at the latest `Draw`; the terminal height before one; 24 rows without a terminal), from `text_entry.*` in `spec/defaults.yaml`. A negative `MaxHeight` is unbounded. `MinHeight` wins over `MaxHeight`. `SetHeights` changes them at runtime. |
| `Frame` | none | Border decoration, see Frame below. |
| `SendButton`, `SendLabel` | off, `" ➔ "` | A clickable send button at the right of the last row. A left click calls `OnSubmit` like a submit key. |
| `SubmitKeys` | `["enter", "ctrl-enter"]` | Keys that call `OnSubmit`. Ctrl+Enter is only reported by terminals with the kitty keyboard protocol (or modifyOtherKeys); elsewhere it arrives as Enter or Ctrl+J. Without `OnSubmit`, they are ignored. |
| `NewlineKeys` | `["shift-enter", "ctrl-j", "alt-enter"]` | Keys that insert a line break. |
| `OnChange` | nil | Called with the new value after each text change. Not called by `SetValue`. |
| `OnSubmit` | nil | Called with the current value. Its result is returned to the caller. |
| `Highlight` | nil | Called on each draw with the value; returns `TextEntrySpan`s that style byte ranges. With bold enabled, selection adds inversion while retaining span colors and bold; otherwise selection keeps the legacy selection-only style. |

The defaults live in Go, not in `spec/defaults.yaml`, so the `specdefaults` API stays unchanged.

### Frame and send button (issue 454)

`Frame` (or `SetFrame`) wraps the entry; `Measure` includes it and `Cursor()` is relative to the whole widget. Styles:

```
box (the taia.legacy look)          simple (full-width rule)
╭─ title ──────── hint ─╮           ── title ─────── hint ──
│ ❯ message             │           ❯ message
│   more         ➔      │             more         ➔
╰─────────────── footer ─╯           ─────────────── footer ──
```

- `Title` and `Hint` sit in the top border (left, right), `Footer` in the bottom border (right); `SetTitle`, `SetHint` and `SetFooter` change them. A label that does not fit is cut, and the hint goes first.
- `BorderColor` and `LabelColor` take `""` or `dim` (default), `default`, `#RRGGBB`, a palette index, or a theme color name such as `border_fg` or `status_fg` (the yaml names of `ThemeColors`, resolved against the theme from `ApplyTheme`). Unknown names are dim.
- A rect too small for the frame (under three rows or too narrow for its corners) draws the content only.
- With a send button and at least two text rows, the scrollbar sits above the button (over its middle cell) and the text keeps its full width; with one row it stays beside the text.
- A press on the border does nothing. The send button is hidden when the width is too small for it.
- The gallery TextEntry demo switches the frame (Alt+F), the send button (Alt+B), the border color (Alt+C) and the height, min and max rows (Alt+G, Alt+M, Alt+X). Its footer names the send key ("Enter sends" or "Alt+Enter sends", following Ctrl+T), and the debug rows show the height settings.

### Placeholder

- `Placeholder()` returns the hint, and `SetPlaceholder(s)` replaces it at runtime, for example with a model-proposed next prompt. Setting it does not call `OnChange`, and the next `Draw` shows the new hint.
- The hint is drawn only while the value is empty, and typing hides it. The hint is never part of `Value()`, and the caret stays at the start of the row while it shows.
- Only the first line of a multi-line hint is drawn, cut at the row width. Wide runes are never split, so a rune that does not fit is left out.
- `AcceptPlaceholder()` turns the hint into the value. It works only when the value is empty and the hint is not empty. It inserts the whole hint, including any further lines, puts the caret at its end, makes one undo step, calls `OnChange` once, and returns `true`. Otherwise it returns `false` and changes nothing. Tab, End and Ctrl+E also accept it when the value is empty and the caret and anchor are at the start; this consumes the key without submitting. Tab is an intentional widget action in this case.


### Soft wrap

- Rows wrap at word boundaries (issue 447): a word that fits on a row is not split, and the space after a word stays at the end of its row. A word with no earlier space on its row, or wider than a row, still breaks at the row edge, so no text is lost. The value and its explicit newlines are unchanged.
- The last cell of each row stays free for the caret. A space that ends a full row may use that cell, except at the end of a line.
- The caret after a wrap space belongs to the start of the next row, as at any soft wrap. Up, Down, mouse placement, selection and `Cursor()` follow the rendered rows.

### Key handling

- **Editing keys** are consumed: printable text, Backspace and Delete, the arrow keys with Home, End, and Ctrl+Left and Ctrl+Right, and Shift with any of them to extend a selection.
- **Ctrl+U** deletes from the current visual segment start to the caret, or deletes only the selection. At a hard-line start it moves before the newline without deleting it; at the first start it is ignored. At a soft-wrap boundary it immediately clears the preceding segment. Before layout it uses hard lines; after layout it uses the latest width. Each deletion is one undo step and calls `OnChange`; navigation does neither.
- **Clipboard keys**: Ctrl+A selects all. Ctrl+C copies, Ctrl+X cuts, and Ctrl+V pastes from the internal clipboard. Copy and cut also write to the host clipboard through the Pane hook (OSC 52). Ctrl+C is ignored when there is no selection. Hosts that handle Ctrl+C before the entry can use `HasSelection()` to route it to the entry when text is selected and otherwise retain their own action. Bracketed paste inserts as one undo step.
- **Undo**: Ctrl+Z undoes, and Ctrl+Y or Ctrl+Shift+Z redoes. Consecutive typed characters form one undo step.
- **Bold**: call `EnableBold()` to make Ctrl+B available to the entry. With a selection it toggles all selected non-newline graphemes together: a mixed selection becomes bold, and an all-bold selection becomes plain. Without a selection it toggles the style for later typing. Formatting and typing-style changes participate in undo/redo without calling `OnChange`; inserted text, newlines, paste, range replacement and accepted placeholders use the current typing style. SetValue clears stored formatting and typing style. Formatting changes never change copied or cut text.
- **Host keys** are returned `Ignored`: Tab outside placeholder acceptance, Shift+Tab, Ctrl+D, Ctrl+O, Ctrl+S, Ctrl+G, Ctrl+F, Esc, and F-keys. Ctrl+D does not edit, submit, or quit the entry; a parent can handle it to move focus or exit. A Frame offers it to the focused child before its own bindings. Up, Down, PageUp and PageDown are ignored at the first or last row so the host can use them for history.
- **Page keys**: the decoder emits `pgup` and `pgdown`. Bindings also accept `pageup`, `pagedown` and `pgdn`.

### Mouse

Coordinates are 0-based and relative to the widget. A press places the caret, and a drag extends the selection. A double click selects a word, and a triple click selects a logical line. The wheel scrolls without moving the caret, and the scroll position holds until the caret moves again.

### Offsets, replacement and highlight

Offsets are byte offsets into plain `Value()`. An offset inside a grapheme cluster snaps to the start of that cluster, and an offset outside the text clamps to `[0, len(Value())]`.

- `CursorOffset()` returns the caret as a byte offset. `SetCursorOffset(off)` moves the caret and clears the selection. Neither makes an undo step or calls `OnChange`.
- `ReplaceRange(start, end, s)` replaces the text between two offsets with `s`. A reversed range is swapped. It is one undo step, the caret moves to the end of `s`, and it does not call `OnChange`. Use it for completion: replace the word at the caret.
- **Scrollbar (issue 448)**: when the text needs more rows than the entry shows, the last column becomes a scrollbar (same half-row thumb as `View`, themed by `ApplyTheme`) and the text wraps one column narrower, so the caret cell never sits under the bar. Rows that fit show no bar and wrap at the full width. A press on the thumb starts a drag that survives leaving the entry (pane mouse capture); `Esc` restores the scroll; a track click jumps. Scrolling by the bar, like the wheel, does not move the caret and stays until the caret moves. The wheel gets the pane acceleration. `Scrollbar` mode follows the spec default (`scrollbar.mode`).
- **Middle scrub and send key (issue 453)**: while the scrollbar shows, a middle press on the text scrubs the scroll range as in `View` (`Esc` restores). `SetSubmitKeys(submit, newline)` swaps the Submit and Newline bindings at run time, for example Alt+Enter sends and Enter breaks the line; empty lists restore the defaults.
- `Highlight` returns `[]TextEntrySpan{Start, End, Style}`. Spans clamp, snap to cluster starts, and a later span wins over an earlier one. A reversed span is ignored. With bold enabled, stored bold is composed with highlight styles and selection adds inversion while retaining color and bold attributes. Without bold, selection keeps the legacy selection-only style.
- `MarkdownValue()` returns escaped Markdown with bold runs wrapped in `**`. Literal ASCII punctuation and backslashes are escaped. Runs close at each hard newline, and the original newline characters are preserved. Leading and trailing whitespace stays outside bold delimiters; whitespace-only bold has no Markdown emphasis semantics. `Value()`, `OnChange`, `OnSubmit`, and clipboard contents remain plain text.

```go
entry := loom.NewTextEntry(loom.TextEntryOptions{
    Highlight: func(value string) []loom.TextEntrySpan {
        if strings.HasPrefix(value, "/") {
            return []loom.TextEntrySpan{{Start: 0, End: len(value), Style: loom.Style{FG: loom.ColorIndex(6)}}}
        }
        return nil
    },
})
```

### Known limits

- Ctrl+J and Shift+Enter are distinct from Enter only when the terminal sends the kitty keyboard protocol (CSI u). In legacy terminals, Ctrl+J arrives as plain LF, which the decoder reports as `enter`, and Shift+Enter arrives as Enter. Alt+Enter works in both.
- There is no design sketch for this widget yet, so the per-cell checks in its tests assert the contract directly: selection inversion, bold and highlight colors, Dim on the placeholder, the prompt on row 0, and the hardware cursor at the caret.

