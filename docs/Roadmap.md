---
title: Loom Roadmap
weight: 20
---

# Loom Roadmap

Loom is a **library**, not an application. Its value is measured by what a
*host program* can do with it: how little ceremony it takes to construct,
theme, compose and run a widget; how correct the compositor is under nesting;
and how honest the public API surface is. The example apps in `examples/` are
not the product — they are the proof that the library contract holds.

Loom's value is reusable, dependency-light inline terminal UI: applications
such as the Harnez/Voxi reference dashboards should share widgets without
taking over the full terminal. An application should declare its UI structure
and presentation, while Go supplies behavior and changing data. The UI should
be easy to maintain, responsive to terminal size, testable without a live
terminal, and free from application-specific coordinate/layout code.

The primary reference targets are documented in
[`HarnezUsageTarget.md`](HarnezUsageTarget.md) (usage dashboard and Voxi monitor)
and [`HarnezSplashTarget.md`](HarnezSplashTarget.md) (startup loading and splash
screen).

Stage numbers below preserve the original plan and remain as historical
acceptance context. The Now/Next/Later sequence is the live priority signal and
supersedes stage order wherever the two disagree.

## Shipped since the last roadmap pass (2026-09-20 Lean Sprint)

The lean sprint on 2026-09-20 landed the core hosted-widget input and capability
contract, fixed longstanding mouse and Unicode bugs, and established systematic
PTY test coverage across all example applications.

### Hosted-widget input & capability contract (057, 058)

| Issue | What landed | Why it matters downstream |
|---|---|---|
| [057](../issues/057-hosted-widget-key-contract-child-first-routing-reserved-host-keybinds-and-quit-containment.md) | `KeyConsumer` child-first routing, `Tabs.ArrowSwitch`, `Tabs.OnChildQuit`, inactive-tab focus clearing | Solved key interception and global quit: hosted widgets handle their own keys first; `loom-demo` contains tab exits |
| [058](../issues/058-widget-declared-pane-requirements-panerequest.md) | `PaneRequest`/`PaneRequester` with composite merge rules (Tabs, Stack, Frame, Grid) | Widgets declare mouse mode, `Resizeable`, `MaxCols`, and quit requirements instead of hardcoding them onto an owned `Pane` |

### Coordinate & Unicode fixes (036, 038, 059)

| Issue | What landed | Why it matters downstream |
|---|---|---|
| [036](../issues/036-keyevent-ergonomic-helpers-e-name-and-e-is-for-unified-key-text-matching.md) | `KeyEvent.Name()`, `KeyEvent.Is()`, `KeyEvent.Rune()` ergonomic helpers | Unified key and text matching; eliminated repetitive `key := e.Key; if key == "" { key = e.Text }` boilerplate across widgets |
| [038](../issues/038-fix-multi-byte-utf-8-string-truncation-in-popup-title-and-borders.md) | Backfilled UTF-8 acceptance tests; fixed `Popup.Draw` border-over-title rendering order | Multi-byte Unicode/emoji titles no longer corrupt popup frames or overwrite title text |
| [059](../issues/059-fix-mouse-coordinate-convention-mismatch-between-frame-and-tabs-stack-grid.md) | Standardized canvas-absolute 0-based mouse coordinates across `Pane`, `Frame`, `Stack`, `Grid`, `Tabs`, `Choice`, `View`, `widget.go` | Eliminated double-subtraction coordinate offsets; restored accurate mouse hit-testing in nested composite containers |

### Test infrastructure (085)

| Issue | What landed | Why it matters downstream |
|---|---|---|
| [085](../issues/085-add-robust-smoke-pty-tests-for-all-example-apps.md) | End-to-end PTY smoke tests across all 9 example apps | Prevents regressions in TTY lifecycle, initial screen rendering, key response, and clean exit |

---

## Themes in the open backlog

The 22 open tickets fall into seven distinct themes:

1. **Hosted-widget example conversion & execution** (060, 061, 062, 063, 064, 065) — converting all example applications into pure `loom.Widget` implementations hosted by `loom-demo` and custom hosts.
2. **Interactive input & modal robustness** (088, 091, 093, 094) — comprehensive key capture coverage, double-click gestures, precise hit-testing bounds, and modal keyboard trapping.
3. **Direct manipulation & spatial UI effects** (089, 090, 092) — scrollbar thumb dragging, mouse proximity/spatial lighting hints, and image-backed application backgrounds.
4. **Developer ergonomics & human-in-the-loop testing** (095, 096, 042) — observable PTY test workflows with Harnez ticket draft integration, visual non-ASCII regression suites, and input documentation.
5. **Rendering, geometry, and layout capabilities** (051, 034, 081, 037, 052) — dynamic frame box height filling, ANSI escape parsing, squarified treemap layout models, and border consolidation.
6. **Diagnostic & edge correctness** (048, 039) — emoji font width drift guidance and eighth-block subchar seam recommendations.
7. **Parked declarative data sources & simulated targets** (014–019, 056) — external system probes and PTY terminal embedding.

---

## Now — immediate conversion payoff & input/modal correctness

**Rationale.** With 057, 058, 059, and 085 shipped, the core hosted-widget foundation is live. The highest-value immediate step is proving the contract on the easiest examples (**062**), fixing active UX/modal traps in the filebrowser (**094**, **093**), expanding key coverage (**088**), and enabling scrollbar mouse drag (**092**).

- [062](../issues/062-example-widget-factories-newwidget-for-split-and-tabs-hosted-loom-demo-mode-headless-bench-smoke.md)
  (P2, Feature): `NewWidget` factories for `split` and `tabs`, hosted `loom-demo` mode, and headless benchmark smoke tests. The primary end-to-end validation of 057 (`KeyConsumer`) and 058 (`PaneRequest`).
- [094](../issues/094-keep-filebrowser-help-modal-in-control-of-keyboard-input.md)
  (P1, Bug): `:help<CR>` modal must maintain exclusive keyboard control and trap quit/action keys, preventing accidental parent application termination while help is open.
- [093](../issues/093-restrict-filebrowser-mouse-interaction-to-item-content.md)
  (P1, Bug): Restrict filebrowser mouse hit-testing to visible text and decorations, ensuring clicks on row whitespace do not accidentally trigger selection or file opening.
- [088](../issues/088-extend-loom-key-capture-coverage-and-sane-action-defaults.md)
  (P1, Enhancement): Broaden key-capture decoding for letters, numbers, punctuation, German umlauts, and modifier combinations (Shift, Ctrl, Alt), establishing consistent action defaults.
- [092](../issues/092-add-mouse-drag-support-for-scrollbars.md)
  (P1, Enhancement): Direct manipulation for vertical and horizontal scrollbars via mouse drag, converting pointer delta into proportional scroll offsets.
- [060](../issues/060-periodic-redraw-without-pane-ownership-ticker-interface-and-pane-invalidate.md)
  (P2, Feature): `Ticker` interface and `Pane.Invalidate` mechanism allowing live periodic updates without requiring the widget to own the `Pane` event loop. Hard prerequisite for 064 (`monitor`/`splash`).

---

## Next — advanced UI gestures, themes, and full widget conversions

**Rationale.** Completes the remaining hosted-widget requirements (theming, ticker propagation) and delivers richer interactive experiences (double click, spatial cursor effects, image backgrounds, non-ASCII test surfaces).

- [091](../issues/091-add-double-click-interaction-for-filebrowser-and-path-trees.md)
  (P1, Enhancement): Reusable double-click recognition for file opening, directory traversal, and tree node activation.
- [061](../issues/061-themeable-host-provided-theme-propagation-through-composite-widgets.md)
  (P3, Feature): `Themeable` interface and dynamic theme propagation across composite container hierarchies. Gates 063.
- [063](../issues/063-convert-filebrowser-example-to-a-hostable-widget.md)
  (P2, Feature): Convert `filebrowser` into a bare `loom.Widget`, validating `KeyConsumer`, `PaneRequest`, 059 mouse hit-testing, and theme propagation.
- [089](../issues/089-add-mouse-cursor-position-hints-and-configurable-visual-effects.md)
  (P1, Enhancement): Expose relative `dx,dy` mouse proximity hints to paint cells; support background brightening, trailing star, and expanding pulse effects.
- [090](../issues/090-support-image-backed-app-backgrounds-and-background-theme-switching.md)
  (P1, Enhancement): Image-backed application backgrounds using `ubunatic.com/cati`, dimmed cell sampling, and theme toggling in `examples/background`.
- [096](../issues/096-follow-up-038-with-a-non-ascii-text-rendering-example-app.md)
  (P1, Enhancement): Dedicated `examples/textrender` visual regression application testing non-ASCII, umlauts, CJK, combining marks, and emoji layout integrity.
- [095](../issues/095-add-human-observable-pty-test-view-mode-and-feedback-flow.md)
  (P1, Enhancement): Human-observable PTY test runner with pacing controls, step feedback prompts, and automated `harnez issue new` draft creation.
- [051](../issues/051-allow-frame-boxes-to-fill-available-content-height.md)
  (P2, Feature): Let `Frame` boxes fill available vertical content height automatically without hardcoded row arithmetic.
- [034](../issues/034-ansi-sgr-escape-sequence-parsing-and-writeansi-canvas-helper.md)
  (P2, Feature): `ParseANSI` and `Canvas.WriteANSI` for rich SGR color/attribute rendering. Hard prerequisite for 065.
- [081](../issues/081-expose-treemapcell-layout-models-squarified-partitioning-and-value-based-color-scales.md)
  (P2, Feature): Expose `TreemapCell` geometry, squarified treemap partitioning, and continuous `ColorScale` palettes. Prerequisite for 065.
- [042](../issues/042-docs-tuiinput-md-referenced-by-5-code-comments-but-does-not-exist.md)
  (P3, Documentation): Author `docs/TuiInput.md` formalizing terminal input parsing, key decoding, and event dispatch architecture.

---

## Later — complex conversions and surface hygiene

**Rationale.** High-complexity widget refactorings and non-urgent cleanup that depend on multiple Next primitives.

- [064](../issues/064-convert-splash-and-monitor-examples-to-hostable-widgets.md)
  (P2, Feature): Convert live-data monitoring and splash examples into hostable widgets with options structs (gated by 060).
- [065](../issues/065-ansi-styled-rows-widget-and-treemap-conversion.md)
  (P3, Feature): Convert treemap into a hostable widget, removing direct `/dev/tty` and raw-mode loops (gated by 034 and 081).
- [037](../issues/037-canvas-drawborder-and-drawbox-primitives-with-configurable-boxstyles.md)
  (P2, Refactor): Consolidated `Canvas.DrawBorder`/`DrawBox` primitives with reusable `BoxStyle`.
- [052](../issues/052-resolve-unused-choicestyle-border-contract.md)
  (P2, Refactor): Reconcile unused `ChoiceStyle.Border` public field.

---

## Close / deprioritize candidates

- [048](../issues/048-emoji-rune-width-discrepancy-causes-horizontal-border-drift.md)
  (P2, Bug) — **Reclassify to Documentation.** Font outline and terminal `wcwidth` variance for neutral-width emoji (e.g. `🖼`) cannot be solved in library code alone. Document standard display-width guarantees and close.
- [039](../issues/039-graph-renderbar-subchar-boundary-glyph-shows-a-visible-seam-without-ansi-background-styling.md)
  (P3, Bug) — **Close as won't-fix-in-code.** Root cause is terminal font bearing and eighth-block outline rendering in VTE terminals. Document recommendation and close.
- [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md)
  (P2) — **Park / Close as milestone.** Convert to reference checklist for future integration consumers.
- [056](../issues/056-embed-real-applications-as-pty-hosted-widgets-tmux-screen-style.md)
  (P3, Aspirational) — Direction 1 (in-process widget hosting) is actively delivered by 057–065. Direction 2 (full VT100 sub-process terminal emulation) is a distinct product effort.

---

## Parked — declarative data sources and simulated targets (014–019)

- [014](../issues/014-configurable-graph-colors-and-glyph-presentation.md) →
  [015](../issues/015-simulated-voxi-transcript-and-daemon-panels.md) →
  [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md)
- [017](../issues/017-external-file-and-socket-adapters-with-separate-producer-fixtures.md) →
  [018](../issues/018-explore-bounded-linux-and-daemon-source-adapters.md)
- [019](../issues/019-evaluate-declarative-source-and-action-wiring.md)

---

## Dependency map

```text
[036, 057, 058, 059, 085 Shipped]
       │
       ├──► 062 ──► 063 ──► 064 ──► 065
       │     ▲       ▲       ▲        ▲
061 ───┴─────┘       │       │        │
060 ─────────────────┴───────┘        │
034 ──────────────────────────────────┤
081 ──────────────────────────────────┘
051 ──► (validated by 063)
088, 089, 090, 091, 092, 093, 094, 095, 096 — UX, interaction & test enhancements
037 ──► (informs 052)
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
| 13 — Hosted widgets contract (Shipped / Active) | [057](../issues/057-hosted-widget-key-contract-child-first-routing-reserved-host-keybinds-and-quit-containment.md) (Shipped), [058](../issues/058-widget-declared-pane-requirements-panerequest.md) (Shipped), [059](../issues/059-fix-mouse-coordinate-convention-mismatch-between-frame-and-tabs-stack-grid.md) (Shipped), [060](../issues/060-periodic-redraw-without-pane-ownership-ticker-interface-and-pane-invalidate.md), [061](../issues/061-themeable-host-provided-theme-propagation-through-composite-widgets.md) |
| 14 — Hosted widgets conversion | [062](../issues/062-example-widget-factories-newwidget-for-split-and-tabs-hosted-loom-demo-mode-headless-bench-smoke.md), [063](../issues/063-convert-filebrowser-example-to-a-hostable-widget.md), [064](../issues/064-convert-splash-and-monitor-examples-to-hostable-widgets.md), [065](../issues/065-ansi-styled-rows-widget-and-treemap-conversion.md) |
| 15 — Compositor & overlays (Shipped) | [066](../issues/066-nested-help-pane-opens-a-second-pane-racing-the-outer-pane-tty-reader.md), [067](../issues/067-animated-loom-background-for-filebrowser-with-proper-compositing.md), [068](../issues/068-formalize-layered-compositor-semantics-and-background-inheritance.md), [069](../issues/069-render-help-as-a-root-level-modal-overlay.md), [070](../issues/070-prevent-text-overflow-in-framework-help-modals.md), [050](../issues/050-support-truecolor-rgb-values-in-theme-specs.md) |
| 16 — Terminal resize & stability (Shipped) | [071](../issues/071-reflow-stacked-dynamic-frames-within-narrow-terminal-heights.md), [072](../issues/072-ensure-stable-responsive-frame-layout-during-terminal-resize.md), [073](../issues/073-add-configurable-winch-resize-diagnostics-app-and-spec-backed-rendering-modes.md), [074](../issues/074-use-measured-winch-speed-for-adaptive-resize-width-guard.md) |
| 17 — Framework primitives (Shipped) | [075](../issues/075-position-cursor-on-previous-folder-when-navigating-up-in-file-browser.md), [076](../issues/076-add-first-class-split-widget-with-ratio-control-dividers-and-nested-focus-traversal.md), [077](../issues/077-support-dynamic-tab-lifecycle-operations-and-configurable-keybindings-in-tabs-widget.md), [078](../issues/078-add-concurrency-safe-metricstore-and-metric-bound-gauge-and-sparkline-widgets.md), [079](../issues/079-extract-reusable-directory-model-filebrowser-primitives-and-platform-file-opener.md), [080](../issues/080-add-declarative-startup-transition-runner-with-deterministic-completion-lifecycle.md) |
| 18 — Examples modernization & smoke (Shipped) | [082](../issues/082-adopt-new-loom-framework-features-across-examples-and-retire-obsolete-code.md), [083](../issues/083-configure-splash-and-treemap-demoargs-with-watch-mode-in-registry-to-prevent-instant-exit-in-loom-demo.md), [085](../issues/085-add-robust-smoke-pty-tests-for-all-example-apps.md), [086](../issues/086-fix-post-refactor-demo-bugs-in-background-split-and-treemap.md), [087](../issues/087-loom-demo-cannot-start-from-codex-terminal.md) |
| 19 — Interactive input & direct manipulation (Active) | [088](../issues/088-extend-loom-key-capture-coverage-and-sane-action-defaults.md), [089](../issues/089-add-mouse-cursor-position-hints-and-configurable-visual-effects.md), [090](../issues/090-support-image-backed-app-backgrounds-and-background-theme-switching.md), [091](../issues/091-add-double-click-interaction-for-filebrowser-and-path-trees.md), [092](../issues/092-add-mouse-drag-support-for-scrollbars.md), [093](../issues/093-restrict-filebrowser-mouse-interaction-to-item-content.md), [094](../issues/094-keep-filebrowser-help-modal-in-control-of-keyboard-input.md) |
| 20 — Developer feedback & verification tools (Active) | [095](../issues/095-add-human-observable-pty-test-view-mode-and-feedback-flow.md), [096](../issues/096-follow-up-038-with-a-non-ascii-text-rendering-example-app.md) |
| Parked — Data sources & targets | [014](../issues/014-configurable-graph-colors-and-glyph-presentation.md), [015](../issues/015-simulated-voxi-transcript-and-daemon-panels.md), [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md), [017](../issues/017-external-file-and-socket-adapters-with-separate-producer-fixtures.md), [018](../issues/018-explore-bounded-linux-and-daemon-source-adapters.md), [019](../issues/019-evaluate-declarative-source-and-action-wiring.md), [056](../issues/056-embed-real-applications-as-pty-hosted-widgets-tmux-screen-style.md) |

## Long-term outcome

The completed target is a declarative Loom application whose document defines
the title/status chrome, responsive boxes, rows, bars, timelines, colors,
visibility controls, and data bindings. Go defines collectors, external
integration, state transitions, and action behavior. The renderer remains
independent of collection frequency and application-specific coordinates.
