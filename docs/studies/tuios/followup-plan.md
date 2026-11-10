# Ticket 457: follow-up research plan

## Review findings

Audit: all seven reports and raw YAML files, merged features and summary; Loom `a6287935fbaaabfc4d593ddfa5d7a4a849d0a8aa` (code unchanged since host baseline `163b3738`). Graph coverage flagged stale/new paths; relevant source was read directly. No runtime canaries were run.

- **Confirmed, qualify:** `canvas.go:576` `FlushWithConfig` emits every canvas row per flush; `pane_run.go:299,432` clears/draws/flushes only when dirty. This proves no frame diff, not continuous idle repaint or measured comparative slowness. AtomicFlush batches canvas bytes, not every terminal write in a redraw.
- **Wrong grouping:** `tuios.uv-style-diff` is SGR transition caching, not temporal cell diff: [`frameRenderer.diffs`](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/internal/app/frame_render.go#L9-L32). Trace TUIOS's Bubble Tea renderer before asserting its main output lacks 2026; [`cursedRenderer`](https://github.com/charmbracelet/bubbletea/blob/v2.1.1/cursed_renderer.go) is a separate layer and version from TUIOS's v2.0.10.
- **Wrong width explanation:** no 2027 negotiation was found, but `measure/zwj.go` probes family-emoji cursor advance via CPR; terminal-name selection in `measure/spec.go:191` is a separate render-path heuristic. Canvas grapheme support is evidenced by `Canvas.Write` → `textClusters` → `measure.Clusters`, not editing tests alone. Querying 2027 does not establish Unicode-version/width equivalence.
- **Capability gaps:** Pane pushes kitty flag 1 (`pane_modes.go:147`); `decodeCSIu` has no event-type field and discards alternate-code details. Query current flags is not a supported-bit mask. `CSI =31u` sets flags; `CSI >31u` pushes ([kitty specification](https://sw.kovidgoyal.net/kitty/keyboard-protocol/#progressive-enhancement)); fix the tcell “query then push” wording after tracing call sites.
- **Post overclaim:** `Pane.Post` uses a mutex and unbounded slice; it avoids waiting for the loop, not all blocking. `runPosted` runs a whole batch synchronously and Close can drop remaining callbacks. Queue latency, memory and input fairness remain unmeasured.
- **Capture mismatch:** tview's interception/rewrite hook is not Loom's drag ownership. `mouse_capture.go` is unexported; `TextEntry.mouseCaptureTarget` claims scrollbar/pan, not selection (`text_entry.go:1240`). Selection leaving bounds and stationary outside-drag ticks need separate investigation; “capture exists” is insufficient.
- **Missed existing behavior:** `Table.Draw` renders `r.H-2` data rows; `computeWidths`/`refilter` still scan the dataset. TextEntry paste goes through `insert` → `textEntryUnits`: CR/CRLF normalization, tab expansion and other control removal. Pane routes paste separately, including oversized-paste discard through its terminator; this does not establish every widget's sanitation policy.
- **Missed concrete VT limitation:** `vt/vt.go:810` overwrites one kitty flag value and pop clears it; no nested/per-screen stack there. `privateModeState` replies 0 for 2026 although its reset captures frames. Distinguish test-frame recording, advertised capability and real synchronized display before changing replies. Pane's screen-switch pop/push already exists (`pane_screen.go:238`).
- **Weak conclusions:** “no forks” does not follow from no `replace`; inspect TUIOS VT provenance. gocui enum aliases alone do not prove runtime collision. Lip Gloss v1/v2 can coexist under distinct module paths. A foot “TODO” origin does not prove 2048 lacks a spec. Tab suggestion acceptance is not categorically n/a: Loom explicitly documents placeholder acceptance. Avoid library-wide editing/quality verdicts from one widget or declaration.
- **Thin rows:** recheck `uv.render-flush-split` (method name only), `ansi.width-method` (type declaration), `input.parsesequence-consumed` (not buffering proof), `lipgloss.canvas.compose-order` (broad inference), and tcell enable rows (constants versus execution). `main@unpinned`, “line n/a”, approximate anchors and README-only negatives cannot count as implementation verification. Keep unquoted `vt.widthmethod-mode-2027`, `vt.scrollback-api`, `vt.insert-mode-irm`, colorprofile constants and truncated kitty/Glamour rows unknown; 179 “verified” is a searcher label, not an audit result.

## Shared protocol and source lock

Inputs for every batch: issue `457-*`, `docs/Widgets.md`, all existing study reports/YAML, `features.md`, `summary.md`, and this source lock. Work stays research-only; any canary harness is disposable, outside tracked files. Preserve GUI-like, mouse-enabled, non-modal requirements.
Models for every batch: **haiku:med** fetches pinned files, quotes and inventories; **terra:med** compares code, traces callers and evaluates canaries (required for input/routing); escalate conflicting protocol evidence to **astra:high** once. Do not delegate merely to repeat the original searches.
Verification V (applies to every batch): quote verbatim (raw YAML excerpts ≤120 characters); report excerpts 3–8 relevant lines with a brief annotation explaining mechanism and limit. Use immutable `blob/<SHA>/file#Lx-Ly` permalinks, never repo-root/pkg-index URLs or raw-URL pseudo line fragments. Follow declarations to executing callers; label inference, absence search scope and unrun checks. Store Loom revision, file/symbol/lines, terminal versions, commands, expected/actual bytes and limitations. Respect source quotation limits.
Stop S (every batch): finish each assigned feature as proved/partial/disproved/unknown; cap at two fetch attempts per unavailable source and one canary per behavior matrix entry. Default deadlines: 5 seconds per synthetic check, 2 seconds per live query, 60 seconds per manual trial; record any justified adjustment before running. Leave timeout results unknown with reason; never substitute a guessed quotation or pin. No implementation, tickets or adoption recommendation before evidence.

Flow fixes after the batch 1 trial (apply to every batch):
- **Same matrix on both sides:** run each canary matrix through one shared harness against Loom and against every compared library (pinned), with named measurement boundaries (for render: Canvas flush alone, and Pane including ancillary writes). A Loom-only number is not a comparison.
- **Deployed pins for TUIOS claims:** for any claim about TUIOS behavior, use the dependency versions in TUIOS's pinned `go.mod`/`go.sum`; the newer pins in the table below are separate comparisons.
- **Loom permalinks:** cite Loom as `https://codeberg.org/ubunatic/loom/src/commit/<SHA>/<file>#L<x>-L<y>` at a revision pushed to `origin`; the GitHub mirror is not kept current.
- **Finish the batch:** use the full time budget to close the comparison; "unknown" is only for blocked sources or failed canaries, not for skipped steps.

Exact repository aliases below use `github.com/` unless stated; new HEAD pins were resolved read-only on 2026-11-10. They freeze follow-up inputs, not retroactive proof of original rows. For moving-source rows, record the version change. Resolve module dependencies from the relevant pinned `go.mod`/`go.sum`; do not compare TUIOS's deployment against a different dependency silently.

| Alias | Repository and immutable revision/tag |
|---|---|
| T | `Gaurav-Gosain/tuios@6fe2c0df9922b2eec28f392fee4e196efd86be36` |
| BT | `charmbracelet/bubbletea@v1.3.10`, `@v2.0.10` (T dependency), `@v2.1.1` (separate comparison) |
| B / H | `charmbracelet/bubbles@1deada954e6eead9c054837f1cd3d669e912cd4d` / `charmbracelet/huh@ffb6a97195b4a031925a31cc70c62fc1d309e69d` |
| UV | `charmbracelet/ultraviolet@a45b7c86333f9dfedb95f81064e80259b02937f0` |
| X / XT | `charmbracelet/x@27ecf3e4d454ff2656204b806b7ee6ef11593746` / teatest snapshot `@c83711a11ffab1031f5f055645bb9ac48e2cadc4` |
| CP | `charmbracelet/colorprofile@51fafca7b44cffe4004bd86edd6bd924f84ce0c9` |
| LG / G | `charmbracelet/lipgloss@ff16a4ef6dafa18b1f234a2687b9c35c18658e51` (also T's `v2.0.6`) / `charmbracelet/glamour@010543cac1a07c9bfd9969e58d683f2ddef4c8f7` |
| TV / TC / GC | `rivo/tview@c15b79fa47f27032f26647a75cb5a81720255076` / `gdamore/tcell@4cedad629677a957ea1cda532dbf3fa58ef42a82` / `jesseduffield/gocui@c81715e95462cf95447bbd619e4dc85f3e4e1006` |
| K / F | `kovidgoyal/kitty@0273bc12c3dcc835f4da4a0ea88233ace073706e` / `codeberg.org/dnkl/foot@152487e494993ad7fb5633a27c7d7ad0f8810ea5` |
| C / U | `contour-terminal/vt-extensions@05050a11e793c8f4362bf4e34a59ed3f7e5105fe` / `contour-terminal/terminal-unicode-core@64f53851ceab9a3cf08db4939bcaef75a0899573` |

Each numbered stage below applies per feature, not merely once per library. Verify paths in the pinned tree; record moved paths before quoting. Output paths are relative to this directory. Stage 4 proposes shared-file changes; only the final integrator applies them.

## 1. Render and diff

1. Read Loom `canvas.go` (`Row`, `FlushWithConfig`), `pane_run.go` (`redraw`, dirty gating), `pane_metrics.go`, `resize.go`; prove byte/write counts, style runs, clear/resize/suspend invalidation and inline/alternate origins separately.
2. Read BT `standard_renderer.go` (v1), `cursed_renderer.go` (both v2 tags); UV `terminal_renderer.go` (`transformLine`, hashes), `terminal_screen.go`; T `internal/app/frame_render.go`, `text_sizing_passthrough.go`. Trace string serialization through the final terminal writer.
3. Capture annotated old/new-buffer comparison, unchanged-line skip, SGR memo and synchronization wrapper excerpts under V; compare identical screen updates, not API names.
4. Write `report-render.md`; propose splitting temporal diff/style-diff/atomic-write rows in features and raw YAML and qualifying summary efficiency claims.
- Canaries: 80×24 blank/unchanged/one-cell/style-only/wide-cell/shrink frames; record bytes, writes, allocations and elapsed time, with animation off. Count ancillary Pane writes separately; check inline and alternate screen correctness.
- Model/input/output: shared models and inputs plus render rows; output only `report-render.md`. V/S apply. Parallel: yes; reserve a dedicated PTY for any terminal run. Stop after the fixed matrix, without selecting a diff implementation.

## 2. Width and graphemes

1. Read Loom `measure/{measure,spec,env,zwj,zwj_unix}.go` (`Clusters`, `StringWidth`, `DetectZWJMode`, `DetectRenderPath`), `canvas.go` (`Set`, `Write`, continuation clipping); prove segmentation, width and terminal detection as distinct features.
2. Read X `ansi/{method,width}.go`, `vt/{emulator,csi_mode}.go`; UV `cell.go`; TC `cell.go` (`CellBuffer.Put`) and pinned width dependency; T `vt/mode.go`; U `spec/terminal-unicode-core.tex`; F `doc/foot-ctlseqs.7.scd` 2027 section.
3. Quote cluster storage, width dispatch, query/state mapping and forced-2027 implementations under V; record Unicode tables/dependency versions, not just “grapheme support”.
4. Write `report-width.md`; propose corrected canvas evidence, 2027 statuses and unknown raw `vt.widthmethod-mode-2027`/`ansi.width-method` claims.
- Canaries: combining marks, VS15/16, ZWJ family, flags, Indic clusters and clipped wide cells; compare CPR advance to Loom widths before/after supported 2027, including LOOM_ZWJ overrides. Coordinate with batch 3's query transcript.
- Model/input/output: shared models/inputs plus width rows; only `report-width.md`. V/S apply. Parallel: source work yes; terminal mode changes serialized with batch 3. Stop if terminal/Unicode mismatch remains, preserving both observed results.

## 3. Terminal queries and negotiation

1. Read Loom `pane_run.go` startup/input scheduling, `pane_modes.go`, `event.go` report filtering, `cmd/loom-probe/keys.go`, `vt/vt.go` (`privateModeState`); locate Pane versus probe versus session ownership for 2026/2027/2048, DA and XTVERSION.
2. Read BT v2 `cursed_renderer.go`, `tea.go`; TC `tscreen.go` query callers; X `input/parse.go`, `ansi/mode.go`; F `doc/foot-ctlseqs.7.scd`, C `synchronized-output.md` and pinned tree's 2048 spec, K `docs/keyboard-protocol.rst`. Record a missing spec path as unknown rather than guessing it.
3. Quote query/reply state machine, deadlines, reply consumption and restoration under V; distinguish recognized reset/set/permanent states, unsupported replies and silence. Current kitty flags are not maximum supported enhancements.
4. Write `report-queries.md`; propose 2026/2027/2048 row corrections, negotiation gaps, terminal/version support matrix and removal of the “TODO implies no spec” claim.
- Canaries: DECRQM 2026/2027 replies in installed foot/kitty (record versions), synthetic states 0–4/no reply/delayed split reply, concurrent user bytes, resize reports and timeout recovery. Restore all changed modes; no real-display tearing claim from byte capture alone.
- Model/input/output: shared models/inputs plus terminal rows; only `report-queries.md`. V/S apply. Parallel: source yes; exclusive terminal fixture, shared transcript read-only to width/input/testing. Stop after bounded reply deadlines; unavailable terminals remain unrun.

## 4. Input, kitty and paste

1. Read Loom `event.go` (`scanKey`, `decodeCSIu`, `DecodePaste`, `DispatchPasteEvent`), `pane_run.go` pending/deadline/overflow handling, `pane_modes.go`, `pane_screen.go`, `pane_suspend.go`; inspect each of `text_entry.go`, `textinput.go`, `textarea.go`, `richtextedit_input.go` for paste and shortcut boundaries.
2. Read BT v2 `key.go`, `keyboard.go`, `paste.go`, input decoder dependency at each tag; X `input/parse.go`, `ansi/kitty.go`; TC `tscreen.go`; K `docs/keyboard-protocol.rst`; T `internal/terminal/paste.go` and the `vt.SanitizePaste` definition located from that caller; TV `application.go` paste dispatch.
3. Quote incomplete-sequence consumption, kitty event/alternate/text parsing, push/set/pop and actual sanitization implementations under V; absence requires scoped caller/decoder searches, not a declaration.
4. Write `report-input.md`; propose split capability/request/representation rows, per-widget paste statuses and corrections to raw kitty/split-escape/sanitization rows and summary.
- Canaries: every byte split of CSI/SS3/UTF-8/kitty sequences, lone ESC at timeout boundaries, press/repeat/release with flags 1/2/4/8/16, legacy fallback, nested screen/suspend restoration; paste Ctrl+C/Enter/ESC/CRLF/C1, split terminator and oversized payload followed by a real key. No arbitrary sleeps.
- Model/input/output: haiku:med fetch, **terra:med routing comparison**; shared inputs plus input rows; only `report-input.md`. V/S apply. Parallel: source/synthetic tests yes; live keyboard fixture exclusive. Stop on first demonstrated unsupported representation per feature; do not enable new flags in Loom.

## 5. Mouse and drag

1. Read Loom `pane_run_mouse.go`, `pane_dispatch.go`, `mouse_capture.go`, `modal_mouse.go`, `text_entry.go` (`mouseCaptureTarget`, `ConsumeMouse`), `view.go`, `viewport.go`, `widget.go` ticker traversal; prove ownership, local 0-based translation, focus loss, release/Esc and stationary-outside behavior separately.
2. Read T `internal/app/{update,wheel_axis}.go` and callers of `AutoScrollTickMsg`; BT v2 `mouse.go`, `options.go`; TV `application.go` (`SetMouseCapture` and invocation); TC `mouse.go`, `screen.go`; GC `keybinding.go` and `gui.go` dispatch to distinguish aliases using event/modifier context.
3. Quote interceptor versus ownership routing, held-button versus all-motion tracking, tick continuation and cancellation under V. Do not infer delivery outside an OS window from library capture alone.
4. Write `report-mouse.md`; propose separate interception/public capture/selection capture/autoscroll/wheel rows; correct the global-hook equivalence and unsupported gocui collision claim.
- Canaries: nested nonzero-origin drag across sibling/modal bounds, stationary outside tick, release and Esc rollback, focus loss/re-entry, held-button 1002 versus 1003, wheel axis and burst behavior. Record terminal-delivered events before attributing gaps to Loom.
- Model/input/output: haiku:med fetch, **terra:med routing comparison**; shared inputs plus mouse rows; only `report-mouse.md`. V/S apply. Parallel: headless/source yes; manual pointer fixture exclusive. Stop if the terminal supplies no event; report that boundary rather than inventing an application failure.

## 6. Widgets and editing

1. Read Loom `table.go` (`Draw`, `computeWidths`, `refilter`), `choice.go` (`refilter`, `fuzzyScore`), `filepicker.go`, `textinput.go`, `textarea.go`, `text_entry.go`, `richtextedit*.go`, `form.go`, `pane_suspend.go`; prove each widget's selection/undo/copy/validation/mask/suggestions/editor behavior and documented Tab exceptions.
2. Read B `table/table.go` (`UpdateViewport`), `list/list.go` (`DefaultFilter`), `textinput/textinput.go`, `textarea/textarea.go`, `tree/tree.go`, `filepicker/filepicker.go`; H `field_{input,text,select,multiselect}.go`, `form.go`, `layout.go`; TV `textarea.go`; GC `text_area.go`, `edit.go`, `view.go`.
3. Quote visible-row loops versus whole-data sizing, ranked fuzzy filtering, edit/key dispatch, validation/masking and editor handoff under V. Inspect unreviewed B viewport/progress/help/key components only for assigned existing rows; inventory new candidates separately.
4. Write `report-widgets.md`; propose Table/paste/Tab status corrections and a per-widget editing matrix in features/raw YAML; qualify summary “desktop-like” and absence claims.
- Canaries: 100 versus 100,000 rows at fixed viewport (draw, SetRows, filter and width costs separately); fuzzy rank/ties; grapheme selection/undo; Ctrl+C with/without selection; Tab acceptance versus focus; masked display versus stored/copied value; child-editor restore.
- Model/input/output: shared models/inputs plus widget rows; only `report-widgets.md`. V/S apply. Parallel: yes with separate fixtures; input batch owns paste transport, widgets own consumer semantics. Stop after listed behaviors, not a redesign or new widgets.

## 7. Session and VT

1. Read Loom `internal/session/{session,server,client,ipc}.go`, `vt/vt.go` query/CSI/OSC/alternate paths, `pane_screen.go`, `clipboard.go`; prove attach snapshot/stream ordering, resize/2048 ownership, clipboard forwarding policy and kitty nested/per-screen state.
2. Read T `internal/session/{daemon_stream,client,terminal_state}.go`, `docs/REHYDRATION.md`, `vt/snapshot/{capture,apply,replay}.go`, `vt/{kitty_keyboard,mode}.go`, `internal/app/clipboard_osc52.go`; X `vt/{emulator,csi_mode}.go`. Trace snapshot and PTY writer call sites from those files; inspect T VT provenance notices against the named upstream revision.
3. Quote snapshot/stream boundary and replay, push/pop stacks, capability replies and clipboard gate under V; compare emulator promises with transport and outer-terminal behavior separately.
4. Write `report-session.md`; propose resolved session/stack/OSC52/VT-scrollback rows and replace “no forks” with scoped dependency and provenance findings in summary.
- Canaries: detach/reattach during output and resize; no lost/duplicated bytes across snapshot boundary; nested kitty pushes/pops on both screens; DECRQM replies versus implemented effects; focused/unfocused OSC52 and query handling; bounded scrollback.
- Model/input/output: shared models/inputs plus session rows; only `report-session.md`. V/S apply. Parallel: source/synthetic yes with isolated sockets/processes; do not reuse owner sessions. Stop if transport evidence is missing; no claim that a field name proves persistence.

## 8. Theming, composition and desktop integration

1. Read Loom `theme.go`, `style.go`, `ansi/style.go` (`DetectColorProfile`), `canvas.go` claim/surface composition, `stack.go`, `popup.go`, `frame.go`, `richtext.go`, `clipboard.go`, `pane_dispatch.go`; map semantic roles, composition, hit/focus routing, Markdown and available focus/theme/clipboard events per feature.
2. Read LG `layer.go`, `canvas.go`, `UPGRADE_GUIDE_V2.md`; G `glamour.go`, `styles/styles.go`; CP `env.go` and profile definitions located in its tree; BT v2 `color.go`, `focus.go`, `clipboard.go` after path verification; T `internal/markdown/{markdown,inline}.go`, `internal/layout/{bsp,master,scrolling}.go`; F ctlseqs 996/2031/OSC11/52/8/22 sections and K clipboard/graphics protocol docs.
3. Quote z-flattening/hit dispatch versus focus/claim handling, environment versus queried color, Markdown rendering, clipboard target/policy and notification handling under V; do not transplant modal leader-key behavior into Loom goals.
4. Write `report-desktop.md`; propose corrected Lip Gloss coexistence/background messages, colorprofile/Glamour unknowns, roles/Markdown comparison and P2/P3 terminal rows in features/YAML/summary.
- Canaries: nested Z with overlapping siblings and transparent blanks; click-through/focus; fixed color-profile snapshots; delayed OSC11/theme/focus reports interleaved with keys; OSC52 target distinctions. Record unsupported terminal protocols without implying SDK absence.
- Model/input/output: shared models/inputs plus theming/desktop rows; only `report-desktop.md`. V/S apply. Parallel: headless/source yes; query/clipboard fixtures serialized with batches 3/7. Stop at existing row scope plus named missing comparisons; no new desktop integration implementation.

## 9. Testing, lifecycle and producer load

1. Read Loom `loomtest/{vt,compare}.go`, `internal/ptytest/session.go`, `pane_{post,kitty,suspend,pty}_test.go`, `mouse_capture_test.go`, `pane_run.go` (`Post`, `runPosted`, recovery), `pane_terminal.go` startup/close and callers; prove cleanup ownership and distinguish Close-required errors from panic restoration.
2. Read XT `exp/teatest/v2/teatest.go`; BT `tea.go` (`Send`, startup, recovery, release/restore); GC `gui.go` (`Update`, `UpdateAsync`); T `vt/ghostty_differential_test.go`, `vt/vtgen/` generators and `internal/app/frame_render_test.go` after locating its exact test path; follow T's early-start cleanup wrapper and cited upstream issue at the pinned revision.
3. Quote wait/deadline synchronization, queue delivery/backpressure, error cleanup and differential oracle mechanisms under V. Keep source-scan tests, visual/golden tests, VT semantics and live-terminal tests distinct.
4. Write `report-testing.md`; propose Post partial/load limits, early-error cleanup status, test-harness capabilities and evidence-qualified quality conclusions in features/YAML/summary.
- Canaries: producer burst with concurrent input (queue depth, heap, callback latency, frame count); slow callback and Close mid-batch; injected startup write/read failures and Draw panic; compare terminal modes before/after; replay the same chunked stream in Loom VT and a pinned reference emulator. Use deadlines/channels, no timing sleeps.
- Model/input/output: shared models/inputs plus robustness/test/event rows; only `report-testing.md`. V/S apply. Parallel: isolated tests yes; never compete with live query/input fixtures. Stop on first reproducible lifecycle discrepancy, record it without fixing or filing a ticket.

## Integration and completion

Run source work for all nine batches in parallel only when explicitly dispatched; each owns its uniquely named report, with shared source locks read-only. One scheduler serializes terminal/clipboard fixtures. Cross-batch references are report links, not edits to another batch's output.
One stronger integrator serially applies stage-4 proposals to `features.md`, every affected `features-*.yaml` **loom column** (preserve row IDs/source history; annotate corrected evidence), and `summary.md`; resolve duplicate rows by feature semantics, not library names. Do not have workers concurrently edit raw YAML or merged files.
Recount rows and verification levels; reject nonverbatim/overlong quotes, broken or moving evidence links, declaration-only proofs and unsupported negatives. Every filled Loom status must have revision/file/symbol evidence; every remaining unknown/unrun canary stays explicit. Rank follow-up value by measured benefit and Loom requirements; end research when these checks pass or all unresolved rows have bounded, documented blockers.
