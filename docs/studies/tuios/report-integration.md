# Integration audit (ticket 458, M2)

## Changes

- `features.md`: merged duplicate kitty row; qualified Post locking, z-order and audit limits.
- `features-bubbles.yaml`, `features-bubbletea.yaml`, `features-charmx.yaml`, `features-contrast.yaml`, `features-lipgloss.yaml`, `features-terminals.yaml`, `features-tuios.yaml`: updated P1 loom fields only, preserving source history and searcher verification labels.
- `summary.md`: restored the gocui row; corrected renderer comparisons, canary scope and labels; ranked study tickets by estimated benefit and added counts.
- This report records the checks and remaining gaps. Batch reports remain read-only.

## Recount

| Inventory | has | partial | missing | unknown | n/a | total |
|---|---:|---:|---:|---:|---:|---:|
| Merged feature rows | 16 | 13 | 12 | 0 | 5 | 46 |
| YAML loom fields | 23 | 15 | 12 | 134 | 5 | 189 |
| YAML searcher verification labels | 177 true | 3 partial | 9 false | 0 absent | — | 189 |

P2/P3 loom fields remain unknown by scope. Six P1 fields remain unknown: `bt.model-view-returns-struct`, `vt.emulator-api`, `tcell.screen.mouse-motion-flag`, `tcell.screen.paste`, `gocui.keys-and-mouse-share-keycodes`, `term.kitty-keyboard.legacy-exceptions` (no matching merged behavior or pinned trace); see their loom comments. The rest of the 134 unknown fields are P2/P3 and were not checked. Counts do not turn searcher labels into independent verification.

Merged claims still unchecked:

- Single filter slot: `pane_dispatch.go:14-46` shows mouse routing, not absence of a separate filter slot.
- Shared program options: `pane_run.go:109-115` binds clipboard and ticker; report-session.md does not trace binary option wiring.

Leader-key design and API-stability policy are design assessments, not implementation findings.

## Loom source spot checks

Each range below was read from `git show e2236900:<file> | sed -n` (ok means the bounded observation was confirmed):

1. `canvas.go:560-605` FlushWithConfig serializes every row: ok.
2. `canvas.go:325-350` setNew records claims: ok.
3. `pane_run.go:45-81` Post locks and appends: fixed prior nonblocking claim.
4. `pane_run.go:130-145` run recovers and closes on panic: ok.
5. `pane_run.go:425-437` redraw is dirty-gated: ok.
6. `pane_run.go:625-695` paste and input buffering: ok.
7. `pane_modes.go:140-165` kitty flag 1 request: ok.
8. `event.go:376-400` decodeCSIu ignores alternate fields: ok.
9. `event.go:525-556` DecodeMouse horizontal wheel fallthrough: ok.
10. `event.go:608-657` scanKey incomplete sequences: ok.
11. `text_entry.go:338-349` SetFocus inverted capture check: ok.
12. `text_entry.go:465-485` placeholder Tab path: ok.
13. `text_entry.go:525-535` Ctrl+C selection behavior: ok.
14. `text_entry.go:1100-1125` textEntryUnits control filtering: ok.
15. `text_entry.go:1237-1255` mouse pan capture: ok.
16. `textinput.go:54-61` ConsumePaste: ok.
17. `textarea.go:81-96` ConsumePaste: ok.
18. `form.go:270-297` ConsumeKey precedence: ok.
19. `table.go:241-261` refilter scans rows: ok.
20. `pane_dispatch.go:14-46` modal priority: ok.
21. `modal_mouse.go:46-95` modal coordinates and ownership: ok.
22. `widget.go:74-119` tick traversal: ok.
23. `vt/vt.go:139-170` scrollback API: ok.
24. `vt/vt.go:798-854` kitty flag handling: ok.
25. `internal/session/session.go:755-790` forwards child frame: ok.
26. `loomtest/vt.go:1-18` VT wrapper: ok.
27. `loomtest/compare.go:20-32` CompareWidget: ok.
28. `pane_terminal.go:190-205` Close drops posts: ok.
29. `vt/vt.go:730-765` privateModeState returns 0 for 2027: ok.
30. `text_entry.go:475-540` word/select/copy paths: ok.
31. `canvas.go:480-560` Row style transitions: ok.
32. `pane_suspend.go:145-210` terminal handoff: ok.
33. `pane_wait_unix.go:30-45` notifyPaneSignals registers SIGWINCH: fixed missing citation.
34. `pane_screen.go:75-105` applyWinch re-queries size: fixed missing citation.

## Upstream excerpts

Checked `report-widgets.md` U1–U11 against the b6 shallow clones at their recorded release SHAs: U2, U5, U8, U9, U10 and U11 match their contiguous cited lines. U1 (183–186), U3 (199–201), U4 (207–209), U6 (222–223) and U7 (229–231) are rejected as contiguous verbatim quotations: they omit or reorder intervening lines, although the individual lines appear in the sources. Treat them as separate fragments, not as a literal continuous excerpt. Other upstream quotations were not rechecked. The long fenced block at `report-width.md:215-271` is an owner diagnostic script, not a quotation.

## Rejected moving links

These report links point to moving branches and are rejected as pinned evidence; the reports retain their history:

- `report-lipgloss.md:67` https://github.com/charmbracelet/lipgloss/blob/main/UPGRADE_GUIDE_V2.md
- `report-terminals.md:111` https://codeberg.org/dnkl/foot/raw/branch/master/doc/foot-ctlseqs.7.scd
- `report-terminals.md:112` https://github.com/contour-terminal/vt-extensions/blob/master/synchronized-output.md
- `report-terminals.md:113` https://github.com/contour-terminal/terminal-unicode-core/blob/master/spec/terminal-unicode-core.tex
- `report-charmx.md:116` https://github.com/charmbracelet/ultraviolet/blob/main/terminal_renderer.go
- `report-charmx.md:117` https://github.com/charmbracelet/ultraviolet/blob/main/terminal_screen.go
- `report-charmx.md:118` https://github.com/charmbracelet/ultraviolet/blob/main/key.go
- `report-charmx.md:120` https://github.com/charmbracelet/x/blob/main/ansi/method.go
- `report-charmx.md:121` https://github.com/charmbracelet/x/blob/main/ansi/width.go
- `report-charmx.md:122` https://github.com/charmbracelet/x/blob/main/ansi/kitty.go
- `report-charmx.md:123` https://github.com/charmbracelet/x/tree/main/vt
- `report-charmx.md:124` https://github.com/charmbracelet/x/blob/main/vt/emulator.go
- `report-charmx.md:125` https://github.com/charmbracelet/x/blob/main/vt/csi_mode.go
- `report-charmx.md:126` https://github.com/charmbracelet/x/blob/main/cellbuf/screen.go
- `report-charmx.md:127` https://github.com/charmbracelet/x/blob/main/term/term.go
- `report-charmx.md:128` https://github.com/charmbracelet/x/blob/main/input/parse.go
- `report-charmx.md:130` https://github.com/charmbracelet/colorprofile/blob/main/env.go

## Limits

The two API claims above remain unchecked. Full semantic reconciliation and checking every cited ticket and quote are incomplete. `harnez find issues` showed 459–462 and 464–480 open, 463 draft, 433 and 259 open; issue files show 392 and 456 open. The wider cited ticket set was not completely checked. Real-terminal probes and new canaries were not run. `git diff --check` passed. No loom session was touched.
