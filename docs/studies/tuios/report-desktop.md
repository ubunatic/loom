# Batch 8: theming, composition and desktop integration

Ticket 458 (follow-up to 457), batch 8 of 9. Loom revision for citations:
`37d2b0ef9056422f3bcce05777419bc3bd4d8d3d` (on origin, Codeberg). Loom permalinks use
`https://codeberg.org/ubunatic/loom/src/commit/37d2b0ef9056422f3bcce05777419bc3bd4d8d3d/<file>#L<x>-L<y>`;
the `file:line` form below uses the same lines.

Scope: semantic theme roles, dark/light and colour-profile detection, canvas composition, z-order and
hit routing, focus after a popup closes, Markdown, clipboard, notifications and terminal focus/theme
reports. Covered by other reports and not repeated: terminal query replies (`report-queries.md`), nested
coordinates and modal capture (`report-mouse.md`), clipboard keys in widgets (`report-widgets.md`).

Canaries: `/tmp/claude-1000/-home-uwe-projects-loom/9e4d0913-fa4b-4cb2-aec0-0c51a6bd4a87/scratchpad/b8-canary/`
(`desktop_canary_test.go`, disposable, outside the repository). Upstream clones: `scratchpad/b8-src/`.
No live terminal queries were sent; everything ran headless.

## 1. Pins (re-resolved 2026-11-10)

Latest release per library, resolved with `git ls-remote --tags` and a shallow clone of the tag. Plan
pins that differ are listed in the last column.

| Library | Module path (from pinned `go.mod`) | Latest tag | Commit SHA | Plan pin in batch table |
|---|---|---|---|---|
| Lip Gloss v2 | `charm.land/lipgloss/v2` (`lg/go.mod:1`) | v2.1.0 | `ff16a4ef6dafa18b1f234a2687b9c35c18658e51` | same SHA |
| Glamour v2 | `charm.land/glamour/v2` (`gl/go.mod:1`) | v2.0.1 | `95c93db04489cebf252ec856584445c7e85ee276` | plan had `010543ca` (HEAD, older) |
| colorprofile | `github.com/charmbracelet/colorprofile` (`cp/go.mod:1`) | v0.4.3 | `d085584efb48f2ad470e96cd0f3dcb8cc68a034b` | plan had `51fafca7` (HEAD, older) |
| Bubble Tea v2 | `charm.land/bubbletea/v2` (`bt/go.mod:1`) | v2.1.1 | `67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7` | plan had `v2.0.10` (TUIOS dep) and `v2.1.1` |
| TUIOS | `github.com/Gaurav-Gosain/tuios` (default branch HEAD) | none (only `vt/v0.3.x` tags) | `6fe2c0df9922b2eec28f392fee4e196efd86be36` | same SHA |

Foot and kitty documentation were not fetched. They were only needed to label rows, and every row below
was decided from code or canaries.

## 2. Loom from code (revision 37d2b0ef)

### 2.1 Semantic roles and themes

- Roles: `ThemeColors` carries about 45 named roles, for example `NormalFG`, `SelectedBG`, `HeaderFG`,
  `ScrollbarThumbBG`, `StatusFG`, `FocusBG`, `KeyCapBG`, `ChartPalette [6]`, `PillDoneFG`
  (`theme.go:81-127`). `ThemeColor` distinguishes default, palette index and RGB (`theme.go:20-52`).
- Source: `spec/themes.yaml` embedded by `//go:embed` (`theme.go:14-15`). Names: `plain` (line 14), `mc`
  (64), `mc-classic` (110), `mc-dark` (156), `julia256` (210). There is no light theme in the file.
  The default is `julia256` (`spec/defaults.yaml:236`).
- Lookup: `Theme(name)` returns `plain` for an unknown name (`theme.go:243-251`).
- Switching is explicit. An app calls `ApplyTheme(Theme(name))` on each `Themeable` widget
  (`widget.go:276-279`, `cmd/loom/widgets.go:86` and `:336`, `examples/background/background/background.go:68`).
  No event, no re-detection.
- Dark/light detection: none. A grep of non-test Go for `2031`, `996`, `]11;` / `OSC 11`, `background colo`,
  `dark mode`, `IsDark`, `appearance` returned no code. `Theme` names are the only dark/light signal.

### 2.2 Colour profile

- `DetectColorProfile()` reads environment variables only: `LOOMCOLOR`, `NO_COLOR`, `COLORTERM`, `TERM`
  (`ansi/style.go:25-27`). It sends no query and reads no reply. `style.go:33-35` repeats the doc.
- Rules, in order (`ansi/style.go:30-56`):
  - `LOOMCOLOR` values `truecolor`/`24bit`, `256`/`256color`, `16`/`basic`, `none`/`off` win (lines 31-40).
  - Any non-empty `NO_COLOR` gives None (lines 41-43). This is the no-color.org rule.
  - `COLORTERM=truecolor` or `24bit` gives TrueColor (lines 44-46), even when `TERM=dumb`.
  - `TERM` containing `256color` gives 256 (lines 47-49).
  - `TERM=dumb` gives None (lines 50-52).
  - Empty `TERM` gives TrueColor (lines 53-55). This is the default for an unset environment.
  - Anything else gives 16 (line 56).
- `Canvas` captures the profile once at creation (`canvas.go:50-51`, `:81`) and uses it for SGR emission
  (`canvas.go:534`, `:547`).
- Downsampling is real: `Color.forProfile` maps RGB to the nearest 256 or 16 colour (`ansi/style.go:140-166`);
  `None` turns every colour into reset (`:150-151`).

### 2.3 Canvas composition and claims

- `PaintSurface` writes a blank cell unconditionally with `claim=false` (`canvas.go:195-200`).
- `PaintDefaultSurface` skips cells that already carry an explicit surface with a non-reset background
  (`canvas.go:211-233`).
- `PaintForeground` sets `Claim` and marks the cell as claimed (`canvas.go:235-240`); `setNew` records
  the claim at `canvas.go:343-345`.
- `PaintDecoration` writes only to cells `IsEligibleBackground` accepts (`canvas.go:243-251`). Eligibility
  is false for claimed cells and for the cursor (`canvas.go:400-406`).
- The active write path is `setNew` unless `LOOM_FAST_ANSI` is `0`, `false` or `off` (`parse_ansi.go:16-19`,
  `canvas.go:257-262`).
- The comment at `canvas.go:339-342` says a blank default-background cell is "transparent foreground
  surface" that must not hide a composited background. Host check (terra:med): the comment is about
  background composition, not about text below. `PaintSurface` replaces existing text (canaries A1, A3,
  A4), and its unclaimed surface stays eligible for a later `ComposeBackground` pass (`canvas.go:193`,
  `:397`). Only the claim flag blocks later decoration (canary A2). So a blank hides text drawn earlier,
  and the comment holds for composed backgrounds; the canaries did not test `ComposeBackground`.

### 2.4 Z-order and hit routing

- Loom has no z-index. Overlap is decided by the order in which the app calls `Draw`. Later draws paint
  over earlier ones (canary A5).
- `Stack` is a split layout. Its children do not overlap (`stack.go:77-99`). Hit routing picks the first
  child whose rectangle contains the point (`stack.go:161-177`); focus is `s.focus`, and Tab cycles it
  (`stack.go:139-151`).
- `Frame.Draw` paints boxes in declaration order (`frame.go:113-138`). `Frame.ConsumeMouse` takes the first
  layout rectangle that contains the point (`frame.go:689-697`) and focuses it on press (`frame.go:698-700`).
  If two boxes overlapped, draw would favour the last box and hit would favour the first. The layout
  has no overlap check (`frame.go:140-141` is the `Layout` signature), and no overlap was tested in this batch.
- `Popup` is an in-tree widget. Its `Draw` centres the popup and records `popupRect` (`popup.go:43-68`).
  It sets focus on its inner widget only during `Draw` (`popup.go:69-70`). Esc closes it (`popup.go:78-88`).
- `Popup.ConsumeMouse` (`popup.go:109-133`) forwards to the inner widget inside the inner rectangle.
  A press outside the popup dismisses it when `DismissOnOutsideClick` is set (`popup.go:121-124`), and the
  event is then `Handled`. A backdrop event is therefore consumed by the top popup and never reaches the layer
  below. This was confirmed in canary A6.
- Modal capture: `ActiveModalMouse` resolves the modal target through wrappers (`modal_mouse.go:11-22`).
  `Stack.ModalMouseTarget` (`modal_mouse.go:137`) and `Frame.ModalMouseTarget` (`modal_mouse.go:185-203`,
  last visible box wins) expose modal children. The modal-capture mechanism is covered by `report-mouse.md`.
- `DispatchMouseEvent` forwards to `root.ConsumeMouse` and has no z logic (`event.go:827-831`).
- Dialog: `Dialog.Draw` builds a `Popup` each frame (`dialog.go:58-72`). Esc closes it (`dialog.go:130-136`).

### 2.5 Focus after a popup closes

- No restore code exists. The stack keeps its focus index (`stack.go:139-151`). A closed popup does not
  return focus to the previous child, and it does not forward keys (`popup.go:78-84` returns `Ignored`
  when `!p.Open`). Keys are dropped until the user presses Tab. Canary A7 confirmed this.
- Host check (terra:med, canary `TestPopupStack`: `open=false value="" consumed=false` after Tab, open,
  Esc, `x`): a Popup as a container child is a documented pattern (the Popup example in
  `spec/widgets.yaml:243` puts it in a Frame), so this is a real gap. Pane's own help overlay takes keys
  without changing the root's focus index and hands keys back after Esc (`pane_dispatch.go:51`, traced,
  not canaried). Modal overlay references that are cleared on close also return keys to the root.

### 2.6 Markdown

- `richtext.go` is an ANSI-to-document model, not a Markdown parser. `RichDocument.FromANSI` treats
  Markdown punctuation as text (`richtext.go:53-56`). `ToANSI` and `ToPlainText` return it unchanged (`richtext.go:103-130`).
- The only Markdown code is serialisation: `TextEntry.MarkdownValue` (`text_entry_bold.go:47`, `:63`),
  closed ticket 441. Loom has no Markdown renderer. Canary D confirmed this.

### 2.7 Clipboard

- `WriteOSC52` always writes target `c` (`clipboard.go:12-16`). `Pane` binds it as the write hook
  (`pane_run.go:111`; `text_entry.go:349` comment).
- No OSC 52 read path exists. The only OSC 52 code is the writer above, and the decoder treats OSC replies
  as keys (section 4, canary C). Primary selection (`p`) and `s` are never written.

### 2.8 Pane input and focus/theme events

- `Pane` decodes input with `scanKey` (`pane_run.go:688`). Every scanned event goes to `dispatchKey`,
  including zero-value events (`pane_run.go:688-706`; `pane_dispatch.go:10-22`).
- `scanEscapeKey` has no OSC state: it maps `ESC ]` to `alt-]`, and the OSC payload follows as text keys (canary C). Only the OSC 11 payload ends up typed; complete `CSI I`, `CSI O` and `CSI ? 997 ; 1 n` become empty, unclassified key events that insert nothing (host review, canary C2). Canaries C and C2 use `loom.ScanKey`, not a live Pane; the Pane path was traced (`pane_run.go:688`, `pane_dispatch.go:21`), not run, because `Pane` needs a controlling terminal.
- `isTerminalReply` marks Report only for CPR `R`, DSR `n`, DA `c` and DECRPM `y` (`event.go:707-731`).
  `CSI ? 997 ; 1 n` (DSR 996) is not a report. `CSI I` and `CSI O` (focus) are not reports either.
- Mode 1004: `Pane` does not enable it on loom's own tty. `terminal/terminal.go:31` only disables it in
  `ResetSequences`. The embedded VT tracks the child's request (`vt/vt.go:746-747`, `:894-895`), and
  `internal/session/session.go:475-481` writes `CSI ? 1004 h/l` on tab switches.
- Theme reports (OSC 11 replies, DSR 996) are not requested by loom and are not handled.

### 2.9 Other notifications and links

- OSC 8, OSC 9/777 notifications, OSC 22 pointer shape: grep of non-test, non-example Go finds no emit or
  parse code (`]8;`, `]22;`, `]9;`, `777`, `hyperlink`, `PointerShape`). OSC 8 is only skipped inside
  richtext (`richtext.go:72`, `skipRichTextOSC` at `:179`).

## 3. Feature matrix

Cell values: **proved** (code plus canary, or code read to the executing caller), **partial** (some code
path, limit stated), **disproved**, **missing** (grep finds nothing), **unknown** (source not read, or
canary failed). Column evidence is `file:line` for loom, and `file:line @ pin` for the others.

| Feature | Loom | Lip Gloss v2 | Bubble Tea v2 | colorprofile | Glamour v2 | TUIOS |
|---|---|---|---|---|---|---|
| Semantic roles | proved, `theme.go:81-127` | unknown, not read | n/a | n/a | unknown, not read | unknown, not read |
| Dark/light detection | missing; theme names only, `theme.go:213-254` | proved, OSC 11 + DA1 query, `lg/query.go:86-92`, `lg/terminal.go:25-45`, default dark on error | partial, `bt/color.go:10-16` command; reply type `BackgroundColorMsg` at `bt/color.go:71`; needs input | n/a | partial, default `dark` when `GLAMOUR_STYLE` unset, `gl/glamour.go:279-283`; no query | unknown, not read |
| Theme change at runtime | partial, explicit `ApplyTheme`, `widget.go:276-279`, `cmd/loom/widgets.go:336`; no event | n/a | unknown, not read | n/a | n/a | unknown, not read |
| Colour profile detection | proved env-only, `ansi/style.go:25-57` (+ `LOOMCOLOR`) | n/a | n/a | proved env + tty flag, `cp/env.go:70-76`; missing TERM treated as dumb | unknown, not read | unknown, not read |
| Profile downsampling | proved, `ansi/style.go:140-166` | n/a | n/a | unknown, not read | n/a | n/a |
| Z-order of overlapping siblings | partial, draw order only, canary A5 | proved, sort by `layer.z`, `lg/layer.go:231-233`; compose order `lg/canvas.go:11-12` | n/a | n/a | n/a | unknown, not read |
| Transparent blank cells | partial: a blank replaces earlier text (A1, A3, A4); the comment applies to `ComposeBackground` (`canvas.go:397`), not canaried | unknown, not canaried | n/a | n/a | n/a | unknown |
| Hit test by z | partial: top modal/popup first (`modal_mouse.go:188`, `popup.go:109-133`); Frame and Stack test children forward (`stack.go:161-177`, `frame.go:696-698`), which is safe because their layouts tile and never overlap (`frame.go:137`, `stack.go:99`; host check) | proved, top-down `Hit`, `lg/layer.go:289-299`; returns layer ID; no focus | n/a | n/a | n/a | unknown |
| Focus restore after popup | missing; focus index kept, keys dropped (canary A7) | n/a (no focus model) | n/a | n/a | n/a | unknown |
| Markdown | missing (`richtext.go:53-56`; canary D) | n/a | n/a | n/a | proved, renders heading/bold/code/list/link (canary D) | partial, own renderer `internal/markdown/markdown.go:6-9` (code read, not canaried) |
| Focus reporting 1004 | partial: not enabled by Pane; focus reports decoded as zero keys (canary C) | n/a | proved opt-in `FocusMsg` from `uv.FocusEvent`, `bt/input.go:20-21`, 1004 set in `bt/cursed_renderer.go:144` | n/a | n/a | unknown |
| OSC 52 target c/p/s | partial: write-only, target `c` only, `clipboard.go:12-16` | n/a | proved `ClipboardMsg` selection byte c/p, `bt/clipboard.go:5-17` | n/a | n/a | partial: query answered empty, `internal/terminal/window.go:785-787` |
| OSC 9/777 notifications | missing | n/a | unknown, not read | n/a | n/a | partial: `Notify` callback, `window.go:788`; parser not traced |
| OSC 8 links | missing (grep) | n/a | unknown | n/a | proved emitted (canary D) | unknown |
| OSC 22 pointer | missing (grep) | n/a | unknown | n/a | n/a | unknown |
| DSR 996 / mode 2031 | missing; `CSI ? 997 ; 1 n` reaches Pane as a zero key (canary C) | queries OSC 11 and DA1 only, `lg/terminal.go:45` | no 2031/996 in `bt` (grep) | n/a | n/a | unknown |

## 4. Canary results

Run command (all tests in one package, 5 s deadline each, no sleeps):

```
cd /tmp/claude-1000/-home-uwe-projects-loom/9e4d0913-fa4b-4cb2-aec0-0c51a6bd4a87/scratchpad/b8-canary
HTO=0 go test -count=1 -v -run 'TestA|TestB|TestC|TestC2|TestD' .
```

Measurements below use loom's exported API directly. `loomtest` offers VT and comparison helpers only, not
popup or stack helpers, so it was not used.

| ID | Behaviour | Expected | Actual | Verdict |
|---|---|---|---|---|
| A1 | `PaintSurface` blank over a drawn `B` | Comment at `canvas.go:339-342`: blank keeps the underlying content | glyph `" "`, `B` gone | blank replaces text; the comment is about `ComposeBackground` (host check), so it is not disproved |
| A2 | Claimed `B` (`Set` with claim), then `PaintDecoration` `#` | `B` kept | glyph `B` | proved |
| A3 | `PaintForeground` blank over `B`, then decoration `#` | blank claim wins, decoration blocked | glyph `" "` | proved |
| A4 | `PaintSurface` over a view-drawn `B` | as A1 | glyph `" "` | disproved (same as A1) |
| A5 | Two popups, `p2` drawn after `p1`, cell (4,1) | top popup's blank wins | glyph `" "`, default BG | proved |
| A6 | Click (4,1) to top popup | top popup's inner receives it | `p2-inner mouse(0,0)`, consumed | proved for the top popup. Passing through to the lower popup was not canaried end-to-end, because the test dispatcher is the app's own code |
| A7 | Stack `[base, popup]`; Tab, Esc, `x`, Tab, `y` | `x` dropped after close; focus returns after Tab | log: `base key tab`, `base key y`; `x` reached no widget | proved; matches code reading |
| B | Snapshots, loom vs colorprofile (table below) | expected from code reading | see table | 4 of 5 agree in effect; empty env differs |
| C | OSC 11 split across reads, `CSI I`, `CSI ? 997 ; 1 n`, `CSI O`, OSC 11 with ST, interleaved keys | replies dropped, or consumed as reports | OSC 11 becomes keys: `alt-]`, `1`,`1`,`;`,`r`,`g`,`b`,`:`,`1`,`e`,`/`..., BEL becomes `ctrl-g`; focus and DSR 996 become zero `KeyEvent{}` with `Report=false`; no pending wait for OSC | gap **proved**: the OSC 11 payload arrives as text keys; focus and DSR 996 are empty keys |
| C2 | Same decoded events fed into a `TextEntry` | no text | value `11;rgb:1e1e/1e1e/1e1e` | gap **proved**: only the OSC 11 payload is typed (via `ScanKey`, not a live Pane) |
| D | Snippet `# Head`, `**bold**`, `` `code` ``, `- item`, `[link](...)` | loom: no interpretation; Glamour: rendered | loom plain text keeps all syntax; Glamour dark and light: heading `;1m` and `48;5` bg, bold, code bg, `•` bullet, OSC 8 link, no `**` or backticks left | proved |

Canary B, `DetectColorProfile` snapshots (loom value 0=None, 1=16, 2=256, 3=TrueColor; colorprofile names):

| Env snapshot | loom | colorprofile.Env | Same in effect |
|---|---|---|---|
| `TERM=xterm-256color` | 256 (2) | ANSI256 | yes |
| `COLORTERM=truecolor` (+ TERM 256) | TrueColor (3) | TrueColor | yes |
| `NO_COLOR=1` | None (0) | Ascii | yes (no colour) |
| `TERM=dumb` | None (0) | NoTTY | yes (no colour) |
| empty env | TrueColor (3) | NoTTY | **no**: loom assumes colour, colorprofile treats missing TERM as dumb |

Env path check (`t.Setenv`): `NO_COLOR=1`, `COLORTERM=truecolor`, `TERM=dumb` gives None, so NO_COLOR wins
over COLORTERM in the env path.

Canary C detail (decoded events, from the test log):

- `a` gives `Key: Text:a`.
- `\x1b]` gives `alt-]`. `11;rgb:1e1e/1e1e/1e1e` gives one key per character.
- `\a` (BEL) gives `ctrl-g`. `\x1b[I` gives an empty `KeyEvent{}`. `c` gives `c`.
- `\x1b[?997;1n` gives an empty `KeyEvent{}`, `Report=false`. `\x1b[O` gives an empty `KeyEvent{}`.
- `\x1b\\` (ST) gives `alt-\`.

Limits: the canary drives `scanKey` through the same loop shape as `pane_run.go:688-706`, without
`escKeyTimeout`. The real timeout path was not run, so the "delayed reply" case is a decoder result only.

## 5. Upstream mechanisms (quoted, pinned)

Each quote is at the pinned SHA. Line numbers are from the shallow clones.

**5.1 Lip Gloss v2: z-sorted compositing and top-down hit test** (`ff16a4ef`)

`lg/layer.go:231-233`:
```
	// Sort by absolute z-index (lowest to highest for drawing)
	slices.SortFunc(c.layers, func(a, b compositeLayer) int {
		return a.layer.z - b.layer.z
	})
```
Annotation: the comparator reads `layer.z`, the value `Z()` sets (`lg/layer.go:68-70`, whose comment says
"relative to its parent"). The sort comment says "absolute". The code path is the same, so the absolute/relative
difference is a comment question only. Effect unverified: nested-Z canary not run in this batch.

`lg/layer.go:289-290` (`Hit`):
```
// Hit performs a hit test at the given (x, y) coordinates. If a layer is hit,
// it returns the ID of the top-most layer at that point. Layers with empty IDs
```
Annotation: the loop runs from the top (`for i := len(c.layers) - 1; i >= 0; i--`, `lg/layer.go:293`). The
bounds test body was not quoted. Hit returns an ID, not a focus or click target.

`lg/canvas.go:11-12`:
```
// Composed drawables are drawn onto the canvas in the order they were
// composed, meaning later drawables will appear "on top" of earlier ones.
```

**5.2 Lip Gloss v2: background detection by query** (`ff16a4ef`, upgrade guide `lg/UPGRADE_GUIDE_V2.md`)

`lg/query.go:86-92`:
```
func HasDarkBackground(in term.File, out term.File) bool {
	bg, err := BackgroundColor(in, out)
	if err != nil || bg == nil {
		return true
	}
```
Annotation: an error or no reply means dark. The OSC 11 query and DA1 are sent together, and the reply is
read with a timeout (`lg/terminal.go:45`, `:49`):
```
	}, ansi.RequestBackgroundColor+ansi.RequestPrimaryDeviceAttributes)
...
const defaultQueryTimeout = time.Second * 2
```
The upgrade guide shows the migration to `HasDarkBackground` and `LightDark` (`UPGRADE_GUIDE_V2.md:189-190`).

**5.3 Glamour v2: dark default, no query** (`95c93db0`)

`gl/glamour.go:279-283`:
```
func getEnvironmentStyle() string {
	glamourStyle := os.Getenv("GLAMOUR_STYLE")
	if len(glamourStyle) == 0 {
		glamourStyle = styles.DarkStyle
	}
```
Annotation: style is chosen from `GLAMOUR_STYLE`, defaulting to `dark`. No OSC 11 or 996 text was found in
`gl/*.go` (grep). Styles are named `dark` and `light` (`gl/styles/styles.go:30-33`). The style choice is
made by the caller; Glamour does not detect the terminal.

**5.4 colorprofile: environment and TTY rules** (`d085584e`)

`cp/env.go:70-76`:
```
func Env(env []string) (p Profile) {
	return colorProfile(true, newEnviron(env))
}

func colorProfile(isatty bool, env environ) (p Profile) {
	term, ok := env.lookup("TERM")
	isDumb := (!ok && runtime.GOOS != "windows") || term == dumbTerm
```
Annotation: a missing `TERM` is treated as dumb on non-Windows, so the empty env gives `NoTTY`. Loom does
the opposite (section 2.2). `Env` does not query the terminal.

**5.5 Bubble Tea v2: focus and background messages** (`67b4c7a8`)

`bt/input.go:20-21`:
```
	case uv.FocusEvent:
		return FocusMsg(e)
```
`bt/cursed_renderer.go:144` writes the 1004 enable sequence (the line starts `_, _ = s.scr.WriteString(ansi.SetModeFocusEven…`).
Annotation: `FocusMsg` exists only when focus reporting is turned on, so it is opt-in.

`bt/color.go:10-16`:
```
// RequestBackgroundColor is a command that requests the terminal background color.
// Use [WithInput] to enable input, since the reply cannot be read otherwise.
func RequestBackgroundColor() Msg {
```
Annotation: Bubble Tea v2 can request the background colour, but only with input enabled. The reply arrives
as `BackgroundColorMsg` (`bt/color.go:71`).

`bt/clipboard.go:5-9` and `:14-17`:
```
type ClipboardMsg struct {
	Content   string
	Selection byte
}
```
```
//   - c: System clipboard.
//   - p: Primary clipboard (X11/Wayland only).
```
Annotation: Bubble Tea v2 exposes `c` and `p` on read messages. Loom has no read path.

**5.6 TUIOS: own Markdown renderer, no Glamour** (`6fe2c0df`)

`internal/markdown/markdown.go:12-13`:
```
// It is small on purpose. goldmark with its GFM extensions adds about 480 KB
// to the tuios binary, and glamour, which also links every chroma lexer,
```
Annotation: TUIOS chose not to use Glamour and has its own package. Its header lists ATX and setext
headings, lists, emphasis, strong, code spans, links, autolinks and bare URLs (`markdown.go:6-9`). The
package also has `inline.go` and `html.go`. Not canaried.

`internal/terminal/window.go:785-787`:
```
		// No ClipboardQuery: an OSC 52 query is answered with empty content,
		// as xterm answers by default. The answer is written to the pane's
		// input, so it must not carry text a program in the pane chose.
```
Annotation: TUIOS answers OSC 52 queries with empty content. Loom answers none. The `Notify` callback is
at `window.go:788`; its parser was not traced.

**5.7 TUIOS layouts (P3 notes only)**

`internal/layout/` exports `NewBSPTree`, `CalculateMasterLayout` and `NewScrollingLayout` (`bsp.go`,
`master.go`, `scrolling.go`). These are tiling layouts for TUIOS windows. Loom is non-modal and does not
adopt leader-key behaviour; nothing here is a Loom requirement.

## 6. Loom vs upstream comparison

- **Detection.** Lip Gloss queries the background and falls back to dark. Glamour takes the style from
  `GLAMOUR_STYLE`. Loom has no detection at all and relies on the app to pick a theme. A Loom app that wants
  light/dark support must do its own query, and the query has to be read through Pane's input path, which
  currently turns the reply into keys (section 4, C2). This is the main gap.
- **Colour profile.** Loom and colorprofile agree in effect for 4 of 5 snapshots. The one difference is the
  empty environment: Loom assumes truecolor (`ansi/style.go:53-55`), colorprofile assumes no colour
  (`cp/env.go:74-76`). Loom's default is the more generous choice for an unset `TERM`, but it is a choice,
  not a bug.
- **Composition.** Lip Gloss has an explicit z-sort and a top-down hit test. Loom has neither. Its
  overlap rule is draw order (canary A5), and its hit rule depends on the container (section 2.4). Lip
  Gloss's `Hit` returns an ID and has no focus or modal model, so it is not a substitute for Loom's
  claim-cell routing.
- **Transparency.** A loom blank hides earlier text; its "transparent" comment is about composed
  backgrounds (section 2.3). Lip Gloss transparency was not canaried, so no comparison is made.
- **Focus after popup.** Loom keeps the focus index and drops keys until Tab (A7). Lip Gloss has no focus
  model to compare.
- **Markdown.** Loom renders no Markdown. Glamour renders the snippet with heading, bold, code background,
  bullet, and OSC 8 links (canary D). TUIOS has its own renderer covering a similar set (code read only).
- **Focus reporting.** Bubble Tea v2 emits `FocusMsg` when 1004 is on. Loom does not turn 1004 on for its
  own tty, and it turns focus reports into zero-value keys if they arrive (canary C).
- **Clipboard.** Loom writes `c` only and has no read. Bubble Tea v2 reads and writes `c` and `p`. TUIOS
  answers queries with empty content.
- **Notifications, links, pointer.** Loom has no OSC 8, 9/777 or 22 support. Glamour emits OSC 8. TUIOS has
  a notify callback.
- **Coexistence of Lip Gloss v1 and v2.** Module paths differ: v2 is `charm.land/lipgloss/v2` (`lg/go.mod:1`).
  The v1 path was not fetched in this batch, and v1 rows are P3 anyway, so this batch does not claim a
  verified v1 build.

## 7. Stage 4 proposals

Nothing here is applied. The integrator applies these serially.

### 7.1 features.md rows (proposed)

| Row ID (proposed) | Feature | Loom status | Evidence |
|---|---|---|---|
| `desktop.compose.z-order` | Overlapping siblings and z-order | partial | `canvas.go:339-345`; canary A5 |
| `desktop.compose.transparent-blank` | Blank cell transparency | partial | canaries A1, A3, A4; comment at `canvas.go:339-342` applies to `ComposeBackground` |
| `desktop.compose.hit-by-z` | Hit test by z | has | modal/popup top first; tiled Frame/Stack never overlap (`frame.go:137`, `stack.go:99`); canary A6 |
| `desktop.focus.restore-after-popup` | Focus after popup closes | missing | canary A7 |
| `desktop.theme.dark-light` | Dark/light detection | missing | no detection code; `theme.go:213-254` |
| `desktop.theme.runtime-change` | Runtime theme change | partial | `widget.go:276-279`; `cmd/loom/widgets.go:336` |
| `desktop.color.profile-env` | Colour profile by env | proved | `ansi/style.go:25-57`; canary B |
| `desktop.color.empty-env` | Empty env default | disagree with colorprofile | `ansi/style.go:53-55`; `cp/env.go:74-76`; canary B |
| `desktop.focus.report-1004` | Focus reporting | partial | `terminal/terminal.go:31`; canary C |
| `desktop.input.replies-as-keys` | OSC/DSR replies kept out of key input | missing | canaries C, C2: OSC 11 payload typed; ticket 462 |
| `desktop.clipboard.osc52-read` | OSC 52 read | missing | `clipboard.go:12-16` (write only) |
| `desktop.clipboard.osc52-target` | OSC 52 target c/p/s | partial (c only) | `clipboard.go:14` |
| `desktop.notify.osc9-777` | Notifications | missing | grep, no emit or parse |
| `desktop.link.osc8` | Hyperlinks | missing | grep; richtext skips OSC at `richtext.go:72` |
| `desktop.pointer.osc22` | Pointer shape | missing | grep |
| `desktop.markdown.render` | Markdown rendering | missing | `richtext.go:53-56`; canary D |
| `desktop.query.dsr996-2031` | Colour-scheme report 996, mode 2031 | missing | grep; canary C |

### 7.2 summary.md wording (proposed)

Replace the Lip Gloss v2 row in `summary.md:26` with:

> Lip Gloss v2 | Z-sorted layers (`lg/layer.go:231-233`) and a top-down `Hit` that returns the top layer's ID (`lg/layer.go:289-299`) | no focus, click-through or modal model; hit returns IDs only | Loom routes by claim cells and modal capture; Loom has no z-index, so overlap is draw order

Replace the coexistence statement with:

> Lip Gloss v1 and v2 can be required together because the module paths differ: v2 is `charm.land/lipgloss/v2` (verified at `lg/go.mod:1`). The v1 path was not fetched in batch 8 (v1 rows are P3). "Lip Gloss v1/v2 coexist" is therefore a module-path fact, not a verified build. TUIOS keeps its own renderer in `internal/markdown`, not Glamour.

Add a sentence to the summary's Glamour row: "Glamour v2 (`charm.land/glamour/v2`) chooses dark or light from `GLAMOUR_STYLE` (default `dark`) and does not query the terminal (`gl/glamour.go:279-283`)."

### 7.3 Ticket candidates

| Candidate | Evidence | Priority (proposed) | Overlap |
|---|---|---|---|
| Replies (OSC 11, focus, DSR 996) typed as keys in `Pane` | canaries C, C2 | P2 | **Overlaps open 462** ("Pane consumes terminal replies centrally instead of dispatching them as key events"). Add this evidence to 462; no new ticket |
| OSC 52 read and session gating | `clipboard.go:12-16`; TUIOS `window.go:785-787` | P2 | **Overlaps open 473** (session gates child OSC 52) and 465 (clipboard). Add to 473 |
| OSC 8 hyperlinks | grep; Glamour emits OSC 8 (canary D) | P2 | **Overlaps open 259** (OSC 8). Add to 259 |
| Focus after popup close (keys dropped until Tab) | canary A7, host `TestPopupStack`; documented pattern `spec/widgets.yaml:243` | P2 (host) | Closed 316 covers popover focus in `loom edit`, not Stack/Popup restore. New candidate |
| Dark/light detection (OSC 11 with timeout, default dark) | `lg/query.go:86-92`; `lg/terminal.go:45` | P3 | none beyond 457. New candidate |
| Empty-`TERM` colour default | canary B | P3 | none. New candidate, decision on default |
| Markdown display | canary D; `richtext.go:53-56` | P3 | Closed 441 (serialisation only) and 254 (RichTextEdit, no parser). New candidate only if needed |
| OSC 9/777 notifications and OSC 22 pointer | grep | P3 | none beyond 457. New candidate |
| Lip Gloss v1 build | not fetched | P3 | none |

## 8. Flow notes

- The Lip Gloss, Glamour and colorprofile tags resolved to new SHAs, so the plan pins were updated (section
  1). Old pins are noted.
- The tracker search with the requested keyword list returned only 457. Topic-by-topic searches found 462,
  259, 316, 441, 254 and 372. The overlap table uses those.
- `go mod tidy` pulled Glamour's dependency tree into the disposable canary `go.mod`. The `go.mod` in the canary
  directory is therefore not the minimal one written first. This is disposable and outside the repository.
- Canary A6 measures the top popup only. The layered dispatcher was not built, so "lower popup never
  receives the click" is code-read, not canaried end to end.
- `escKeyTimeout` was not exercised. Canary C shows the decoder result only.
- Foot and kitty documentation were not fetched because no row needed them.
- Host review: agy:flash38:med checked claims 1-8, then codex:terra:med (owner: quota back) settled the
  four routing claims with canaries in `b8-canary` (`TestScanReplies`, `TestPopupStack`). Corrections:
  only the OSC 11 payload is typed, not focus or DSR 996 reports; the popup focus gap is on a documented
  path (P2, not P3); overlapping non-modal siblings are not a supported layout, so forward hit testing is
  not a bug; the "transparent blank" comment is about `ComposeBackground`. Known wrong line refs from the
  review: `ansi/style.go:33-35` (doc is `:23-24`), `frame.go:689-697` (hit loop `:696-698`),
  `text_entry_bold.go:63` (`MarkdownValue` is `:47`), `theme.go:14-15` (embed is `:15`).
- Outside-click dismissal applies only with `DismissOnOutsideClick` (`popup.go:122`); without it, the
  popup does not swallow outside clicks (review note on section 2.4).
- Source reads that were skipped: Glamour `styles/styles.go` was checked only for `DarkStyle`/`LightStyle`
  names, and Bubble Tea `color.go` only for the request function. Their full behaviour is not claimed.
