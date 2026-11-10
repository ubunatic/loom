# Topic 3: Bubbles and Huh widgets and forms vs loom

Scope: ticket 457, topic 3. Feature rows are in `features-bubbles.yaml` (38 rows). The `loom` column is left `unknown` for the host.

## Versions and sources

- Both libraries are on the Bubble Tea v2 line. Module lines verified in raw `go.mod`:
  `module charm.land/bubbles/v2` (bubbles main `1deada9`) and `module charm.land/huh/v2` (huh main `ffb6a97`).
  All rows are tagged `main@<sha>`. No Bubble Tea v1 code was read, so no v1 claims are made.
- Pins came from the GitHub API for `main` at research time. Line numbers come from `raw.githubusercontent.com`
  files on `main`, located with `grep -n`. Source links are pinned to the full SHA.
- Web search gave only mirrors and third-party skill pages. These were used as leads only. Every row cites a fetched file.

## Bubbles: component list (verified)

The bubbles README names: Spinner, Text Input, Text Area, Table, Progress, Paginator, Viewport, List, File Picker,
Timer, Stopwatch, Help, Key. The repo also has `tree/` and `cursor/` packages. `tree` is not in the README.
(Source: bubbles README and repo tree, `main@1deada9`.)

Read in this run: `textarea`, `textinput`, `filepicker`, `table`, `list` (partial), `tree` (signatures only).
Not read: `viewport`, `progress`, `paginator`, `spinner`, `timer`, `stopwatch`, `help`, `key`, `cursor`.

## Bubbles findings against loom

**Input and key bindings (P1 items).** Bubbles uses the same key sets in `textarea` and `textinput`:
- Word jump: `alt+right`, `ctrl+right`, `alt+f` (`textarea.go` L93, `textinput.go` L72).
- Shift-select by word: `ctrl+shift+right`, `alt+shift+right` (`textarea.go` L120).
- Select all is `ctrl+g` (`textarea.go` L124), not Ctrl+A. Ctrl+A is line start (`home`, `ctrl+a`).
  Loom uses Ctrl+A for select all, so bubbles is a different desktop mapping.
- Copy selection is `ctrl+shift+c` (`textarea.go` L125). Ctrl+Shift keys only arrive with the kitty keyboard protocol.
- Paste is `ctrl+v` in both (`textarea.go` L108, `textinput.go` L82).
- Undo and redo: none. A grep of the full `textarea.go` and `textinput.go` for `undo|redo` returned no match.
  Loom TextEntry has undo/redo, so loom is ahead here.
- `textinput` binds `tab` to `AcceptSuggestion` (`textinput.go` L83). This is a widget consuming Tab. It conflicts with
  the loom Tab policy (containers move focus). Decide the policy before copying the idea.

**Table (P1).** `table.go` L267 renders only the rows around the cursor (`cursor ± viewport.Height`) and caps the
rendered rows at 2 × viewport height (L269). This is the virtualisation the ticket asks for. Missing from bubbles:
sorting, filtering, and column resize (none found in the file). Loom Table/Choice should match the virtualisation and
can go beyond bubbles on sort and filter.

**List filtering (P1).** `list.DefaultFilter` (`list.go` L98) is fuzzy with ranking by `sahilm/fuzzy`.
`UnsortedFilter` (L111) keeps the original order. Filter state is explicit: `FilterState` (L128) with
`SetFilterText` (L280), `SetFilterState` (L295), `SetFilteringEnabled` (L258), `SetShowFilter` (L314).
`FilterInput` is an exported `textinput.Model` (L185). `SetFilterInput` was not found in the part of the file read.

**Textarea growth and cursor.** `DynamicHeight` (`textarea.go` L321) grows and shrinks the box between `MinHeight` and
`MaxHeight` (L312). The cursor can be virtual or real: `SetVirtualCursor` (L496), default virtual (`New`, L164),
and `Cursor() *tea.Cursor` (L1751) for the real cursor. Loom TextEntry auto-size and cursor placement can be compared
against this. The fetched comment says upstream plans to make max lines dynamic in v2. The line was not checked.

**Textinput.** Single-line features: `Validate` (L141), `EchoMode` with normal, password and none (L96), suggestions
via `SetSuggestions` (L287) with `ShowSuggestions` (L147), and next/prev bound to down/ctrl+n and up/ctrl+p.

**Filepicker.** `AllowedTypes` (L138), `DirAllowed`/`FileAllowed` (L145-L146), `ShowSize` (L143, default true in `New`
L35), size column 7 cells wide (`fileSizeWidth`, L62). Loom FilePicker can be compared on these knobs.

**Tree (new, P2).** Bubbles has a `tree` package (`tree/tree.go` L1: "Package tree provides a tree component for
Bubble Tea"). It has `SetScrollOff` (L222), `SetOpenCharacter`/`SetClosedCharacter`/`SetCursorCharacter` (L227-L237),
and `SetNodes` (L241). Compare directly with loom `Tree`.

## Huh: findings against loom

**Select filter is substring, not fuzzy (P1).** `field_select.go` L751 is
`strings.Contains(strings.ToLower(option), strings.ToLower(s.filter.Value()))`. The filter is typed with `/`
(doc comment L29), and the filter field is a bubbles `textinput.Model` (L49). This is weaker than bubbles
`list.DefaultFilter` (fuzzy with ranking). Loom's fuzzy ranking item should use the bubbles list model as a reference.

**MultiSelect.** `Limit` (`field_multiselect.go` L190) caps selections. Select-all and select-none exist (L436) and
are turned off when a limit is set (L719-L729). Filterable is on by default (fetch report). Loom Choice or Table
multi-select should check the limit and select-all behaviour.

**Input and Text.** `Input.Suggestions` (`field_input.go` L149) and `SuggestionsFunc` (L166); `Input.Validate` (L231).
`Text.ExternalEditor` (`field_text.go` L191) opens `$EDITOR` through `tea.ExecProcess` on a temp file (fetch report).
Default editor is `nano` (L196) when `EDITOR` is unset. Loom's external-editor support was not checked (host to verify in loom code).

**Form layouts (P2, composition).** `layout.go` has `LayoutDefault` (L16, one group at a time), `LayoutStack` (L19),
`LayoutColumns(columns)` (L22) and `LayoutGrid(rows, columns)` (L27). They are set with `Form.WithLayout` (`form.go`
L365). This is group composition only. It is not the cell-level composition that the astra claim-cell model does.

**Accessibility (P2).** `Form.WithAccessible` (`form.go` L235) switches to standard prompts. `TERM=dumb` also turns
accessible mode on and sets width 80 (`form.go` L129). The README recommends an `ACCESSIBLE` environment variable.
`accessibility.go` returned 404 in the fetch, so the file was not confirmed.

**Themes (P2, dark/light).** `type ThemeFunc func(isDark bool) *Styles` (`theme.go` L15). The built-in themes take the
same `isDark` flag (`ThemeCharm` L139, `ThemeDracula`, `ThemeBase16`, `ThemeCatppuccin`). The library receives the
dark-background signal from the program. Loom's dark/light detection item can use this shape as a reference.

**Others.** `Note` field is display only (`field_note.go` L55, with `TitleFunc` L68). `FilePicker` in huh has
`AllowedTypes` (`field_filepicker.go` L144) and `ShowSize` (L89, default off).

## Notes on Huh and Bubbles v2 changes

- Huh v2 imports `charm.land/bubbles/v2`, `charm.land/bubbletea/v2`, and `charm.land/huh/v2/internal/...` (fetch report).
  The README import is `charm.land/huh/v2`.
- The search pages say that v2 moved getters and setters onto `Table`, `Filepicker`, `Help`, `Progress`, `Textinput`
  and `Viewport`. `SetHeight` exists on `table.go` L329, and `Height()` is on `filepicker`. This is partly
  verified. The full migration list was not read.
- Light and dark styles for help, list, textarea and textinput were in a search summary. Not verified: the grep for
  `isDark|HasDarkBackground` in textarea, textinput and list found only cursor fields.
- The search summary said textarea and textinput now use real terminal cursors. The `Cursor()` and
  `SetVirtualCursor` methods above show this is possible. Whether it is the default was not checked (the default is virtual).

## Gaps and what would close them

1. Not read: bubbles `viewport`, `progress`, `paginator`, `spinner`, `timer`, `stopwatch`, `help`, `key`, `cursor`.
   Close: fetch each `*.go` file on `main` and grep for exported symbols.
2. Not read: `UPGRADE_GUIDE_V2.md`. Close: fetch the file and list every removed or renamed symbol.
3. Huh `Spinner` package and `Form` `Run` details. The README names a spinner section, but the spinner source was not read.
4. Huh `keymap` file (key characters for select and multi-select). Close: fetch the keymap file in `huh`.
5. `accessibility.go` in huh returned 404. Close: search the huh repo tree for `accessible` or read `form.go` lines
   around `RunWithContext` (L706).
6. Bubbles `list` after the first ~1000 lines (`SetFilterInput`, pagination details). Close: fetch the raw file.
7. Bubbles textarea `MaxHeight` TODO (fetched comment, line unknown). Close: grep `XXX` in raw `textarea.go`.
8. Loom side: the `loom` column is unknown. The host should check `spec/widgets.yaml` for Form layouts, Select filter,
   TextEntry undo and external editor, and Table virtualisation.

## Rows written and verified

- Rows written: 38 (`features-bubbles.yaml`).
- Rows verified: 38, each with a pinned source URL and a quote taken from raw `main` files. Line numbers are from `grep -n`
  on the raw file at research time.
- Open gaps: 8 (listed above).
