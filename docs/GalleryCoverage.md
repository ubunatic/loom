# Gallery coverage audit

Inventory of the public widget catalog against the live gallery, after issue 423 changes. A tab demonstrates a sample, not necessarily every option or interaction.

| Widget | Gallery | Notes |
| --- | --- | --- |
| `loom.AlignBox` | AlignBox |  |
| `loom.AnsiEditor` | AnsiEditor |  |
| `loom.Box` | Box |  |
| `loom.Chart` | Chart |  |
| `loom.Choice` | Choice |  |
| `loom.ColorPicker` | ColorPicker |  |
| `loom.Confirm` | Confirm |  |
| `loom.DatePicker` | DatePicker |  |
| `loom.Dialog` | Dialog |  |
| `loom.FilePicker` | FilePicker |  |
| `loom.Form` | Form |  |
| `loom.Frame` | Frame |  |
| `loom.Gauge` | Gauge |  |
| `loom.Grid` | All | The combined grid is itself a Grid. |
| `loom.HintBar` | HintBar |  |
| `loom.KeyHelp` | KeyHelp |  |
| `loom.Menu` | MenuBar | Dropdowns are hosted by MenuBar. |
| `loom.MenuBar` | MenuBar |  |
| `loom.Notif` | Notif |  |
| `loom.NumberInput` | NumberInput |  |
| `loom.Paginator` | Paginator |  |
| `loom.PaintCanvas` | PaintCanvas |  |
| `loom.PillCluster` | PillCluster |  |
| `loom.Popup` | Popup |  |
| `loom.ProgressBar` | ProgressBar |  |
| `loom.RichTextEdit` | RichTextEdit |  |
| `loom.Router` | Router |  |
| `loom.Rows` | Rows |  |
| `loom.SearchBar` | SearchBar |  |
| `loom.Settings` | Settings |  |
| `loom.Sparkline` | All | Compact graph in the combined grid. |
| `loom.Spinner` | Spinner |  |
| `loom.SplashView` | SplashView |  |
| `loom.Split` | Split |  |
| `loom.Stack` | Stack |  |
| `loom.Stopwatch` | Stopwatch |  |
| `loom.StyledRows` | StyledRows |  |
| `loom.Table` | Table |  |
| `loom.Tabs` | Tabs |  |
| `loom.TextArea` | TextArea |  |
| `loom.TextEntry` | TextEntry | Wrap, command highlighting, Ctrl+B bold and Markdown submission are shown. Submitted Markdown appears as a quoted, escaped, width-wrapped preview so newlines and controls remain visible; overflow is marked. F2 changes the placeholder, F3 or Tab, End, and Ctrl+E accept it when empty, and F4 moves the caret. The entry is four rows high so a few lines already scroll. Below it, between two separator lines, a debug view shows length, lines, caret and selection (it yields first on small windows). One `HintBar` is the only key bar: Ctrl+B and Enter are labels, F2-F4 are the controls above, F5-F7 insert a short, multi-line and long (wrapping, wide-character) sample at the caret; keys and clicks run the same actions. F8-F10 belong to the gallery. |
| `loom.TextInput` | TextInput |  |
| `loom.Timer` | Timer |  |
| `loom.Toggle` | Toggle |  |
| `loom.Tree` | Tree |  |
| `loom.View` | Viewport | View supplies the scrollable child. |
| `loom.Viewport` | Viewport |  |
| `media.Widget` | Media | Generated still image; video is not shown. |

| Other visual feature | Gallery | Notes |
| --- | --- | --- |
| `graph` bar and sparkline | Gauge, All | Gauge shows a bar; All shows a sparkline. |
| `graph` stacked bar and treemap | Gauge | Static sample data illustrates both renderers. |
| `syntax` highlighting | TextArea, TextEntry | TextArea uses the Go lexical engine; TextEntry uses a custom command span hook. |

The gallery illustrates representative options, not every rendering variant. Graph layout and aggregation APIs and syntax coordinates and outline APIs produce data rather than interactive widgets; their effects are exercised by the displayed renderers and editor. Media video requires an external source and is not included in the self-contained sample.
