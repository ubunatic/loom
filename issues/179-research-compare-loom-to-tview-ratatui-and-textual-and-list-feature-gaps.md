# 179 — Research: compare loom to tview, Ratatui and Textual and list feature gaps

**Status**: Closed — findings consumed by 180
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Research
**Related**: issues/180 (roadmap), issues/177-179

---

## 1. Problem & Motivation
These are the leading widget-oriented TUI frameworks in Go, Rust and Python. We want a list of features tview, Ratatui and Textual offers that loom lacks, so the roadmap (180) can close the gaps that matter.

## 2. Scope & Rules
- Stay true to loom's design (read AGENTS.md, docs/Widgets.md, README, `loom widgets`): a gap is only a gap if it fits loom's model. Features that contradict the design go in a "Deliberately not adopted" list with one-line reasons.
- Check live code before claiming a gap; cite loom file/symbol for what exists.
- Cross-reference existing open tickets (`harnez find -d . issues -a status:open`, e.g. 155 modal, 158 keymap, 167-173 widgets) instead of re-listing them as new.
- Output: append a `## Findings` section to this ticket: a table `Feature | <other> | loom today | Gap? | Fits design? | Existing ticket | Size (S/M/L)`, then the not-adopted list. No code changes.

## 3. Implementation & Verification Plan
/goal The Findings section lists every relevant tview, Ratatui and Textual feature with loom's status and a design-fit verdict, committed as `docs(issues): findings for 179`. Stop and report when blocked on a user decision or denied permission.

## Findings

Compared the frameworks' documented built-in widgets and app facilities with Loom's live widget catalog (`loom widgets`), `spec/widgets.yaml`, and the implementations cited below. “Gap” means a missing built-in capability, not that an application cannot compose it itself. Framework documentation checked 2026-09-30.

| Feature | Other | loom today | Gap? | Fits design? | Existing ticket | Size (S/M/L) |
|---|---|---|---|---|---|---|
| Single-line input, filtered choice/list, multiline editor, confirmation | tview: InputField, List, DropDown; Textual: Input, Select, OptionList; Ratatui: List | `TextInput`, filterable/multi-select `Choice`, `TextArea`, and `Confirm` are catalogued (`spec/widgets.yaml`; `choice.go`, `textarea.go`, `textinput.go`, `confirm.go`). Settings also covers typed values. | No core gap; see masked input row. | Yes; present. | — | — |
| Typed standalone number and toggle controls | tview: Checkbox; Textual: Checkbox, Switch; Ratatui leaves controls to app widgets | `Settings` supports `KindBool`, `KindString`, `KindChoice`, and bounded `KindNumber` (`settings.go`), but there are no standalone NumberInput or Toggle catalog entries. | Yes, ergonomic standalone controls. | Yes | [#174](174-add-standalone-numberinput-and-toggle-widgets.md) | S |
| Masked/password input | tview: password form field; Textual: password `Input` / `MaskedInput` | `TextInput` is present; the catalog and `textinput.go` have no password/masking mode. | Yes | Yes | [#173](173-add-masked-password-mode-to-textinput.md) | S |
| Form composition, field navigation, validation summary | tview: `Form` composes text, password, dropdown, checkbox, and buttons; Textual apps compose controls and handle messages/validation | `Frame` and focus traversal can compose widgets, but no Form/FieldGroup, field validation contract, or validation summary is catalogued (`frame.go`, `focus.go`, `spec/widgets.yaml`). | Yes, as a reusable form abstraction. | Yes | [#169](169-add-form-and-fieldgroup-composite-layout-widget-with-tab-navigation-and-validation-summary.md) | M |
| Hierarchical tree with expansion and selection | tview: `TreeView`; Textual: `Tree`; Ratatui ecosystem: tree widget | No library tree widget in `spec/widgets.yaml`; the graph treemap is a visualization, not a navigable node tree. | Yes | Yes | [#168](168-add-tree-widget-with-hierarchical-nodes-expand-collapse-and-keyboard-navigation.md) | M |
| Menus and shortcut accelerators | tview: List shortcuts and composed menus; Textual: `OptionList` and command palette; Ratatui leaves menu behavior to applications | Loom has key dispatch and `Choice` command mode (`choice.go`), but no menu/menu-bar widget. | Yes for discoverable nested menus; command-palette behavior is a separate app-level choice. | Yes for local menus | [#170](170-add-menu-and-menubar-widget-with-nested-dropdowns-and-shortcut-accelerators.md) | M |
| Date picker / month calendar | tview has no dedicated calendar in its documented widget list; Ratatui has `Calendar`; Textual provides date-oriented app composition rather than a core date picker | No DatePicker or Calendar in `spec/widgets.yaml`; `loom widgets` confirms the current catalog. | Yes | Yes | [#175](175-add-a-datepicker-widget.md) | M |
| Reusable filesystem picker | tview `TreeView` and Textual `DirectoryTree` provide navigable hierarchy; Textual also has file-dialog patterns | Loom has `ReadDirectory` and the example `filebrowser.NavigationPane`, but no catalogued library `FilePicker` (`docs/Widgets.md` §4 and §13). | Yes, as a library widget on the existing directory model. | Yes | [#172](172-add-a-reusable-filepicker-widget-on-top-of-the-directory-model.md) | M |
| Modal/dialog overlay with conventional buttons and outside-click behavior | tview: `Modal`; Textual: modal `Screen`; Ratatui: `Clear` plus app composition | `Popup` centers one child and captures events while open (`popup.go`); `Confirm` supplies yes/no input, but there is no complete dialog stack/button/placement contract. | Partial; generic overlay exists, richer built-in dialog behavior is tracked. | Yes | [#155](155-built-in-modal-and-dialog-overlay-primitive.md) | M |
| Data table with cell cursor, fixed headers/columns, rich cells and per-cell events | tview: selectable/highlightable table; Ratatui: stateful table selection; Textual: `DataTable` supports cell/row/column cursors, fixed rows/columns and selection messages | `Table` has string cells, row selection, sorting, filtering and scrolling (`table.go`); no cell cursor, fixed panes, or rich renderable cells. | Yes for the richer data-table interaction model. | Yes | — | M |
| Multi-series charts and line/bar chart widgets | Ratatui: `Chart`, `BarChart`, `LineGauge`; Textual: charting is supplied by composable/custom widgets; tview emphasizes tables/text rather than a built-in chart | Loom has `Gauge`, `ProgressBar`, `Sparkline`, and standalone graph renderers (`widget_graph.go`, `graph/bar.go`, `graph/stackedbar.go`, `graph/treemap.go`); no axes-based line or grouped bar chart widget. | Yes, for common multi-series plots. | Yes; a focused widget complements Loom's existing graph primitives. | — | M |
| Tabs and layouts (split/flex/grid/pages) | tview: Flex, Grid, Pages; Ratatui: composable layout plus Tabs; Textual: containers, tabs and screens | `Frame`, `Grid`, `Split`, `Stack`, `Tabs`, and `Router` cover these patterns (`spec/widgets.yaml`; `frame.go`, `grid.go`, `split.go`, `stack.go`, `tabs.go`, `router.go`). | No material gap in the widget-oriented layout model. | Yes; already present. | — | — |
| Styled scrollable text, syntax-highlighted editor, progress and compact metrics | tview: TextView/TextArea; Ratatui: Paragraph, Gauge, Sparkline, Scrollbar; Textual: Static, TextArea, ProgressBar, Sparkline | `View`, `TextArea` with syntax/folding, `ProgressBar`, `Gauge`, and `Sparkline` are present (`view.go`, `textarea.go`, `syntax/`, `progressbar.go`, `widget_graph.go`). | No core widget gap. Loom's editor is intentionally smaller in scope than a full editor. | Yes; present. | #167 is still listed open, although the catalog and `progressbar.go` show its determinate ProgressBar is implemented. | — |
| Image/video rendering | tview: `Image`; Textual supports app-defined image rendering; Ratatui ecosystem offers protocol-aware image widgets | Loom `media.Widget` renders still images and video frames using half-block cells (`media/` and `docs/Widgets.md` §12). Playback controls and poster frames remain separate work. | Partial interaction gap; base rendering exists. | Yes | [#112](112-add-media-controls-play-pause-zoom-and-panning-for-the-image-media-widget.md), [#160](160-media-widget-playback-controls-and-poster-frame-support.md) | M |
| System clipboard integration | tview's ecosystem exposes cross-platform clipboard support; terminal frameworks can also receive bracketed paste | `AnsiBuffer` has an editor-local cell clipboard (`ansibuffer.go`); no general system clipboard API is present in the widget catalog. | Potential gap, but no portability contract or user need is established by this comparison. | Unclear; assess separately before roadmap commitment. | — | — |

### Deliberately not adopted

- **Full-screen application ownership as the default runtime** (tview Application, Textual App, and the usual Ratatui loop): Loom's `Pane` reserves and redraws an inline terminal region while restoring terminal state (`pane.go`); replacing that with mandatory full-screen ownership would contradict Loom's inline integration model. Hosts may still choose a full-screen environment around Loom widgets.
- **Textual's DOM, reactive message tree, and external CSS/TCSS layout/styling system:** Loom uses explicit Go widget composition (`Widget`, `Frame`, `Split`, `Grid`, `Stack`) and spec-backed themes (`docs/Widgets.md`, `theme.go`). A second DOM/style engine is not a widget gap; reusable controls and themes should remain in Loom's existing model.
- **Framework-managed general-purpose asynchronous workers:** Loom provides `Ticker`, goroutine-safe `Pane.Invalidate`, and lazy background media rendering (`pane.go`, `media/`; `docs/Widgets.md` §§5, 12). Owning arbitrary application tasks, cancellation policy, and result messaging would expand the inline widget layer into an application framework; applications can manage their own workers and invalidate on updates.
- **Terminal-specific graphics protocol as a required image path:** Loom's half-block media renderer is portable across ordinary cell terminals. Protocol-specific image output is an optional backend direction, not a gap in the current portable rendering contract.

### Framework references

- tview's built-in widget list and model: [package documentation](https://github.com/rivo/tview/blob/master/doc.go); [Application event loop and input/paste/mouse capture](https://github.com/rivo/tview/blob/master/application.go).
- Ratatui's built-in widget list: [widget concepts](https://ratatui.rs/concepts/widgets/); [calendar example](https://ratatui.rs/examples/widgets/calendar/); [third-party tree and image widgets](https://ratatui.rs/showcase/third-party-widgets/).
- Textual's widget catalog and representative controls: [widget gallery](https://textual.textualize.io/widget_gallery/); [DataTable](https://textual.textualize.io/widgets/data_table/); [Tree](https://textual.textualize.io/widgets/tree/); [Select](https://textual.textualize.io/widgets/select/).
- Textual application features considered: [screens](https://textual.textualize.io/guide/screens/); [CSS](https://textual.textualize.io/guide/CSS/); [workers](https://textual.textualize.io/guide/workers/); [command palette](https://textual.textualize.io/guide/command_palette/); [testing/Pilot](https://textual.textualize.io/guide/testing/).
