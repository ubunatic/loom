# Topic 7: Contrast with tview, tcell, gocui (and Textual for one feature)

Ticket 457, searcher 7 of 7. Read-only research; no loom files changed. Output: `features-contrast.yaml` (27 rows, all with a file-and-line source). Date of source snapshot: 2026-11-10.

## Scope and method

- **tview** (`rivo/tview`, the widget library built on tcell): `application.go`, `textarea.go`, `doc.go` at `master`.
- **tcell** (`gdamore/tcell`, now module `tcell/v3`, latest release v3.6.0 on the releases page): `mouse.go`, `screen.go`, `tscreen.go`, `key.go`, `cell.go`, `go.mod` at `main`.
- **gocui** (`jesseduffield/gocui`, the lazygit library): `gui.go`, `keybinding.go`, `view.go`, `text_area.go`, `edit.go` at `master`.
- **Textual** (Python, only one feature): the command palette doc.
- I downloaded the raw source files with `curl` into a fresh temp directory and grepped them, so every line number in the YAML is exact for the snapshot date. I did not run any of the downloaded code.
- Web-fetch helpers read the rendered pages for release tags, README feature lists and doc prose. Where a claim came only from a rendered page with no line number, the row says so or is left out.

## Key findings

1. **tview has a multi-line TextArea with the full desktop key set.** It has undo and redo (Ctrl-Z, Ctrl-Y, stack in `textarea.go#L179-L180`), Shift-select (`textarea.go#L145`), word jumps (`textarea.go#L118`) and double-click word select (`textarea.go#L188`). Loom already has these, so for loom it is a reference, not a gap.

2. **tview picks Ctrl-Q for copy, not Ctrl-C.** The reason given in the doc is that Ctrl-C is the default key to stop the application (`textarea.go#L154-L160`). Loom uses Ctrl+C for copy. This is a real design split to keep in mind for the P1 "desktop-like input" work. The copy buffer in tview is internal to the widget unless the app replaces it.

3. **tview's paste rule is clear and worth copying.** Paste is for inserting text, never for navigation or selection (`application.go#L251`). Pasted key events also skip the app-wide input capture (`application.go#L159`). Both matter for loom's bracketed paste and huge-paste robustness.

4. **tview's mouse capture is one global hook.** `SetMouseCapture` (`application.go#L172-L178`) can rewrite or drop an event, and returning nil consumes it and triggers a redraw. This is the same kind of routing point as loom's pane mouse capture, so the two designs can be compared directly.

5. **tview mouse is off by default.** The doc warns that turning mouse on changes the terminal's own mouse behaviour (`doc.go#L161-L165`). Loom's default should be checked against this.

6. **tcell reports mouse as flags, and drags are motion events.** `MouseMotionEvents` is a flag that turns on all motion, click and drag events (`screen.go#L315`). Click-drag is a motion event with the button held (`mouse.go#L20-L22`). Some terminals send only press and release, so a drag that leaves the window depends on the terminal. This is the same problem loom faces with pane drags leaving bounds.

7. **tcell's kitty keyboard path is query first, then push.** "CSI" is the terminal escape-sequence prefix (ESC [). tcell sends a query (`CSI ?u`, asking whether the terminal supports the kitty keyboard protocol) and then `CSI =31u`, which asks for disambiguation, events, alternate keys, all keys and text (`tscreen.go#L203-L205`). Loom sends `CSI >1u`. The flag sets differ, so the two should be compared before loom changes anything. tcell can also be forced with the env var `TCELL_KEYBOARD_PROTOCOL` (`tscreen.go#L376`), which helps in tests.

8. **tcell writes one grapheme cluster per cell.** `CellBuffer.Put` writes only the first grapheme of the string (`cell.go#L60-L63`). The doc does not say how many cells a ZWJ emoji takes, so this needs a test, not a claim. tcell v3 measures width with `clipperhouse/displaywidth`, not with `x/ansi` or `runewidth` (`go.mod#L6`). Loom's `measure` package is a separate choice and should be compared row by row.

9. **gocui shares one key enum between keys and mouse, and it collides.** `MouseLeft` and `KeyShiftArrowDown` are both `tcell.KeyF63` (`keybinding.go#L279`, `keybinding.go#L229`). The mouse event therefore shares a code space with keyboard keys. Loom's event model should not repeat this.

10. **gocui's text area is basic.** It has an internal clipboard string (`text_area.go#L36`, used at `text_area.go#L452`). No undo, redo or Shift-select was found in `text_area.go`. Its view selection is by line only (`view.go#L396`). It is a weaker reference than tview's TextArea.

11. **gocui's Update is the thread-safe entry point.** `Gui.Update` (`gui.go#L662`) runs a function safely from another goroutine. `UpdateAsync` does not spawn one. This is the closest analogue to loom's `Pane.Post`, and the two can be compared on backpressure under high-frequency updates.

12. **Textual's command palette is the only feature here that Go libs may lack.** It opens on Ctrl+P, uses fuzzy matching over commands from providers (`command_palette.md#L13`) and can be rebound with `COMMAND_PALETTE_BINDING`. This is a Python library, so loom would only copy the idea. Whether a Go lib has the same is not checked (open gap).

## What this means for the P1 priorities

These are reading notes, not decisions. The owner decides.

- **Input and copy keys:** tview's Ctrl-Q choice is the one concrete contrast. Loom's Ctrl+C copy is already decided, so the question is whether loom needs a second copy key or a documented exception.
- **Paste:** tview's rule that paste never navigates, plus skipping global key capture during paste, are testable against loom's paste path.
- **Event model:** gocui's shared enum collision is a warning. Mouse and keyboard should stay in separate types.
- **Mouse drags and motion:** tcell's motion flag and drag-as-motion description give loom a reference for drags that leave bounds.
- **Kitty keyboard:** compare tcell's query-then-push sequence with loom's `CSI >1u` before any change.
- **Width:** tcell's single-grapheme `Put` and its own width library are a second reference next to `measure`.

## Open gaps

- **ratatui**: only the widget list was read (docs.rs, Ratatui 0.30.2: Canvas, Chart, Scrollbar, Table, Sparkline, Calendar and more). It has no line-anchored source, so no rows were written. It is Rust, and the ticket limits it to features no Go lib has. Not researched further.
- **awesome-gocui** (the tcell-based gocui fork): the README says it is built on tcell with one-line views and mouse support, but I did not fetch a source file with line numbers. No rows.
- **tview release date**: the latest tag is v0.42.0 ("27 Aug" on the page, year not visible). The tview repo may have newer tags after that.
- **tcell `EnableMouse` doc**: lives in `screen.go` near line 125; the flag list was read, the full doc comment was not. Its release date: v3.6.0 is dated "11 Oct" on the page, year not confirmed beyond the © 2026 footer.
- **gocui last commit date**: not visible on the page; the repo was not checked for maintenance status. Release cadence is in the host's `summary.md` task.
- **Textual screens and modal behaviour**: the doc was truncated in the fetch; the modal part came from a summary and was not verified. Not used in a row.
- **tview TextArea `SetClipboard`**: named in a fetched summary, not grepped. Not used in a row.
- **Go equivalent of the command palette**: not searched (Bubbles, Huh, others).
- **Line numbers** refer to the `master`/`main` snapshot on 2026-11-10. They can drift; re-run the grep before filing a ticket.

## Files

- `docs/studies/tuios/features-contrast.yaml`: 27 rows. Priority: 10 P1, 16 P2, 1 P3. All rows have `verified: true` (checked against the downloaded source). `loom: unknown` on every row for the host to fill.
- `docs/studies/tuios/report-contrast.md`: this file.

Rows written: 27. Rows verified: 27 (all grepped against downloaded source, quotes under 120 characters). Open gaps: ratatui rows (no line-anchored source), awesome-gocui rows, tview release year, tcell EnableMouse doc, gocui maintenance status, Textual modal screens, command-palette Go equivalent.
