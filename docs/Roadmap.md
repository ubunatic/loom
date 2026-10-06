---
title: Loom Roadmap
weight: 20
---

# Loom Roadmap

Loom is a **library**, not an application. Its value is measured by what a
*host program* can do with it: how little ceremony it takes to construct,
theme, compose and run a widget; how correct the compositor is under nesting;
and how honest the public API surface is. The example apps in `examples/` and
the widget gallery (`loom widgets --show`) are not the product — they are the
proof that the library contract holds.

Loom's value is reusable, dependency-light inline terminal UI: applications
such as the Harnez/Voxi reference dashboards should share widgets without
taking over the full terminal. An application should declare its UI structure
and presentation, while Go supplies behavior and changing data. The UI should
be easy to maintain, responsive to terminal size, testable without a live
terminal, and free from application-specific coordinate/layout code.

The primary reference targets are documented in
[`HarnezUsageTarget.md`](HarnezUsageTarget.md) (usage dashboard and Voxi monitor)
and [`HarnezSplashTarget.md`](HarnezSplashTarget.md) (startup loading and splash
screen). Since the last pass, `RichTextEdit` (254 onward) became the most
demanding in-repo host: its workaround audit ([273](../issues/273-audit-richtextedit-workarounds-and-plan-loom-library-fixes.md)) showed which library
contracts still push work onto callers.

The lens for this pass is the project rule **"Fix the library, not the
caller"**: when a widget or demo needs local modal state, manual coordinate
math or focus side effects, the library contract is incomplete, and fixing it
there is worth more than another widget.

Stage numbers below preserve the original plan and remain as historical
acceptance context. The Now/Next/Later sequence is the live priority signal and
supersedes stage order wherever the two disagree.

## Shipped since the last roadmap pass (2026-09-29)

Almost the whole previous Now and Next list shipped, mostly through the gap
roadmap [180](../issues/180-roadmap-close-feature-gaps-found-in-the-framework-comparisons.md) built from the framework comparisons 177 to 179.

- Gap roadmap 180, widgets and primitives: 154 ticker cadence, 155 Dialog, 156 SubCanvas/Blit, 158 KeyMap, 160 media playback, 167 ProgressBar, 168 Tree, 169 Form, 170 Menu/MenuBar, 172 FilePicker, 173 password mask, 174 NumberInput/Toggle, 175 DatePicker, and 181 to 194 (bracketed paste, color profile downsampling, key help, TextInput scrolling, TextArea height, Spinner, Timer/Stopwatch, Paginator, fuzzy Choice, Viewport, Table cell cursor, Chart, submenus, OSC 52 copy).
- Widget gallery and one event contract: 176 and 195 to 207 (`loom widgets --show`, vertical tabs, mouse, theme switch, app options), 209 one `EventResult` contract for keys and mouse, 196 and 211 dispatcher/wrapper checks, 210 `loom info`, 212 to 223 gallery bug sweep, 228 to 233 (All tab, F8/F9/F10, astra on redraw), 237, 239, 248 Shift-arrow Grid navigation, 249 Grid cell borders, 250 `SearchBar`, 251 nested focus and cursor propagation.
- RichTextEdit: 254 to 258, 260, 261, 263, 264, 268 to 273, and 278 (key-cap hint bar, now the library `HintBar`). 273 produced the ranked library plan 277, 275, 274, 276 that drives this pass.
- Media, release, module: 202, 226, 227, 230 media scaling/zoom/pan; 208 and 216 PaintCanvas; 234, 235 module path `ubunatic.com/loom`; 267 v0.3.0 README; 224 bash lint.
- 262 closed as a duplicate of 247.

Earlier passes: 162 to 166, 171, 167 M1 (2026-09-29); 141 to 146, 148 to 153,
142, 159 (2026-09-28); 138, 140; 089, 091, 106, 134; 065, 042, 111, 103; 064,
124, 128, 130, 131; 063, 099, 101, 105, 129; 117 to 123; 034, 037, 051, 060 to
062, 081, 088, 092 to 097, 107, 108; 036, 038, 057 to 059, 085. See the stage table.

### Closed by the backlog triage (2026-10-06)

Before the cleanup sprint every open ticket was checked against the current code.
Eleven were already done or obsolete and are closed: 236 (F10 Quit truncation,
fixed by 278's `HintBar`), 243 (cell background, M1 to M4, last `c588df6`),
244 (`c3b3f76`), 245 (`297f222`), 246 (`8c1ccfd`), 253 (`7a35858`, two green
`make test-q1` runs), 098 (superseded by 233), 177 to 179 (consumed by 180),
and 048 (width model specced 2026-09-24; the DE flag finding no longer reproduces).

## What moved since the last pass

- **The current sprint is a cleanup sprint** (user decision 2026-10-06): fix verified bugs and bring the repo into better shape before new features. Every ticket in it was re-checked against the code on 2026-10-06; re-check again before starting work on any older ticket.
- 052 (unused `ChoiceStyle.Border`), 100 (ansiviewer top bar), 225 (split large files) and 252 (spec defaults audit) move Later to Now as repo-shape work.
- 238 (Grid rows sized to content) and 240 (All-tab Table cell selection) move Next to Now as verified bugs.
- 136, 247, 135, 242, 277 and 275 stay in Now; 253, 243, 246, 236, 244, 245 left the roadmap (closed above).

---

## Themes in the open backlog

The 31 open tickets fall into six themes:

1. **Test and terminal reliability** (247, 136, 135, 133): a flaky PTY test, a startup-resize race, an environment-leaking test, a debug overlay.
2. **Event routing, focus and overlays** (242, 277, 275, 274, 276, 238): the library contract that 273 found incomplete.
3. **Repo shape** (225, 252, 052, 240, 100): large files, unspecced defaults, a dead style field, demo drift, an ANSI replay gap.
4. **RichTextEdit as a tool** (279, 265, 266, 259).
5. **Visual testing and CLI asset tools** (157, 095, 139, 137, 147, 102).
6. **Extensions and presentation** (014, 090, 116, 120, 161, 241).

---

## Now: cleanup sprint — verified bugs, then repo shape

**Rationale.** The backlog check showed that a third of the open tickets were
already done, so the tracker itself had drifted. This sprint fixes only bugs
confirmed in the current code and the structural debt that slows every later
change, before any new feature. The flaky test goes first: under Quota-1 one red
run blocks closing any ticket. Event-routing tickets start on `codex:sol:med`
(AGENTS.md).

Bugs, in order:

1. [247](../issues/247-flaky-testpanefirstdrawusesscreenbounds-under-full-test-run.md) (P2, Bug) + [136](../issues/136-pane-misses-a-resize-that-happens-during-startup-foot-app-draws-at-initial-size.md) (P2, Bug): the first-draw PTY test sometimes sees width 99 instead of 100; 136 is the same code path. Verified 2026-10-06: `New()` still reads the size before `installSignalHandler`, and `run` re-queries only for `full && !alt`. 136 has a host-reviewed M1 (red test) / M2 (re-query after handler install) plan.
2. [135](../issues/135-testbrowserusesanimatedbackground-fails-when-loom-evidence-1.md) (P3, Bug): reproduced 2026-10-06: `LOOM_EVIDENCE=1 go test ./examples/filebrowser/...` fails `TestBrowserUsesAnimatedBackground`. Set the variable in the test with `t.Setenv`.
3. [242](../issues/242-gallery-all-tab-dialog-demo-leaks-left-right-arrow-keys-to-the-grid.md) (P2, Major): `Dialog.ConsumeKey` still returns `Ignored()` after moving the selection, so arrows also move Grid focus. Same consume rule as 253.
4. [240](../issues/240-gallery-all-tab-table-selects-the-full-row-table-tab-selects-cells.md) (P3, Bug): the All tab builds its own Table without `CellCursor`; share one constructor with the Table demo.
5. [052](../issues/052-resolve-unused-choicestyle-border-contract.md) (P2): `ChoiceStyle.Border` is configurable but never drawn; decide remove or implement.
6. [238](../issues/238-gallery-all-grid-borders-and-content-fitted-row-heights.md) (P2): Grid borders shipped in 249; rows sized to their tallest widget remain. Grid owns its cell rectangles, so mouse routing follows them without waiting for 276.
7. [100](../issues/100-investigate-full-width-ansi-top-bar-background-in-ansiviewer.md) (P3): ansiviewer top-bar background does not reach the right edge; not re-checked on 2026-10-06, so reproduce first.

Repo shape:

8. [225](../issues/225-modularize-high-loc-core-components-and-example-packages.md) (P3): split the largest files (`richtextedit.go` about 2,450 lines, `pane.go` about 1,600, plus `frame.go`, `ansibuffer.go`, `yaml.go`). Pure moves, no behaviour change.
9. [252](../issues/252-audit-library-defaults-that-are-missing-from-the-spec.md) (P2): library defaults missing from the spec.
10. [277](../issues/277-standardize-click-only-focus-and-cursor-invariants-across-compound-widgets.md) (P2, 273 rank 1): hover never moves focus, field activation or the text cursor.
11. [275](../issues/275-decouple-focus-state-updates-from-popup-and-widget-draw-rendering.md) (P2, 273 rank 2): `Popup.Draw` calls `SetFocus`; move focus changes to open/close and navigation. Prerequisite for 274.

---

## Next: shared overlay layer, container mouse model, then RichTextEdit as a tool

**Rationale.** 274 and 276 finish 273's plan and depend on the sprint (274 needs
275's explicit focus transitions). The RichTextEdit tickets wait for 274 so their
popups use the overlay layer instead of more local modal fields.

- [274](../issues/274-overlay-and-modal-layer-for-canvas-to-eliminate-local-popup-stacks.md) (P2, 273 rank 3): overlay and modal layer on Canvas/Pane; generalize the hook pattern in [`RootOverlays.md`](RootOverlays.md) and migrate RichTextEdit's `helpPopup`, `savePopup`, `savePicker`.
- [276](../issues/276-container-owned-coordinate-translation-and-mouse-event-clipping.md) (P2, 273 rank 4): containers translate and clip mouse events; find a current caller doing its own math first (the gallery RTE case was removed 2026-10-06).
- [279](../issues/279-add-loom-edit-command-to-open-a-file-in-richtextedit.md) (P3): `loom edit <file>`; F10 asks Save/Discard/Cancel on unsaved changes. After 274 so the prompt is an overlay.
- [265](../issues/265-richtextedit-right-click-context-menu-and-clear-style-popover-button.md) (P2): RichTextEdit right-click menu and clear-style button. After 274.
- [157](../issues/157-ansi-golden-mockup-visual-test-comparator.md) (P2): golden ANSI canvas comparator (`loomtest`) with cell-level diffs.
- [133](../issues/133-pane-debug-mode-with-ruler-overlay-shift-f12-loom-debug.md) (P2): pane debug mode with a ruler overlay; after 276.

---

## Later: tooling, ANSI core, extensions

- [266](../issues/266-richtextedit-s-f5-box-push-drawing-mode.md) (P3): RichTextEdit box-push drawing mode (S-F5); after 265.
- [259](../issues/259-osc-8-terminal-hyperlink-support-human-assisted.md) (P2): OSC 8 hyperlinks; needs human terminal probes.
- [241](../issues/241-add-public-grab-package-sdl3-transparent-input-grabber-and-examples-grabber-demo.md) (P2): public `grab` package and `examples/grabber`; self-contained, can run in parallel.
- [161](../issues/161-align-media-example-with-redesign-005.md) (P2): media example to redesign 005.
- [095](../issues/095-add-human-observable-pty-test-view-mode-and-feedback-flow.md) (P1): human-observable PTY test view mode; re-check scope after 157.
- [139](../issues/139-add-loom-play-for-interactive-tui-commands-and-ansi-capture.md) (P2): `loom play`; needs a reuse-vs-build decision.
- [137](../issues/137-feedback-cli-asset-tools-workflow-and-multi-box-ansi-validation.md) (P3) + [147](../issues/147-detect-unclosed-boxes-and-prioritize-box-validation-errors.md) (P2): CLI asset-tool feedback and unclosed-box detection.
- [116](../issues/116-extract-zero-alloc-ansi-styling-and-parsing-into-dedicated-ansi-subpackage.md) (P2): zero-alloc `ansi` subpackage; performance only.
- [014](../issues/014-configurable-graph-colors-and-glyph-presentation.md) (P2): configurable graph colors and glyphs.
- [090](../issues/090-support-image-backed-app-backgrounds-and-background-theme-switching.md) (P1): image-backed app backgrounds; no host has asked.
- [120](../issues/120-implement-wazero-backed-tree-sitter-syntax-engine-with-embedded-grammars-and-queries.md) (P2): wazero Tree-Sitter engine; park candidate at the next pass.

---

## Standing (user-owned)

- [102](../issues/102-human-review-collection-manual-checks-for-lean-sprint-deliveries.md) (P2): manual checks no headless test can settle, plus the roadmap-180 host-choice queue (H1 to H5).

---

## Dependency map

```mermaid
graph LR
  T136[136 startup resize] --- T247[247 flaky first draw]
  T247 --> T135[135 LOOM_EVIDENCE test]
  T242[242 Dialog arrows] -. same consume rule as 253 .- T240[240 All-tab Table]
  T277[277 click-only focus] --> T274[274 overlay layer]
  T275[275 focus out of Draw] --> T274
  T274 --> T279[279 loom edit]
  T274 --> T265[265 RTE context menu]
  T265 --> T266[266 box-push mode]
  T276[276 container mouse translation] --> T133[133 debug ruler]
  T225[225 split large files] -. ansibox.go .- T147[147 unclosed boxes]
  T137[137 CLI feedback] --> T147
  T157[157 golden comparator] -. re-check .- T095[095 PTY view mode]
  T116[116 ansi subpackage] -. may absorb .- T100[100 ansiviewer top bar]
  T274 --> T252[252 spec key bindings]
```

---

## Stage table

| Stage | Tickets |
|---|---|
| 1–5 — Shipped Foundation | [006](../issues/006-static-declarative-monitor-shell-with-a-minimal-validated-contract.md), [007](../issues/007-clock-watch-mode-with-independent-collection-and-redraw.md), [008](../issues/008-responsive-declared-box-layout.md), [009](../issues/009-declarative-box-visibility-controls.md), [010](../issues/010-geometry-and-visual-evidence-milestone-before-rich-content.md) |
| Hygiene & Fixes (Shipped) | [021](../issues/021-cmd-help-tty-blocking-in-tests.md), [022](../issues/022-decouple-static-shell-golden-from-evolving-monitor-spec.md), [023](../issues/023-rows-schema-validation-and-negative-controls.md), [026](../issues/026-investigate-monitor-pty-smoke-test-idle-redraw-regression.md), [047](../issues/047-remove-deprecated-uzu-brand-and-default-global-commands-from-cmdbar.md), [036](../issues/036-keyevent-ergonomic-helpers-e-name-and-e-is-for-unified-key-text-matching.md), [038](../issues/038-fix-multi-byte-utf-8-string-truncation-in-popup-title-and-borders.md), [059](../issues/059-fix-mouse-coordinate-convention-mismatch-between-frame-and-tabs-stack-grid.md) |
| 6 — Rows & Graphs (Shipped) | [011](../issues/011-aligned-dashboard-rows-and-ansi-safe-truncation.md), [020](../issues/020-review-ticket-011-rows-and-target-state-alignment.md), [024](../issues/024-port-harnez-rograph-primitives-with-provenance.md), [025](../issues/025-integrate-graph-renderers-into-declarative-monitor.md), [012](../issues/012-copy-harnez-rograph-with-verified-provenance-and-bounded-adapters.md) |
| 7 — Simulated live data (Shipped) | [013](../issues/013-deterministic-live-snapshots-and-independent-rolling-histories.md) |
| Reusable measurement & layout (Shipped) | [028](../issues/028-add-reusable-measurement-and-dynamic-box-layout-primitives.md) |
| Typed file prototype (Shipped) | [027](../issues/027-introduce-first-spec-driven-collector-prototype.md) |
| 10 — Splash & startup screen (Shipped) | [029](../issues/029-declarative-centered-layout-and-viewport-alignment-primitives.md), [030](../issues/030-braille-activity-spinner-and-bracketed-progress-bar-primitives.md), [031](../issues/031-provider-status-pill-cluster-and-lifecycle-state-presentation.md), [032](../issues/032-splash-lifecycle-controller-async-provider-coordination-and-key-dismissal.md), [033](../issues/033-harnez-target-splash-screen-integration-and-golden-tests.md) |
| 13 — Hosted widgets contract (Shipped) | [057](../issues/057-hosted-widget-key-contract-child-first-routing-reserved-host-keybinds-and-quit-containment.md) (Shipped), [058](../issues/058-widget-declared-pane-requirements-panerequest.md) (Shipped), [059](../issues/059-fix-mouse-coordinate-convention-mismatch-between-frame-and-tabs-stack-grid.md) (Shipped), [060](../issues/060-periodic-redraw-without-pane-ownership-ticker-interface-and-pane-invalidate.md), [061](../issues/061-themeable-host-provided-theme-propagation-through-composite-widgets.md) |
| 14 — Hosted widgets conversion (Shipped) | [062](../issues/062-example-widget-factories-newwidget-for-split-and-tabs-hosted-loom-demo-mode-headless-bench-smoke.md), [063](../issues/063-convert-filebrowser-example-to-a-hostable-widget.md), [064](../issues/064-convert-splash-and-monitor-examples-to-hostable-widgets.md), [065](../issues/065-ansi-styled-rows-widget-and-treemap-conversion.md) |
| 15 — Compositor & overlays (Shipped) | [066](../issues/066-nested-help-pane-opens-a-second-pane-racing-the-outer-pane-tty-reader.md), [067](../issues/067-animated-loom-background-for-filebrowser-with-proper-compositing.md), [068](../issues/068-formalize-layered-compositor-semantics-and-background-inheritance.md), [069](../issues/069-render-help-as-a-root-level-modal-overlay.md), [070](../issues/070-prevent-text-overflow-in-framework-help-modals.md), [050](../issues/050-support-truecolor-rgb-values-in-theme-specs.md) |
| 16 — Terminal resize & stability (Shipped) | [071](../issues/071-reflow-stacked-dynamic-frames-within-narrow-terminal-heights.md), [072](../issues/072-ensure-stable-responsive-frame-layout-during-terminal-resize.md), [073](../issues/073-add-configurable-winch-resize-diagnostics-app-and-spec-backed-rendering-modes.md), [074](../issues/074-use-measured-winch-speed-for-adaptive-resize-width-guard.md) |
| 17 — Framework primitives (Shipped) | [075](../issues/075-position-cursor-on-previous-folder-when-navigating-up-in-file-browser.md), [076](../issues/076-add-first-class-split-widget-with-ratio-control-dividers-and-nested-focus-traversal.md), [077](../issues/077-support-dynamic-tab-lifecycle-operations-and-configurable-keybindings-in-tabs-widget.md), [078](../issues/078-add-concurrency-safe-metricstore-and-metric-bound-gauge-and-sparkline-widgets.md), [079](../issues/079-extract-reusable-directory-model-filebrowser-primitives-and-platform-file-opener.md), [080](../issues/080-add-declarative-startup-transition-runner-with-deterministic-completion-lifecycle.md) |
| 18 — Examples modernization & smoke (Shipped) | [082](../issues/082-adopt-new-loom-framework-features-across-examples-and-retire-obsolete-code.md), [083](../issues/083-configure-splash-and-treemap-demoargs-with-watch-mode-in-registry-to-prevent-instant-exit-in-loom-demo.md), [085](../issues/085-add-robust-smoke-pty-tests-for-all-example-apps.md), [086](../issues/086-fix-post-refactor-demo-bugs-in-background-split-and-treemap.md), [087](../issues/087-loom-demo-cannot-start-from-codex-terminal.md) |
| 19 — Interactive input & direct manipulation (Shipped) | [088](../issues/088-extend-loom-key-capture-coverage-and-sane-action-defaults.md), [089](../issues/089-add-mouse-cursor-position-hints-and-configurable-visual-effects.md), [091](../issues/091-add-double-click-interaction-for-filebrowser-and-path-trees.md), [092](../issues/092-add-mouse-drag-support-for-scrollbars.md), [093](../issues/093-restrict-filebrowser-mouse-interaction-to-item-content.md), [094](../issues/094-keep-filebrowser-help-modal-in-control-of-keyboard-input.md), [106](../issues/106-make-scrollbar-enabled-by-default-for-panes.md), [134](../issues/134-loom-output-drops-zwj-joiners-and-pads-flag-rows-2-cells-short.md) |
| 20 — Developer feedback & verification tools (096, 107 Shipped; 095 open, Later) | [095](../issues/095-add-human-observable-pty-test-view-mode-and-feedback-flow.md), [096](../issues/096-follow-up-038-with-a-non-ascii-text-rendering-example-app.md) |
| 21 — Mouse correctness (Shipped) | [107](../issues/107-pty-mouse-test-system-detect-loomoji-hover-vs-highlight-offset-via-rendered-fg-bg-cells.md), [108](../issues/108-fix-loomoji-hover-highlight-one-row-above-pointer-dy-1.md) |
| 22 — loom CLI and asset validation (Shipped) | [128](../issues/128-add-box-frame-geometry-and-line-width-validator-for-ansi-assets.md), [129](../issues/129-build-loom-cli-tool-for-tui-asset-validation-measurement-and-interactive-viewing.md), [130](../issues/130-add-loom-format-command-to-auto-align-re-pad-and-normalize-ansi-files.md), [131](../issues/131-add-loom-frame-command-to-wrap-ansi-text-in-styled-box-borders.md) |
| 23 — Real-host proof and media (Shipped) | [042](../issues/042-docs-tuiinput-md-referenced-by-5-code-comments-but-does-not-exist.md), [103](../issues/103-provide-a-hostable-migration-loop-for-coexisting-loom-views.md), [111](../issues/111-add-an-image-media-widget-rendered-with-cati.md) |
| 24 — 2D panning and lazy media (Shipped) | [138](../issues/138-allow-panning-in-loom-panes-and-make-ansiviewer-preview-pane-pannable.md), [140](../issues/140-lazy-load-media-widget-images-and-video-previews.md) |
| 25 — Media example, ansiviewer desktop, release (Shipped) | 141 to 146, 148 to 153, 159 |
| 26 — ANSI core, CLI plain output, Settings number (Shipped) | 162 to 166, 171; 167 M1 |
| 27 — Gap roadmap 180 (Shipped) | 154 to 156, 158, 160, 167 to 170, 172 to 175, 181 to 194; plan: [180](../issues/180-roadmap-close-feature-gaps-found-in-the-framework-comparisons.md), status: [progress/roadmap.md](progress/roadmap.md) |
| 28 — Widget gallery and one event contract (Shipped) | 176, 195 to 207, 209 to 223, 228 to 233, 237, 239, 248 to 251 |
| 29 — RichTextEdit (Shipped) | 254 to 258, 260, 261, 263, 264, 268 to 273, 278 |
| 30 — Media, PaintCanvas, module path, release (Shipped) | 202, 208, 216, 224, 226, 227, 230, 234, 235, 267 |
| 31 — Cleanup sprint (Now) | 247, 136, 135, 242, 240, 052, 238, 100, 225, 252, 277, 275; triage closed 236, 243 to 246, 253, 098, 177 to 179, 048 |
| Closed — Data sources & targets (parked; 014 still open, see Later) | [014](../issues/014-configurable-graph-colors-and-glyph-presentation.md), [015](../issues/015-simulated-voxi-transcript-and-daemon-panels.md), [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md), [017](../issues/017-external-file-and-socket-adapters-with-separate-producer-fixtures.md), [018](../issues/018-explore-bounded-linux-and-daemon-source-adapters.md), [019](../issues/019-evaluate-declarative-source-and-action-wiring.md), [056](../issues/056-embed-real-applications-as-pty-hosted-widgets-tmux-screen-style.md) |

## Long-term outcome

The completed target is a declarative Loom application whose document defines
the title/status chrome, responsive boxes, rows, bars, timelines, colors,
visibility controls, and data bindings. Go defines collectors, external
integration, state transitions, and action behavior. The renderer remains
independent of collection frequency and application-specific coordinates.
