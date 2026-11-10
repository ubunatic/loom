# Widgets and editing follow-up (batch 6)

Scope: loom widgets for Table, Choice, FilePicker, TextInput, TextArea, TextEntry, RichTextEdit, Form and Pane.Suspend, compared with the latest Bubbles, Huh, tview and gocui releases. Plan: `followup-plan.md` section 6. Research only: no loom code changed.

## Pins and scope

Loom revision for citations: `c96fbf66412240973abbf56dda9bedab3414d059`. Loom links use `https://codeberg.org/ubunatic/loom/src/commit/c96fbf66412240973abbf56dda9bedab3414d059/<file>#Lx-Ly`.

Upstream pins are the latest release tags, re-resolved on 2026-11-10 with `git ls-remote` and `gh api`. Commit SHAs are peeled from the tags. The plan's HEAD pins are kept for reference.

| Alias | Latest release | Commit SHA | Plan HEAD pin (moved) |
|---|---|---|---|
| B (Bubbles, `charm.land/bubbles/v2`) | v2.2.1 | `490948109eb4731927ba2f4cd464d8175a9630d7` | `1deada954e6eead9c054837f1cd3d669e912cd4d` |
| H (Huh, `charm.land/huh/v2`) | v2.0.3 | `3c0116c6fd1f578ef50a25d868d2db19bd985222` | `ffb6a97195b4a031925a31cc70c62fc1d309e69d` |
| TV (tview) | v0.42.0 | `5ce6a2b588145610060000a4f75d7e2af081a794` | `c15b79fa47f27032f26647a75cb5a81720255076` |
| GC (gocui) | v0.4 | `a319843434acb2446b85fdcaac286d45dd9e27ca` | `c81715e95462cf95447bbd619e4dc85f3e4e1006` |

Shallow clones at these tags are in the scratchpad (`b6-src/`). Permalinks use `https://github.com/<org>/<repo>/blob/<SHA>/<file>#Lx-Ly`.

Tab policy (`docs/Widgets.md:326-331`): Tab belongs to containers (Split, Form, FilePicker); a widget consumes Tab only when Tab is its documented action.

Not run: Bubbles textinput suggestions (source only); Huh form and field tests (not read); tview and gocui canaries (source only). The `sahilm/fuzzy` tie rule used by `list.DefaultFilter` was not read.

## Loom: per-widget findings (code)

**Table** (`table.go`)
- Draw loops only over the viewport: `for row := 0; row < itemRows; row++` at `table.go:404`, rows taken from `t.filtered[t.viewOffset+row]`.
- Whole-data work is elsewhere. `SetRows` (`table.go:212-219`) calls `refilter` (`table.go:241-261`), which calls `computeWidths` (`table.go:265-283`). `computeWidths` scans every cell of every row when a column has no fixed width.
- Each query edit (`table.go:582` `t.query += e.Text`, then `refilter`) scans all cells with `strings.Contains` on lowercased text (`table.go:241-261`). The filtered slice is rebuilt on every keystroke.
- Sort: the Table never sorts. `cycleSortNext` (`table.go:627-639`) and `sort_flip` (`table.go:567`) only call `OnSort`. The app must sort and call `SetRows`. The string `sort.` does not occur in `table.go`.
- `!` is intercepted by `sort_flip` before the text branch (`table.go:567`), so `!` cannot be typed into the filter.
- Backspace removes one rune (`table.go:573-580` pattern: `runes[:len-1]`), not a grapheme.
- Tab: `sort_cycle` (`table.go:561-564`, spec `spec/defaults.yaml:364`). Documented action, so this conforms to the Tab policy.

**Choice** (`choice.go`)
- `refilter` (`choice.go:245-283`): empty query keeps `c.Items` (no copy). With `c.Fuzzy`, each item is scored on Name and Desc with `fuzzyScore`, then `sort.SliceStable` by score descending (`choice.go:266`). Equal scores keep input order. Without Fuzzy, a lowercased substring test on Name or Desc.
- `fuzzyScore` (`choice.go:289-313`): greedy first-match subsequence. +1 per matched rune, +4 at a word start, +5 for an adjacent match. It returns on the first full match, so it does not search for the best alignment.
- Draw calls `fuzzyScore` again for highlighting (`choice.go:695`).
- Tab: no `"tab"` key case in `choice.go` (grep), so Tab is not consumed.

**FilePicker** (`filepicker.go`)
- In save mode, Tab and Shift+Tab toggle focus between the name field and the list (`filepicker.go:284-287`). This is a container role, which the Tab policy allows (`Widgets.md:327`).
- `OnSave` validates the destination (`filepicker.go:272-275`); an error keeps the dialog open (comment `filepicker.go:35-37`).

**TextInput** (`textinput.go`)
- Editing: runes, caret by grapheme (`measure.PrevCluster`/`NextCluster`). ConsumeKey covers left, right, home, end, backspace, delete and printable text (`textinput.go:77-155`). No shift-arrow selection, no mouse handler (no `ConsumeMouse` in the file), no undo, no cut, no Ctrl+V.
- Copy (`textinput.go:78-91`): the `copy` binding writes the selection, or the whole value when there is no selection (`start == end` branch). With a Mask set, copy returns `Ignored()`.
- Mask: `Mask rune` (`textinput.go:25`); display via `displayRunes` (`textinput.go:238-243`, one mask rune per stored rune); the cursor is drawn at `StringWidth(strings.Repeat(mask, caret))` (`textinput.go:229-231`). `Value()` returns the stored text (`textinput.go:52`).
- Paste: `ConsumePaste` (`textinput.go:55`).
- No validation hook, no suggestions, no $EDITOR hand-off.

**TextArea** (`textarea.go`)
- Editing: lines and caret (row, col); key switch `textarea.go:240-343`. No shift-selection cases, no `ConsumeMouse`, no undo, no cut. Tab has no case, so it is ignored and moves focus.
- Copy (`textarea.go:241-251`): selection, or the whole value when there is no selection.
- Paste: `ConsumePaste` (`textarea.go:82`). Enter inserts a newline here (`enter` case, `textarea.go:329-331`).
- No mask, no validation, no suggestions, no $EDITOR.

**TextEntry** (`text_entry.go`, with `text_entry_bold.go`, `text_entry_frame.go`)
- Units: `textEntryUnits` (`text_entry.go:1103-1113`) normalises tabs to four spaces, strips control runes, then uses `measure.Clusters`. This is the source of the emoji split (see canary C4 and ticket 433).
- Keyboard selection: shift-prefixed keys extend (`text_entry.go:483`, `extend := strings.HasPrefix(name, "shift-")`); mouse selection in `ConsumeMouse` (`text_entry.go:561-646`: press places caret, drag extends, double click word, triple click line).
- Copy: `ctrl-c` with a selection is `Consumed`; without one it is `Ignored()` so the key passes through (`text_entry.go:532-537`).
- Cut (`text_entry.go:538-544`), internal Ctrl+V (`text_entry.go:545-549`), Ctrl+Z undo (`text_entry.go:550-552`), Ctrl+Y and Ctrl+Shift+Z redo (`text_entry.go:553`).
- Undo: `change` (`text_entry.go:976-992`) merges a typed run into one step (`typingRun`); `pushUndo` caps history at `RichTextEdit.UndoLimit` (`text_entry.go:994-999`); `undoStep`/`redoStep` at `text_entry.go:1001-1008`.
- Clipboard writes go to the host hook (`setClipboard`, `text_entry.go:1032-1038`; `setClipboardWriter` at `text_entry.go:350`, unexported; Pane binds it to OSC 52 through the `clipboardSink` interface in `bindClipboardTree`, `clipboard.go:22-35`, also for TextInput and TextArea).
- Tab (`text_entry.go:467`): consumed only when the caret and anchor are at 0 and `AcceptPlaceholder` (`text_entry.go:223-241`) inserts the placeholder. Otherwise Tab is ignored. `end` and `ctrl-e` also accept the placeholder there.
- Enter submits and Shift+Enter inserts a newline by default (`text_entry.go:56-57`, `SubmitKeys`/`NewlineKeys` at `text_entry.go:33-35`).
- Paste: `ConsumePaste` (`text_entry.go:353`), sanitised through `textEntryUnits`.
- No mask, no validation hook, no suggestions list (ticket 463 is a Draft), no $EDITOR.

**RichTextEdit** (`richtextedit_input.go`, `richtextedit_clipboard.go`, `richtextedit_mouse.go`)
- Tab (`richtextedit_input.go:131`): toggles italic only when there is a selection (`... || (key.Is(tab) && e.HasSelection)`). Otherwise it falls through to `default:` with empty Text and returns `Ignored()`, so focus can move. Documented in the help text (`spec/defaults.yaml:598`, "Italic (selected text)").
- Copy (`richtextedit_input.go:154-160`) calls `copySelectionOrWord` (`richtextedit_clipboard.go:9-14`): with no selection it copies the word at the caret.
- Cut (`richtextedit_input.go:161-169`) deletes what was copied, Paste (`:170-171`) inserts the internal clipboard (`e.clipboard`). Undo `:172-173`, redo follows.
- No mask, no validation, no suggestions (the popover is a style and hotkey menu, not completion), no $EDITOR.

**Form** (`form.go`)
- Key handling (`form.go:271-306`): before any field sees a key, the form checks its bindings. Spec `form.bindings` (`spec/defaults.yaml:405-410`): `next` = tab, down, pgdown, pgdn, pagedown; `previous` = shift-tab, up, pgup, pageup; `submit` = enter. So Down, PgDn and Enter never reach a TextArea inside a Form.
- `activate` (`form.go:220-262`): Enter runs validation (`validate`, `form.go:159-177`, `Required` plus `FormField.Validate`) and then `submit` (`form.go:264-268`, calls `OnSubmit`). Validation runs at submit only, not per edit.
- `formValue` (`form.go:179-196`) reads `Value()`, which is the stored value for a masked TextInput.

**Pane.Suspend** (`pane_suspend.go:166-208`)
- Generic API: halts the input reader, leaves the alternate screen, disables mouse and paste modes, restores cooked mode, runs `fn`, then restores raw mode and modes, re-reads the size and repaints (comment `pane_suspend.go:149-165`). It preserves unread input and drops Ctrl+C that reached the pane during `fn`.
- No widget calls it. The only caller is the example `examples/suspend/suspend/suspend.go:63`, which runs `$VISUAL` or `$EDITOR` (`suspend.go:76-82`). The widget's state is not touched by Suspend, so restore is by the app.

## Per-widget editing matrix (loom)

Cells: **has** = implemented and reachable from the key or mouse path, **partial** = limited (reason given), **missing** = not present. "n/a" = not an editor. Citations are in the text above.

| Feature | TextInput | TextArea | TextEntry | RichTextEdit | Form | Choice / Table / FilePicker |
|---|---|---|---|---|---|---|
| Keyboard selection (Shift+arrows) | missing (`textinput.go:77-155`) | missing (`textarea.go:240-343`) | has (`text_entry.go:483`) | has (move/select keys, `richtextedit_input.go`) | n/a | n/a |
| Mouse selection | missing (no `ConsumeMouse`) | missing (no `ConsumeMouse`) | has (`text_entry.go:561-646`) | has (`richtextedit_mouse.go`, not read in detail) | n/a | Choice click selects (`choice.go:533`); Table click (`table.go:603`) |
| Ctrl+C with selection | has (`textinput.go:78-91`) | has (`textarea.go:241-251`) | has (`text_entry.go:532-537`) | has (`richtextedit_input.go:154-160`) | n/a | n/a |
| Ctrl+C without selection | partial: copies whole value (`textinput.go:82-90`) | partial: copies whole value (`textarea.go:248-250`) | has: passes through, `Ignored` (`text_entry.go:533-534`) | partial: copies word at caret (`richtextedit_clipboard.go:9-14`) | n/a | n/a |
| Cut (Ctrl+X) | missing | missing | has (`text_entry.go:538-544`) | has (`richtextedit_input.go:161-169`) | n/a | n/a |
| Paste (Ctrl+V / paste event) | partial: paste event only (`textinput.go:55`) | partial: paste event only (`textarea.go:82`) | has: paste event + internal Ctrl+V (`text_entry.go:353`, `:545-549`) | partial: internal Ctrl+V only | n/a | n/a |
| Undo / redo | missing | missing | has (`text_entry.go:976-1025`) | has (`richtextedit_input.go:172-173`) | n/a | n/a |
| Tab key | conforms: ignored (`textinput.go:77-155`) | conforms: ignored (`textarea.go:240-343`) | partial: accepts placeholder when empty only (`text_entry.go:467`) | partial: italic when selected, else ignored (`richtextedit_input.go:131`) | container: next field (`form.go:271-279`, spec `:407`) | Choice conforms (no case); Table: sort cycle, documented (`table.go:561`); FilePicker: name/list toggle, container role (`filepicker.go:284`) |
| Validation | missing (no hook) | missing | missing | missing | has at submit only (`form.go:159-177`, `:257`) | n/a |
| Password mask | has (`textinput.go:25`, `:238-243`); copy refused when masked (`:78-81`); `Value()` stored | missing | missing | n/a | `formValue` reads stored value (`form.go:179-196`) | n/a |
| Suggestions / completion | missing (ticket 463 covers TextEntry, Draft) | missing | partial: placeholder only (`text_entry.go:223-243`); popup missing (463) | missing (popover is style/hotkey menu) | n/a | Choice filter is not completion |
| $EDITOR hand-off | missing | missing | missing | missing | n/a | n/a (Pane.Suspend API exists; only the example uses it, `examples/suspend/suspend/suspend.go:63`) |
| Grapheme-aware caret | partial: `PrevCluster`/`NextCluster` used, but `measure` splits modifiers (canary C4) | not checked | partial: `measure.Clusters` splits modifiers (canary C4) | not checked | n/a | n/a |
| Stored vs displayed value | same for unmasked; masked: stored runes, displayed mask per rune | same | same | same | `Value()` | Table: display = cells |

## Canary results

Harness: `scratchpad/b6-canary` (module `b6canary`, `replace ubunatic.com/loom => /home/uwe/projects/loom`, `replace charm.land/bubbles/v2 => scratchpad/b6-src/bubbles`). No sleeps; each test enforces a 5 s deadline. Benchmark host: 12 threads (`-12` suffix).

Commands:
- Correctness: `cd scratchpad/b6-canary && go test -count=1 -v -run 'Test' .`
- Benchmarks: `go test -count=1 -run XXX -bench 'Table' -benchmem -benchtime=1s .`
- Segmenter switch: `LOOM_FAST_MEASURE=0 go test -count=1 -v -run 'TestTextEntryGraphemes|TestClusterSplit' .`

Measurement boundaries (named): loom = `Table.Draw` into a `Canvas` (flush not included); Bubbles = `Table.View()` string (lipgloss output, no terminal flush). Loom `SetRows` includes `computeWidths`; `computeWidths` is unexported, so the width cost is the `SetRows` figure. Filter = one `ConsumeKey` text edit, timed as a pair with Backspace; per-edit cost is half the pair.

**Table: 80x24 viewport, 3 columns**

| Case | ns/op | B/op | allocs/op | Verdict |
|---|---|---|---|---|
| Loom Draw, 100 rows | 90,006 | 12,856 | 156 | |
| Loom Draw, 100,000 rows | 91,644 | 14,072 | 163 | constant: proved |
| Loom SetRows, 100 rows | 3,140 | 24 | 1 | |
| Loom SetRows, 100,000 rows | 3,447,718 | 24 | 1 | linear (width scan): proved |
| Loom filter pair, 100 rows | 14,209 (about 7.1 µs per edit) | 4,499 | 6 | |
| Loom filter pair, 100,000 rows | 13,386,956 (about 6.7 ms per edit) | 4,006,291 | 6 | linear: proved |
| Bubbles View, 100 rows | 116,787 | 20,713 | 869 | |
| Bubbles View, 100,000 rows | 115,573 | 20,420 | 869 | constant: proved |
| Bubbles SetRows, 100 rows | 373,334 | 19,466 | 805 | |
| Bubbles SetRows, 100,000 rows | 598,631 | 1,623,308 | 805 | O(N) bytes per call: proved (see B2) |
| Bubbles filter | n/a: no filter in `table/table.go` (grep count 0) | | | missing upstream |

**Choice fuzzy rank and ties, query "alp", same word list in both**

Word list: `alphabet, a-l-p-h-a, xalpx, alp, plate, lap, Alpha Beta, bb alp, aa alp, alpine, cap, a l p`.

- Loom `Choice` (`Fuzzy = true`): `["alphabet" "alp" "Alpha Beta" "bb alp" "alpine" "a-l-p-h-a" "a l p" "xalpx" "aa alp"]`
- Bubbles `list.DefaultFilter`: `["a l p" "a-l-p-h-a" "alp" "alpine" "alphabet" "Alpha Beta" "bb alp" "aa alp" "xalpx"]`

Ties: both keep input order among equal scores in the cases shown ("bb alp" before "aa alp" on both sides). The rank order differs: loom puts "alphabet" first and "a l p" low, Bubbles puts "a l p" first. Loom's greedy scoring explains the first; the Bubbles tie and score rule was not read (`sahilm/fuzzy` not fetched), so the cause of the second is unverified.

**TextEntry: grapheme selection and undo, `a👍🏽b` and `é`**

| Input | caret at end (bytes) | Shift+Left then Backspace | Undo | Redo | Plain Left from end |
|---|---|---|---|---|---|
| `a👍🏽b` | 10 | `a👍🏽` (deletes `b`) | `a👍🏽b` | `a👍🏽` | caret 9 (before `b`) |
| `é` as `e` + U+0301 | 3 | `""` (whole cluster) | `é` | `""` | caret 0 |
| `é` precomposed | 2 | `""` (whole cluster) | `é` | `""` | caret 0 |

Earlier run, Shift+Left started from before `b` (caret 9): Backspace removed only the skin-tone modifier, giving `a👍b`. Cause: `measure.Clusters` returns four units `["a" "👍" "🏽" "b"]` (`cluster_test`), with `LOOM_FAST_MEASURE=1` and `=0` alike. TextInput shows the same split: two Left presses from the end of `a👍🏽b` leave the caret at rune 2 (of 4), not 1. Verdict for skin-tone sequences: **disproved** (split). This is ticket 433 (Open), and ticket 432 (closed) names it. Verdict for decomposed and precomposed `é`: **proved** (one cluster each).

Undo of a typed run: typing `xyz` and one Ctrl+Z gives `""` (the run is one step). **Proved.**

**Ctrl+C with and without selection**

| Widget | No selection | With selection | Verdict |
|---|---|---|---|
| TextEntry | consumed=false (key passes through) | consumed=true | proved |
| TextInput (unmasked) | consumed=true, copies whole value `secret` | not run | partial: copies whole field |
| TextInput (masked `•`) | consumed=false, copies nothing | not run | proved: masked copy refused |

**Tab in TextEntry**

- With placeholder `hint`, empty: Tab 1 consumed=true, value `hint`; Tab 2 consumed=false.
- Without placeholder: Tab consumed=false. **Proved** (`text_entry.go:467`).

**Masked TextInput and Bubbles password mask**

- Loom TextInput with `Mask='•'`, value `secret`: display `••••••`, `Value()` = `secret`, copy refused. **Proved.**
- Bubbles `EchoPassword`, value `secret`: View contains `••••••`, does not contain `secret`, `Value()` = `secret`. **Proved** for both.
- TextEntry and TextArea have no mask (grep: no `Mask` in those files). **Proved** (missing).

**Form and multi-line TextArea**

- Field 0 holds a TextArea with `one\ntwo`, caret at (0,0). Down moves focus to field 1 and the TextArea caret stays on row 0. Enter submits (`OnSubmit` count 1), and the value stays `one\ntwo`. **Proved**: Form captures Down and Enter before the TextArea.

**Table Tab**: `ConsumeKey(tab)` consumed=true and `OnSort` called with column 0. **Proved** (documented sort action).

## Upstream mechanisms (quotes at pinned SHAs)

**U1. Bubbles table: visible rows only, but the row slice is sized by all rows.** `table/table.go` at 490948109eb4731927ba2f4cd464d8175a9630d7:

> `renderedRows := make([]string, 0, len(m.rows))`
> `// Render only rows from: m.cursor-m.viewport.Height to: m.cursor+m.viewport.Height`
> `// Constant runtime, independent of number of rows in a table.`
> `for i := m.start; i < m.end; i++ {`

Permalink: https://github.com/charmbracelet/bubbles/blob/490948109eb4731927ba2f4cd464d8175a9630d7/table/table.go#L265-L277. Annotation: the loop is viewport-bounded (measured flat). The `make` call is O(rows) per update (measured 1.6 MB per call at 100k rows). The comment's "constant runtime" holds for CPU, not allocation.

**U2. Bubbles list ranks with sahilm/fuzzy and a stable sort.** `list/list.go` at the same SHA:

> `ranks := fuzzy.Find(term, targets)`
> `sort.Stable(ranks)`

Permalink: https://github.com/charmbracelet/bubbles/blob/490948109eb4731927ba2f4cd464d8175a9630d7/list/list.go#L98-L100. Annotation: ranking is delegated to `sahilm/fuzzy`; ties follow that library's `Less`, not read here.

**U3. Bubbles textinput: Tab accepts a suggestion.** `textinput/textinput.go`:

> `AcceptSuggestion:        key.NewBinding(key.WithKeys("tab")),`
> `func (m *Model) canAcceptSuggestion() bool {`
> `return len(m.matchedSuggestions) > 0`

Permalink: https://github.com/charmbracelet/bubbles/blob/490948109eb4731927ba2f4cd464d8175a9630d7/textinput/textinput.go#L83 and `#L850-L852`. Annotation: Tab is a documented suggestion action (a Tab-policy exception that loom does not have). Source only; not run.

**U4. Bubbles password echo is per display width.** `textinput/textinput.go` `#L566-L569`:

> `func (m Model) echoTransform(v string) string {`
> `case EchoPassword:`
> `return strings.Repeat(string(m.EchoCharacter), uniseg.StringWidth(v))`

Annotation: the mask count follows display width (uniseg), not rune count. Loom's TextInput masks one mask rune per stored rune (`textinput.go:238-243`), so a wide or multi-rune cluster differs. Not canaried.

**U5. Bubbles validation runs on every SetValue and edit.** `textinput/textinput.go#L214-L215`:

> `err := m.validate(runes)`
> `m.setValueInternal(runes, err)`

Annotation: validation is per value, not only at submit. Loom validates only in `Form.validate` (`form.go:159-177`).

**U6. Huh text field validates on blur, Huh input on every change.** `huh/field_text.go#L243-L247`:

> `t.accessor.Set(t.textarea.Value())`
> `t.err = t.validate(t.accessor.Get())`

Annotation: Huh's Text field validates at blur (`field_text.go:243`). Huh's Input validates through Bubbles (U5).

**U7. Huh external editor: `$EDITOR` through `tea.ExecProcess`.** `huh/field_text.go`:

> `editor := strings.Fields(os.Getenv("EDITOR"))`
> `const defaultEditor = "nano"`
> `cmds = append(cmds, tea.ExecProcess(cmd, func(error) tea.Msg {`

Permalinks: `#L199-L204` (getEditor), `#L196` (default), `#L336` (exec). Annotation: the editor runs on a temp file with the text, and the content is read back on exit. Loom has no equivalent widget; the closest is `Pane.Suspend`.

**U8. Huh Next keys include tab and enter.** `huh/keymap.go#L113`:

> `Next:             key.NewBinding(key.WithKeys("enter", "tab"), key.WithHelp("enter", "next")),`

Annotation: the keymap block's name was not checked. Enter and Tab move between fields in that block. Not canaried.

**U9. Huh input validation hook.** `huh/field_input.go#L231`:

> `func (i *Input) Validate(validate func(string) error) *Input {`

Annotation: a single function per field. Loom's FormField has the same shape (`form.go:13-19`) but calls it only at submit.

**U10. tview textarea: Tab inserts unless a form takes it; Ctrl-Q copies.** `tview/textarea.go#L2119-L2122`:

> `case tcell.KeyTab: // Insert a tab character. It will be rendered as TabSize spaces.`
> `// But forwarding takes precedence.`
> `if t.finished != nil {`

Permalink: https://github.com/rivo/tview/blob/5ce6a2b588145610060000a4f75d7e2af081a794/textarea.go#L2119-L2122. Copy: `#L2269-L2272` (`case tcell.KeyCtrlQ: // Copy to clipboard.`); undo `#L2291` (`case tcell.KeyCtrlZ: // Undo.`), redo `#L2313`. Annotation: tview has undo, cut and copy in the textarea, and keeps Tab for forms when a finish handler is set. Ctrl-C is left to the app.

**U11. gocui simple editor: Tab inserts a newline.** `gocui/edit.go#L48-L49`:

> `case key == KeyTab:`
> `v.EditNewLine()`

Permalink: https://github.com/jesseduffield/gocui/blob/a319843434acb2446b85fdcaac286d45dd9e27ca/edit.go#L48-L49. Annotation: gocui's default editor takes Tab as a text key, with no focus role.

**U12. Bubbles textarea: system-clipboard copy, no undo.** `textarea/textarea.go`:

> `case key.Matches(msg, m.KeyMap.CopySelection):`
> `func (m *Model) CopySelection() tea.Cmd {`
> `if !m.HasSelection() {`

Permalinks: `#L125` (binding `ctrl+shift+c`), `#L1428-L1429`, `#L1829-L1831`. Annotation: copy needs a selection and uses `clipboard.WriteAll`. The file has no undo or redo (grep count 0).

## Loom vs upstream comparison

| Question | Loom | Upstream | Verdict |
|---|---|---|---|
| Table draw cost vs rows | flat, 90 µs at 100 and 100k | flat CPU, 116 µs (Bubbles View) | equal: proved |
| Table rebuild cost vs rows | SetRows and filter linear (3.4 ms, 6.7 ms at 100k) | SetRows linear allocation (1.6 MB at 100k, 0.6 ms) | both linear; different cost |
| Table filter | yes, all cells, rebuilt per key | none (Bubbles table has no filter) | loom ahead |
| Table sort | none; callback only | not checked (Bubbles sort not grepped) | unknown |
| Fuzzy ranking | greedy subsequence, score tie = input order | sahilm/fuzzy, stable sort | different order; tie rule not verified upstream |
| Text undo | TextEntry and RichTextEdit | Bubbles textarea none; tview textarea yes | loom partial; tview ahead |
| Text copy / cut | TextEntry, RichTextEdit; TextInput copy only; TextArea copy only | Bubbles textarea copy (system clipboard); tview cut and copy | loom writes the system clipboard by OSC 52 (Pane `bindClipboardTree`, `clipboard.go:22-35`), never reads it; paste is the terminal's bracketed paste or the internal clip (465, 466) |
| Mouse selection | TextEntry and RichTextEdit | none in Bubbles textarea grep (not checked fully); tview not checked | loom ahead on TextEntry |
| Password mask | TextInput only, stored value kept, copy refused | Bubbles EchoPassword, width-based mask | parity for basic mask; loom per-rune mask |
| Validation | Form at submit | Bubbles per value, Huh per blur | loom behind |
| Suggestions | none (463 Draft) | Bubbles Tab accepts suggestion | loom behind |
| $EDITOR | none in widgets; Pane.Suspend API | Huh Text via ExecProcess | loom behind (no widget) |
| Tab policy | matches on TextInput, TextArea, Table, Choice; RichTextEdit and TextEntry conditional | Bubbles: Tab accepts suggestion; tview: Tab inserts unless form; gocui: Tab inserts newline | loom matches the Tab policy |
| Form Enter / Down in TextArea | captured by Form | Huh Enter next (U8) | loom defect |

## Stage 4 proposals

Proposals only: the integrator applies them. Each row cites the evidence above.

**features.md, widgets**

- `Table renders only visible rows` (`bubbles.table.virtualised-render`): change `unchecked` to `partial`. Draw is viewport-bounded (proved, 90 µs flat); SetRows, width and filter scan all rows (proved, 3.4 ms and 6.7 ms per edit at 100k). Note the Bubbles row slice is also O(N) in allocation (U1).
- New row, Tab in TextEntry: `Tab accepts a placeholder when empty` (`loom`): proved (`text_entry.go:467`, canary). Status `proved`. Compare Bubbles U3 (Tab accepts a suggestion, source only).
- New row, Form keys: `Form captures Down and Enter before a TextArea`: proved (canary); status `proved` as a defect. The Tab part (`Form captures Tab`) is proved by code only (`form.go:271-279`, spec `:407`).
- Paste row: none exists. Propose `Text widgets accept paste events; TextEntry sanitises control runes`: proved (`textinput.go:55`, `textarea.go:82`, `text_entry.go:353`, `:1103-1113`). Overlaps 464 (paste sanitiser, Open), 181 (closed, decode paste), 465 (clipboard), 466 (image paste).

**summary.md, line 22 ("desktop-like" qualifier)**

- Current: "no undo/redo in text inputs; select-all on Ctrl+G; Tab accepts suggestions; table has no sort or filter".
- Proposed: "undo/redo is in TextEntry and RichTextEdit, not TextInput or TextArea (`text_entry.go:994-1025`, `richtextedit_input.go:172`). Select-all is Ctrl+A in TextEntry; no Ctrl+G select-all exists in the loom text widgets (grep: none). Tab accepts suggestions in Bubbles only; loom TextEntry accepts a placeholder. Table has a filter (all cells) and a sort callback, but no sort of its own."
- Qualify "desktop-like" for loom: mouse selection, copy, cut, paste and undo exist in TextEntry and RichTextEdit only. TextInput and TextArea have no mouse or keyboard selection, no undo and no cut. The text inputs are not uniformly desktop-like.

**Spec note (`spec/defaults.yaml:40`, `:53-56`)**: no change. "undo/redo are not provided by these widgets" and "editors have no undo stack" are scoped by `applies: "TextInput, TextArea"` (lines 39, 55) and are correct (host review). In TextArea, Ctrl+A moves to the line start (`textarea.go:303-305`), not select-all.

**Ticket candidates** (none filed; searches with `harnez find -d . issues "..."`):

1. **Form captures Enter, Down and PgDn inside a TextArea** (P2 candidate). Evidence: canary `submits=1`, value unchanged; code `form.go:271-296`, spec `:407-410`. Overlap: none found by "textarea form enter". Related to 169 (closed Form layout).
2. **TextInput Ctrl+C without selection copies the whole value; TextEntry passes through** (P3 candidate). Evidence: canary; `textinput.go:82-90` vs `text_entry.go:533-534`. Overlap: 465 (clipboard). Could be folded there.
3. **TextInput and TextArea have no keyboard or mouse selection, undo or cut** (P2 candidate, owner decision). Evidence: `textinput.go:77-155`, `textarea.go:240-343`, no `ConsumeMouse` in either file. Overlap: 467 (drag, for mouse selection), 464/465 for clipboard. Not a duplicate of 173 (password mask, closed).
4. **Table: SetRows, width and filter scan all rows per call** (P3 candidate). Evidence: canary 3.4 ms and 6.7 ms at 100k; `table.go:212-283`, `:241-261`. Overlap: none found; "table virtualisation" search returned no Table ticket.
5. **Emoji modifier split in measure.Clusters** (no new ticket). Ticket 433 (Open) already covers it. Attach this evidence (TextEntry and TextInput caret and backspace, canary C4) to 433. 432 (closed, TextInput and TextArea backspace) is related.
6. **Validation hook for TextInput and TextEntry, per edit or blur** (P3 candidate, owner decision). Evidence: U5, U6; loom only at submit (`form.go:159-177`). Overlap: 169 (closed).
7. **Choice fuzzy: greedy first match, not best alignment** (P3 research). Evidence: the rank table above; `choice.go:289-313`. Overlap: 463 (completion popup, Draft) for the suggestion side.
8. **Table filter: `!` cannot be typed; Backspace removes a rune** (P3). Evidence: `!` is taken by `sort_flip` at `table.go:567-572` before Backspace (`:573-580`) and text input (`:581-585`). Overlap: none.
9. **$EDITOR hand-off for TextArea** (P3, owner decision). Evidence: U7; only `examples/suspend` uses `Pane.Suspend`. Overlap: none found.

## Flow notes

- The Tab, Form and Table results are reproducible with the harness. The tests log values with `t.Logf`; re-run with `-v`.
- Two rows in the matrix are "not checked" (TextArea grapheme caret, RichTextEdit mouse selection in detail). They are listed as gaps, not as proofs.
- The emoji canary used a second test file (`cluster_test.go`) written through a Bash heredoc. A hook flagged it as an edit from Bash. The file is in the disposable harness only. Later files were written with the Write tool.
- The Bubbles textinput copy grep returned 2 hits that were not read; the Bubbles textinput copy claim is therefore not stated.
- The `go mod tidy` in the harness changed only the harness `go.mod`.
- Host review (agy:flash38:med plus host rerun of the canaries and reading `form.go`, `textinput.go`): claims 1-8 of the Stage 4 list confirmed; line numbers corrected (`text_entry.go:483`, `:223-241`, `table.go:241-261`, `:567-572`); the "internal clipboard only" and spec-note claims were wrong and are corrected above. Loom citations are file:line at the pinned revision, not instantiated Codeberg permalinks; build them from the template in "Pins and scope".
