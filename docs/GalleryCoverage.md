# Gallery coverage audit

Inventory of the public widget catalog against the live gallery, before issue 423 changes. A tab demonstrates a sample, not necessarily every option or interaction.

| Widget | Gallery | Notes |
| --- | --- | --- |
| `loom.AlignBox` | missing |  |
| `loom.AnsiEditor` | missing |  |
| `loom.Box` | Box |  |
| `loom.Chart` | Chart |  |
| `loom.Choice` | Choice |  |
| `loom.ColorPicker` | ColorPicker |  |
| `loom.Confirm` | missing |  |
| `loom.DatePicker` | DatePicker |  |
| `loom.Dialog` | Dialog |  |
| `loom.FilePicker` | FilePicker |  |
| `loom.Form` | Form |  |
| `loom.Frame` | missing |  |
| `loom.Gauge` | missing |  |
| `loom.Grid` | All | The combined grid is itself a Grid. |
| `loom.HintBar` | HintBar |  |
| `loom.KeyHelp` | KeyHelp |  |
| `loom.Menu` | MenuBar | Dropdowns are hosted by MenuBar. |
| `loom.MenuBar` | MenuBar |  |
| `loom.Notif` | missing |  |
| `loom.NumberInput` | NumberInput |  |
| `loom.Paginator` | Paginator |  |
| `loom.PaintCanvas` | PaintCanvas |  |
| `loom.PillCluster` | PillCluster |  |
| `loom.Popup` | Popup |  |
| `loom.ProgressBar` | ProgressBar |  |
| `loom.RichTextEdit` | RichTextEdit |  |
| `loom.Router` | missing |  |
| `loom.Rows` | missing |  |
| `loom.SearchBar` | SearchBar |  |
| `loom.Settings` | missing |  |
| `loom.Sparkline` | All | Compact graph in the combined grid. |
| `loom.Spinner` | Spinner |  |
| `loom.SplashView` | missing |  |
| `loom.Split` | missing |  |
| `loom.Stack` | missing |  |
| `loom.Stopwatch` | Stopwatch |  |
| `loom.StyledRows` | missing |  |
| `loom.Table` | Table |  |
| `loom.Tabs` | Tabs |  |
| `loom.TextArea` | TextArea |  |
| `loom.TextEntry` | TextEntry | Wrap, submit and command highlighting are shown; runtime placeholder and caret API are not exposed as controls. |
| `loom.TextInput` | TextInput |  |
| `loom.Timer` | Timer |  |
| `loom.Toggle` | Toggle |  |
| `loom.Tree` | Tree |  |
| `loom.View` | Viewport | View supplies the scrollable child. |
| `loom.Viewport` | Viewport |  |
| `media.Widget` | Media | Generated still image; video is not shown. |

| Other visual feature | Gallery | Notes |
| --- | --- | --- |
| `graph` bar and sparkline | All | ProgressBar and Sparkline samples; graph rendering options are not exhaustively shown. |
| `graph` stacked bar and treemap | missing | Rendering functions have no dedicated gallery demo. |
| `syntax` highlighting | TextEntry | Custom command span hook only; lexical syntax providers are not demonstrated. |

This inventory is a starting point: it does not establish coverage of every option, interaction, or exported visual function.
