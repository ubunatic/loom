# TUIOS study: feature list (ticket 457)

Merged from `features-*.yaml` (189 rows from 7 searchers, 179 verified by quote).
The `loom` column was filled by the host from loom code at `163b3738`, only for P1 rows.
P2 and P3 rows stay in the YAML files with `loom: unknown`; check them when a ticket needs them.

Legend: **has**, **partial**, **missing**, **n/a** (not wanted for loom's GUI-like model).
"UNVER" = the searcher could not quote the source.

## P1 gaps and partials (ticket candidates)

| Area | Feature | Where seen | loom | Evidence in loom |
|---|---|---|---|---|
| render | Temporal row/cell diff: write only changed rows/cells | uv.transformline-diff (cell comparison), bt.renderer.cell-based; uv.line-hash-diff is only a fullscreen scroll optimisation | **missing** | `canvas.go` `FlushWithConfig` writes every row on a dirty flush; redraw is dirty-gated (`pane_run.go:432`). Canary 80×24 (report-render.md): one changed cell costs Canvas and Pane 2269 bytes, Bubble Tea v2.1.1 9 bytes; a forced unchanged frame costs 2269 vs 0 |
| render | Within-frame style transitions (SGR only on change) | tuios.uv-style-diff (SGR pair cache) | has | `Canvas.Row` emits SGR on style change |
| render | One writer call per canvas frame | uv.render-flush-split | has | `ResizeConfig.AtomicFlush`; Pane ancillary writes separate, unmeasured |
| render | Mode 2026 synchronized output | bt.sync-output-2026, uv.sync-output-2026, term.sync-output-2026 | has | `ResizeConfig.SynchronizedOutput`: configured output, not capability detection or whole-redraw atomicity |
| render | Query DECRQM 2026 before using it | term.decrqm-2026-replies | **missing** | Pane wraps 2026 by config, never queries; Bubble Tea v2 queries and enables on reset; loom VT answers 0 (report-queries.md V1, V3, V4) |
| input | Terminal replies kept out of widget key input | uv decoder typed events, tcell initQ | **partial** | Pane forwards replies as `KeyEvent{Report}`; kitty `CSI ? u` and `CSI 48 t` arrive as plain empty keys; late fragments become text keys (report-queries.md matrix; ticket 462) |
| input | Startup query batch ended by DA1 with a backstop | tcell (1 s), TUIOS (300 ms) | **missing** | only `loom-probe keys` does query + DA1 (1 s); Pane has no startup query |
| width | Mode 2027 grapheme width, queried by DECRQM | term.grapheme-2027, ansi.width-method, vt.widthmethod-mode-2027, tuios.width-mode-2027 | **missing** | no 2027 query or state; loom VT answers DECRQM 2027 with 0 (unknown). `measure` probes only ZWJ width by CPR (`LOOM_ZWJ` override). ultraviolet queries 2027 at start and switches buffer and renderer width (report-width.md V6) |
| width | Grapheme segmentation matches current Unicode | ansi.stringwidth-grapheme-default, tcell.cell.put-grapheme | partial | loom splits Indic conjunct `क्ष` into two clusters (width 2); x/ansi and tcell v3 keep one cluster (width 1). Flags, marks, ZWJ handled (report-width.md V1) |
| width | Cell buffer stores grapheme clusters | tcell.cell.put-grapheme | partial | leading cell plus continuations, using loom's own segmentation; a wide glyph crossing the right edge is rejected whole (report-width.md V2) |
| input | Partial escape stays buffered until complete | input.parsesequence-consumed | has | `esc_key_timeout` 50 ms (`docs/TuiInput.md:56`) |
| input | Kitty keyboard flags beyond 1 (event types, alternate keys) | ansi.kitty-keyboard-flags, term.kitty-keyboard.flags (UNVER) | partial | loom pushes flag 1 only (`docs/TuiInput.md:96`) |
| input | Query kitty support (`CSI ? u`) before enabling | tcell.kitty-keyboard.enable | partial | only in `loom-probe keys`, not in `Pane` (opt-in flag) |
| input | In-band resize, mode 2048 | term.inband-resize-2048 | **missing** | Pane uses SIGWINCH; a 2048 report becomes an empty key; loom VT does not recognise 2048. tcell queries, enables, decodes (with pixels) and disables it (report-queries.md V6). Not ticketed: SIGWINCH works |
| input | Paste sanitised of control chars | tuios.bracketed-paste-sanitize | partial | TextEntry strips controls; TextInput keeps ESC and Ctrl+C, TextArea inserts raw, RichTextEdit ignores paste. Canvas drops controls on output (report-input.md; ticket 464) |
| input | Paste never triggers shortcuts or navigation | tview.application.paste-no-navigation | has | paste frame separate from keys; no shortcut fired through the Pane route; 1 MiB cap drops overflow (report-input.md) |
| input | Kitty event types and associated text (flags 2, 4, 8, 16) | ansi.kitty-keyboard-flags | partial | `decodeCSIu` drops fields after `:`; `KeyEvent` has no event type. Unreachable today since Pane pushes flag 1 only; not ticketed (report-input.md) |
| events | High-rate producers: `Send`/`Post` without blocking | bt.send-blocks | has | `Pane.Post` appends under a mutex and invalidates; no block, but no cap |
| events | Gesture ownership after press (drag capture) | tview.application.mouse-capture is an interceptor, not ownership | partial | private Pane capture for scrollbar and pan, `ModalMouseCapture` takes precedence; TextEntry selection does not capture, and `SetFocus(false)` never ends a capture (report-mouse.md; ticket 467) |
| events | Drag auto-scroll while held outside content | tuios.drag-autoscroll | **missing** | no tick; 0 events over 300 ms held outside (report-mouse.md; ticket 467). TUIOS: 50 ms tick, started by the first outside motion |
| events | Horizontal wheel (SGR 66/67) | uv, tcell decode both axes | **missing** | decoded as presses of buttons 2 and 3 (report-mouse.md; ticket 468) |
| events | Child-local 0-based coordinates in nested containers, modal precedence | | has | canary with nested Stacks at nonzero origin and a modal (report-mouse.md) |
| events | Click-drag as motion while button held | tcell.mouse.drag-motion | has | `MouseDrag` |
| robust | Panic recovery restores terminal | bt.crash.recover | has | `recover()` in `pane_run.go:137` |
| robust | Suspend / release terminal for a child | bt.suspend, bt.terminal.release-restore, tview.application.suspend | has | `Pane.Suspend` |
| robust | SIGWINCH resize | bt.resize.sigwinch | has | `docs/TuiInput.md:35` |
| robust | Cleanup when start fails before first frame | tuios.run-before-init-gap | unchecked | check `Pane.Run` error paths |
| widgets | Fuzzy filter with ranking | bubbles.list.fuzzy-default-filter | has | `Choice.Fuzzy`, FilePicker |
| widgets | Table renders only visible rows | bubbles.table.virtualised-render | partial | Draw is viewport-bounded (about 90 µs at 100 and 100,000 rows); SetRows, widths and each filter keystroke scan all rows (3.4 ms, 6.7 ms at 100,000) (report-widgets.md) |
| widgets | Container keys do not break a focused multi-line editor | huh form field routing | **missing** | Form takes Tab/Down/PgDn/Enter before the focused TextArea (`form.go:276-297`); canary: Down moves focus, Enter submits (report-widgets.md) |
| widgets | Selection, undo and cut in every text widget | bubbles.textarea.*, tview textarea | partial | TextEntry and RichTextEdit have them; TextInput and TextArea only `SetSelection` from code, no undo, no cut (report-widgets.md) |
| widgets | Select filter in forms | huh.select.filter | has | Choice filter |
| widgets | Word jump, Shift-select, copy, select all in text inputs | bubbles.textarea.*, bubbles.textinput.word-forward | has | TextEntry only; loom uses Ctrl+A/C (Bubbles: Ctrl+G, Ctrl+Shift+C); in TextArea Ctrl+A is line start |
| widgets | Tab accepts a suggestion | bubbles.textinput.accept-suggestion-tab | n/a | loom Tab policy; TextEntry accepts placeholder only |
| compose | Z-ordered layers with hit test | lipgloss.layer.z-order, lipgloss.compositor.hit-test | has | root overlays and claim-cell routing; no z-index, overlap is draw order; modal popups hit first, Frame/Stack tile so siblings never overlap (report-desktop.md) |
| compose | Focus returns after a popup closes | | **missing** | Popup in a container: focus stays on the closed popup, keys lost until Tab (ticket 475); Pane help overlay is fine |
| theme | Dark/light background detection (OSC 11, mode 2031) | bt.theme.background-color, lipgloss query | **missing** | no query or report handling; an OSC 11 reply is typed as text (tickets 476, 462) |
| api | Copy key that does not collide with quit | tview.textarea.copy-is-ctrl-q | has | Ctrl+C copies only with a selection, else passes (`HasSelection`) |
| api | One filter slot, last wins | tuios.filter-single-slot | n/a | loom routes by widget tree |
| api | Shared program options across binaries | tuios.program-options-shared | n/a | |
| session | Daemon with attach/detach, one VT per pane | tuios.session-daemon | partial | `loom session` + `vt.VT`; one client, second attach refused; reattach repaints 24 screen rows, no scrollback or byte-offset resume (TUIOS has both); no live-stream gap (report-session.md) |
| session | OSC 52 write policy per pane | tuios.osc52-write-policy | **missing** | child OSC 52 writes and queries forwarded raw (`internal/session/session.go:764`, `:783`); TUIOS: off/on/ask/focused (ticket 473) |
| session | Kitty flag stack per screen | tuios.kitty-keyboard-stack | **missing** | one value (`vt/vt.go:58`, `:809-819`), leaks across `?1049`, `=N u` restores the cursor; reattach sends fixed `>1u` (ticket 472) |
| session | VT scrollback API | vt.scrollback-api | has | `SetScrollbackLimit`/`ScrollbackCells` (`vt/vt.go:139-170`); 5000 rows by spec, bound holds at 100,000 lines (report-session.md) |
| model | Leader-key modal design | tuios.leader-key-model | n/a | loom is non-modal by design |
| test | Headless model test harness | bt.test.teatest-wait | has | `loomtest.VT`, `CompareWidget`; teatest polls with sleep, loom forbids that |
| status | Ultraviolet API still unstable | uv.status-active-dev | n/a | P3 note: API stability is not a goal while loom is alpha |

## P2 and P3 highlights (not verified against loom)

- Focus in/out reporting (mode 1004): loom enables it only in `internal/session`, not in Pane.
- Dark/light detection: now a P1 row above (ticket 476).
- OSC 52 clipboard read and primary selection (loom writes only; checked in 458 batch 8).
- Markdown display: loom has none (Glamour and TUIOS render it; report-desktop.md canary D).
- Colour profile from env only in loom, as in colorprofile; empty `TERM` gives true colour in loom
  and "no TTY" in colorprofile (report-desktop.md canary B).
- XTVERSION and DA1/DA2 terminal identification (loom: only in `loom-probe` and `vt`).
- Text input: suggestions, validation, password echo; Huh: external `$EDITOR`, accessible mode,
  grid and column form layouts, multi-select limit.
- Key press/release and repeat events (kitty flag 2); wheel gesture axis lock (TUIOS).
- Frame rate cap (Bubble Tea 60 FPS default, 120 max).
- P3: OSC 8 links (ticket 259), OSC 9/777 notifications, OSC 22 pointer shape, OSC 133, kitty
  graphics, sixel. Batch 8 confirmed loom has none of OSC 8, 9/777, 22.
