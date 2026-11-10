# TUIOS study: feature list (ticket 457)

The seven `features-*.yaml` files contain 189 source records. Their `verified` flags are
original searcher labels, not an integration audit; nine are false and three are partial.
All Loom file references below are at revision `e2236900`. Where a row lacks a file,
symbol and line citation, its provisional status is **unchecked**: its batch report or
canary supplies the observation, but the implementation trace has not been pinned.
The YAML Loom fields have not yet been fully reconciled with this list.

Legend: **has**, **partial**, **missing**, **n/a** (not wanted for loom's GUI-like model).
"UNVER" = the searcher could not quote the source.

## P1 gaps and partials (ticket candidates)

| Area | Feature | Where seen | loom | Evidence in loom |
|---|---|---|---|---|
| render | Temporal row/cell diff: write only changed rows/cells | uv.transformline-diff (cell comparison), bt.renderer.cell-based; uv.line-hash-diff is only a fullscreen scroll optimisation | **missing** | `Canvas.FlushWithConfig` writes every row (`canvas.go:576-599`); `Pane.run` gates redraw on dirty (`pane_run.go:432`). Canary 80×24 (report-render.md): one changed cell costs Canvas and Pane 2269 bytes, Bubble Tea v2.1.1 9 bytes; a forced unchanged frame costs 2269 vs 0 |
| render | Within-frame style transitions (SGR only on change) | tuios.uv-style-diff (SGR pair cache) | has | `Canvas.Row` emits SGR on style change (`canvas.go:519-560`) |
| render | One writer call per canvas frame | uv.render-flush-split | has | `Canvas.FlushWithConfig` buffers with `ResizeConfig.AtomicFlush` (`canvas.go:576-610`); Pane ancillary writes separate, unmeasured |
| render | Mode 2026 synchronized output | bt.sync-output-2026, uv.sync-output-2026, term.sync-output-2026 | has | `Canvas.FlushWithConfig` wraps configured output (`canvas.go:576-599`), not capability detection or whole-redraw atomicity |
| render | Query DECRQM 2026 before using it | term.decrqm-2026-replies | **missing** | `Canvas.FlushWithConfig` wraps configured 2026 (`canvas.go:576-610`); Pane dispatches replies (`pane_run.go:688-708`), with no query found in report-queries.md V1, V3, V4; VT answers 0. Bubble Tea v2 queries and enables on reset |
| input | Terminal replies kept out of widget key input | uv decoder typed events, tcell initQ | **partial** | `scanCSI` marks some reports (`event.go:689-725`), `Pane.run` dispatches keys (`pane_run.go:688-708`); kitty `CSI ? u` and `CSI 48 t` arrive as empty keys and late fragments as text (report-queries.md matrix; ticket 462) |
| input | Startup query batch ended by DA1 with a backstop | tcell (1 s), TUIOS (300 ms) | **missing** | only `loom-probe keys` sends query + DA1 and waits 1 s (`cmd/loom-probe/keys.go:82-105`); Pane has no startup query (report-queries.md) |
| width | Mode 2027 grapheme width, queried by DECRQM | term.grapheme-2027, ansi.width-method, vt.widthmethod-mode-2027, tuios.width-mode-2027 | **missing** | no Pane 2027 query or state; `VT.privateModeState` returns 0 for 2027 (`vt/vt.go:735-759`). `measure` probes only ZWJ width by CPR (`LOOM_ZWJ` override). ultraviolet queries 2027 at start and switches width (report-queries.md; report-width.md V6) |
| width | Grapheme segmentation matches current Unicode | ansi.stringwidth-grapheme-default, tcell.cell.put-grapheme | partial | `nextClusterBoundary` segments text (`measure/measure.go:130-195`); loom splits Indic conjunct `क्ष` into two clusters (width 2), while x/ansi and tcell v3 keep one (report-width.md V1) |
| width | Cell buffer stores grapheme clusters | tcell.cell.put-grapheme | partial | `Canvas.setNew` stores a leading cell plus continuations (`canvas.go:360-368`, `:387-394`), using loom segmentation; a wide glyph crossing the right edge is rejected whole (report-width.md V2) |
| input | Partial escape stays buffered until complete | input.parsesequence-consumed | has | `scanKey` buffers incomplete sequences (`event.go:608-657`); `Pane.run` waits for completion or timeout (`pane_run.go:625-695`; report-input.md) |
| input | Query kitty support (`CSI ? u`) before enabling | tcell.kitty-keyboard.enable | partial | only in `loom-probe keys`, not in `Pane`; `setKittyKeyboard` pushes flag 1 (`pane_modes.go:140-165`; report-queries.md) |
| input | In-band resize, mode 2048 | term.inband-resize-2048 | **missing** | Pane uses SIGWINCH (`pane_wait_unix.go:37-40`, `pane_screen.go:91`); `VT.privateModeState` returns 0 for 2048 (`vt/vt.go:735-759`); a 2048 report becomes an empty key (report-queries.md V6). Not ticketed: SIGWINCH works |
| input | Paste sanitised of control chars | tuios.bracketed-paste-sanitize | partial | `textEntryUnits` strips controls (`text_entry.go:1100-1125`); `TextInput.ConsumePaste` keeps ESC and Ctrl+C (`textinput.go:54-61`); TextArea inserts raw, RichTextEdit ignores paste (report-input.md; ticket 464) |
| input | Paste never triggers shortcuts or navigation | tview.application.paste-no-navigation | has | `Pane.run` buffers paste separately (`pane_run.go:625-695`); no shortcut fired through the Pane route; 1 MiB cap drops overflow (report-input.md) |
| input | Kitty event types and associated text (flags 2, 4, 8, 16) | ansi.kitty-keyboard-flags | partial | `decodeCSIu` drops fields after `:` (`event.go:376-400`); `KeyEvent` has no event type. Unreachable today since Pane pushes flag 1 only (`pane_modes.go:140-165`); not ticketed (report-input.md) |
| events | High-rate producers: `Send`/`Post` without blocking | bt.send-blocks | partial | At `e2236900`, `Pane.Post` takes a mutex and appends to an unbounded slice (`pane_run.go:50-62`); it does not wait for the loop. Keys wait behind whole batches (458 b9 canary A, 478) |
| events | Gesture ownership after press (drag capture) | tview.application.mouse-capture is an interceptor, not ownership | partial | `TextEntry.SetFocus` has an inverted capture check (`text_entry.go:338-349`); modal capture takes precedence (`modal_mouse.go:46-95`); selection does not capture (report-mouse.md; ticket 467) |
| events | Drag auto-scroll while held outside content | tuios.drag-autoscroll | **missing** | `mouseCaptureTarget` captures scrollbar and pan (`text_entry.go:1237-1255`); tick traversal is in `widget.go:74-119`; selection produced 0 events over 300 ms held outside (report-mouse.md; ticket 467) |
| events | Horizontal wheel (SGR 66/67) | uv, tcell decode both axes | **missing** | `DecodeMouse` falls through to button handling (`event.go:525-556`); decoded as presses of buttons 2 and 3 (report-mouse.md; ticket 468) |
| events | Child-local 0-based coordinates in nested containers, modal precedence | | has | `Stack.ConsumeMouse` routes local coordinates and modal first (`stack.go:161-177`); nested Stack canary at nonzero origin (report-mouse.md) |
| events | Click-drag as motion while button held | tcell.mouse.drag-motion | has | `DecodeMouse` emits `MouseDrag` (`event.go:525-550`; report-mouse.md) |
| robust | Panic recovery restores terminal | bt.crash.recover | has | `recover()`, close, re-panic in `pane_run.go:136-141`; mode bytes balanced (458 b9 canaries C1, C2) |
| robust | Suspend / release terminal for a child | bt.suspend, bt.terminal.release-restore, tview.application.suspend | has | `Pane.Suspend` hands off and restores the terminal (`pane_suspend.go:166-208`; report-widgets.md) |
| robust | SIGWINCH resize | bt.resize.sigwinch | has | `notifyPaneSignals` subscribes (`pane_wait_unix.go:37-40`); `Pane.applyWinch` re-queries and applies size (`pane_screen.go:91`; report-queries.md V6) |
| robust | Cleanup when start fails before first frame | tuios.run-before-init-gap | partial | `Run` closes on panic (`pane_run.go:130-145`); `Close` restores terminal (`pane_terminal.go:190-245`); caller must close after a Run error or signal (report-testing.md C3–C6; 479) |
| widgets | Fuzzy filter with ranking | bubbles.list.fuzzy-default-filter | has | `Choice.refilter` scores and sorts fuzzy matches (`choice.go:245-283`; report-widgets.md) |
| widgets | Table renders only visible rows | bubbles.table.virtualised-render | partial | `Table.Draw` is viewport-bounded (`table.go:404`, about 90 µs at 100 and 100,000 rows); `refilter` scans all rows (`table.go:241-261`, 3.4 ms and 6.7 ms at 100,000; report-widgets.md) |
| widgets | Container keys do not break a focused multi-line editor | huh form field routing | **missing** | Form takes Tab/Down/PgDn/Enter before the focused TextArea (`form.go:276-297`); canary: Down moves focus, Enter submits (report-widgets.md) |
| widgets | Selection, undo and cut in every text widget | bubbles.textarea.*, tview textarea | partial | `TextEntry.ConsumeKey` handles cut and undo (`text_entry.go:525-553`); `TextInput.ConsumeKey` and `TextArea.ConsumeKey` lack those actions (`textinput.go:77-155`, `textarea.go:240-343`); selection and RichTextEdit comparison in report-widgets.md |
| widgets | Select filter in forms | huh.select.filter | has | `Choice.refilter` filters choices (`choice.go:245-283`; report-widgets.md) |
| widgets | Word jump, Shift-select, copy, select all in text inputs | bubbles.textarea.*, bubbles.textinput.word-forward | has | `TextEntry.ConsumeKey` implements these (`text_entry.go:475-540`); loom uses Ctrl+A/C (Bubbles: Ctrl+G, Ctrl+Shift+C); in TextArea Ctrl+A is line start (report-widgets.md) |
| widgets | Tab accepts a suggestion | bubbles.textinput.accept-suggestion-tab | n/a | loom Tab policy; `TextEntry.ConsumeKey` accepts placeholder only (`text_entry.go:465-485`; report-widgets.md) |
| compose | Z-ordered layers with hit test | lipgloss.layer.z-order, lipgloss.compositor.hit-test | partial | `Canvas.setNew` records claims (`canvas.go:325-345`); `Pane.dispatchMouse` prioritises modals (`pane_dispatch.go:14-46`); overlap uses draw order, not z-index (report-desktop.md) |
| compose | Focus returns after a popup closes | | **missing** | `Popup.ConsumeKey` closes on Escape (`popup.go:77-84`), while Stack routes to focused child (`stack.go:139-151`); focus stays on closed popup until Tab (report-desktop.md; ticket 475) |
| theme | Dark/light background detection (OSC 11, mode 2031) | bt.theme.background-color, lipgloss query | **missing** | `Theme` selects named themes (`theme.go:213-254`); report-desktop.md found no query or report handling, and an OSC 11 reply is typed as text (tickets 476, 462) |
| api | Copy key that does not collide with quit | tview.textarea.copy-is-ctrl-q | has | `TextEntry.ConsumeKey` handles Ctrl+C only with a selection, else passes (`text_entry.go:525-535`; report-widgets.md) |
| api | One filter slot, last wins | tuios.filter-single-slot | n/a | `Pane.dispatchMouse` routes through the widget tree (`pane_dispatch.go:14-46`; report-mouse.md); absence of a separate filter slot is unchecked |
| api | Shared program options across binaries | tuios.program-options-shared | n/a | unchecked: `Pane.run` binds clipboard and ticker (`pane_run.go:109-115`), but report-session.md does not trace binary option wiring |
| session | Daemon with attach/detach, one VT per pane | tuios.session-daemon | partial | session attach snapshots (`internal/session/session.go:1376-1405`), `VT.Repaint` redraws (`vt/vt.go:299-365`); one client, second attach refused; reattach repaints 24 rows, no scrollback or byte-offset resume; no live-stream gap (report-session.md) |
| session | OSC 52 write policy per pane | tuios.osc52-write-policy | **missing** | child OSC 52 writes and queries forwarded raw (`internal/session/session.go:764`, `:783`); TUIOS: off/on/ask/focused (ticket 473) |
| session | Kitty flag stack per screen | tuios.kitty-keyboard-stack | **missing** | one value (`vt/vt.go:58`, `:809-819`), leaks across `?1049`, `=N u` restores the cursor; reattach sends fixed `>1u` (ticket 472) |
| session | VT scrollback API | vt.scrollback-api | has | `SetScrollbackLimit`/`ScrollbackCells` (`vt/vt.go:139-170`); 5000 rows by spec, bound holds at 100,000 lines (report-session.md) |
| model | Leader-key modal design | tuios.leader-key-model | n/a | loom is non-modal by design |
| test | Headless model test harness | bt.test.teatest-wait | has | `loomtest.NewVT` (`loomtest/vt.go:12`), `CompareWidget` (`loomtest/compare.go:23`); teatest polls with a 50 ms sleep; VT split bug 477 |
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
