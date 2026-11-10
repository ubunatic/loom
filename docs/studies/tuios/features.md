# TUIOS study: feature list (ticket 457)

Merged from `features-*.yaml` (189 rows from 7 searchers, 179 verified by quote).
The `loom` column was filled by the host from loom code at `163b3738`, only for P1 rows.
P2 and P3 rows stay in the YAML files with `loom: unknown`; check them when a ticket needs them.

Legend: **has**, **partial**, **missing**, **n/a** (not wanted for loom's GUI-like model).
"UNVER" = the searcher could not quote the source.

## P1 gaps and partials (ticket candidates)

| Area | Feature | Where seen | loom | Evidence in loom |
|---|---|---|---|---|
| render | Cell or row diff: write only changed cells/rows | uv.transformline-diff, uv.line-hash-diff, bt.renderer.cell-based, tuios.uv-style-diff | **missing** | `canvas.go` `FlushWithConfig` writes every row on every frame (`pane_run.go:299`) |
| render | Render into a buffer, one write per frame | uv.render-flush-split | has | `ResizeConfig.AtomicFlush` in `canvas.go` |
| render | Mode 2026 synchronized output | bt.sync-output-2026, uv.sync-output-2026, term.sync-output-2026 | has | `ResizeConfig.SynchronizedOutput` |
| render | Query DECRQM 2026 before using it | term.decrqm-2026-replies | unchecked | config switch only, no query found |
| width | Mode 2027 grapheme width, queried by DECRQM | term.grapheme-2027, ansi.width-method, vt.widthmethod-mode-2027 (UNVER), tuios.width-mode-2027 | **missing** | no `2027` in loom; `measure` probes ZWJ width by terminal instead |
| width | Cell buffer stores grapheme clusters | tcell.cell.put-grapheme | has | cluster-aware editing (`text_cluster_edit_test.go`) |
| input | Partial escape stays buffered until complete | input.parsesequence-consumed | has | `esc_key_timeout` 50 ms (`docs/TuiInput.md:56`) |
| input | Kitty keyboard flags beyond 1 (event types, alternate keys) | ansi.kitty-keyboard-flags, term.kitty-keyboard.flags (UNVER) | partial | loom pushes flag 1 only (`docs/TuiInput.md:96`) |
| input | Query kitty support (`CSI ? u`) before enabling | tcell.kitty-keyboard.enable | partial | only in `loom-probe keys`, not in `Pane` (opt-in flag) |
| input | In-band resize, mode 2048 | term.inband-resize-2048 | **missing** | only in `internal/session`; Pane uses SIGWINCH |
| input | Paste sanitised of control chars | tuios.bracketed-paste-sanitize | unchecked | |
| input | Paste never triggers shortcuts or navigation | tview.application.paste-no-navigation | has | `ConsumePaste` path (TextEntry) |
| events | High-rate producers: `Send`/`Post` without blocking | bt.send-blocks | has | `Pane.Post` appends under a mutex and invalidates; no block, but no cap |
| events | Application-wide mouse capture hook | tview.application.mouse-capture | has | pane drag capture, `ModalTarget` |
| events | Drag auto-scroll while held outside content | tuios.drag-autoscroll | **missing** | capture exists, no autoscroll tick |
| events | Click-drag as motion while button held | tcell.mouse.drag-motion | has | `MouseDrag` |
| robust | Panic recovery restores terminal | bt.crash.recover | has | `recover()` in `pane_run.go:137` |
| robust | Suspend / release terminal for a child | bt.suspend, bt.terminal.release-restore, tview.application.suspend | has | `Pane.Suspend` |
| robust | SIGWINCH resize | bt.resize.sigwinch | has | `docs/TuiInput.md:35` |
| robust | Cleanup when start fails before first frame | tuios.run-before-init-gap | unchecked | check `Pane.Run` error paths |
| widgets | Fuzzy filter with ranking | bubbles.list.fuzzy-default-filter | has | `Choice.Fuzzy`, FilePicker |
| widgets | Table renders only visible rows | bubbles.table.virtualised-render | unchecked | Table is sortable and filterable; virtualisation not checked |
| widgets | Select filter in forms | huh.select.filter | has | Choice filter |
| widgets | Word jump, Shift-select, copy, select all in text inputs | bubbles.textarea.*, bubbles.textinput.word-forward | has | TextEntry; loom uses Ctrl+A/C (Bubbles: Ctrl+G, Ctrl+Shift+C) |
| widgets | Tab accepts a suggestion | bubbles.textinput.accept-suggestion-tab | n/a | loom Tab policy; TextEntry accepts placeholder only |
| compose | Z-ordered layers with hit test | lipgloss.layer.z-order, lipgloss.compositor.hit-test | has | root overlays and claim-cell routing |
| api | Copy key that does not collide with quit | tview.textarea.copy-is-ctrl-q | has | Ctrl+C copies only with a selection, else passes (`HasSelection`) |
| api | One filter slot, last wins | tuios.filter-single-slot | n/a | loom routes by widget tree |
| api | Shared program options across binaries | tuios.program-options-shared | n/a | |
| session | Daemon with attach/detach, one VT per pane | tuios.session-daemon | partial | `loom session` + `vt.VT`; attach compare open |
| session | OSC 52 write policy per pane | tuios.osc52-write-policy | unchecked | |
| session | Kitty flag stack per screen | tuios.kitty-keyboard-stack | unchecked | `vt.VT` |
| session | VT scrollback API | vt.scrollback-api (UNVER) | has | `Scrollback` in `pane_screen.go` |
| model | Leader-key modal design | tuios.leader-key-model | n/a | loom is non-modal by design |
| test | Headless model test harness | bt.test.teatest-wait | has | `loomtest.VT`, `CompareWidget`; teatest polls with sleep, loom forbids that |
| status | Ultraviolet API still unstable | uv.status-active-dev | n/a | quality note |

## P2 and P3 highlights (not verified against loom)

- Focus in/out reporting (mode 1004): loom enables it only in `internal/session`, not in Pane.
- Dark/light detection: DSR 996, mode 2031, OSC 11 query; Lip Gloss `HasDarkBackground`.
- OSC 52 clipboard read and primary selection (loom writes only).
- XTVERSION and DA1/DA2 terminal identification (loom: only in `loom-probe` and `vt`).
- Text input: suggestions, validation, password echo; Huh: external `$EDITOR`, accessible mode,
  grid and column form layouts, multi-select limit.
- Key press/release and repeat events (kitty flag 2); wheel gesture axis lock (TUIOS).
- Frame rate cap (Bubble Tea 60 FPS default, 120 max).
- P3: OSC 8 links, OSC 9/777 notifications, OSC 22 pointer shape, OSC 133, kitty graphics, sixel.
