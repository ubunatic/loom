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

## Shipped since the last roadmap pass (2026-09-24 reconcile)

Since the 2026-09-20 pass, lean sprints closed most of the former Now and Next
buckets, and a hover-testing session closed a real mouse-correctness bug in the
loomoji example.

| [123](../issues/123-bug-textedit-immediate-exit-on-clicks-keystrokes-due-to-unhandled-root-return-contract.md) | Fixed event return loop contract (`false` = continue, `true` = quit), global `F10`/`Ctrl-Q` quit | Solves critical interactive crash across TUI applications |
| [117](../issues/117-textedit-tui-visual-hierarchy-refinement-and-ansi-mockup.md) | Line gutters, active line indicator, status bar hints matching `.ansi` mockup | Clean visual hierarchy and production TUI standard |
| [118](../issues/118-spike-tree-sitter-wasm-runtime-in-internal-canary-treesitter.md) | Pure-Go Wasm Tree-Sitter runtime evaluation in `internal/canary/treesitter` | Establishes zero-CGO Wasm parsing benchmark |
| [119](../issues/119-ui-neutral-syntax-package-ast-point-span-and-tree-sitter-binding.md) | `syntax/` package (`Point`, `Edit`, `Span`, `Engine`, `ThemeMap`, `StyleResolver`, `LexicalEngine`) | UI-neutral coordinates, single-pass highlighting architecture |
| [121](../issues/121-integrate-syntax-engine-into-loom-textarea-and-textedit.md) | Viewport-bounded single-pass syntax rendering in `loom.TextArea` | Zero-overhead syntax highlighting during rendering |
| [122](../issues/122-ast-driven-code-navigation-symbol-outline-breadcrumbs-folding-in-textedit.md) | Scope breadcrumbs (`ScopeAt`), symbol outline sidebar (`Navigator`), code folding (`FoldRanges`) | Full structural code editing and navigation |

Earlier pass (2026-09-24): 034, 037, 051, 060, 061, 062, 081, 088, 092, 093, 094, 096, 097, 107, 108.
Earlier pass (2026-09-20): 036, 038, 057, 058, 059, 085. See the stage table.

---

## Themes in the open backlog

The 26 open tickets fall into seven themes:

1. **Hosted-widget conversion** (063, 064, 065, 103, 105): every gate is now
   shipped. What remains is converting the examples and proving migration in a
   real host.
2. **Mouse & direct manipulation** (089, 091, 106): hover correctness is now
   proven by 107/108, which unblocks pointer-driven effects.
3. **ANSI viewer fidelity** (099, 100): recording and replaying real terminal output.
4. **Verification & review tooling** (095, 102): human-observable PTY runs and manual checks.
5. **Docs & hygiene** (042, 098, 101, 052).
6. **Visual extras** (090).
7. **Parked or close candidates** (014 to 019, 039, 048, 056).

---

## Now: finish hosted-widget conversion and fix correctness bugs

**Rationale.** The library's value is shared, hostable widgets for real
dashboards. Every prerequisite for the conversions is shipped, so the
conversions are now the cheapest proof of the contract. 099 is the only open
P1 bug.

- [063](../issues/063-convert-filebrowser-example-to-a-hostable-widget.md) (P2): convert `filebrowser` to a bare `loom.Widget`. Its gates (060, 061, 093, 094) are closed. Pair it with **105** so the conversion targets one shared navigation pane.
- [105](../issues/105-unify-filebrowser-navigation-pane-across-examples.md) (P2): one shared filebrowser navigation pane for `ansiviewer`, `filebrowser` and future consumers.
- [064](../issues/064-convert-splash-and-monitor-examples-to-hostable-widgets.md) (P2): hostable `splash`/`monitor`. Moved up from Later because 060 shipped.
- [099](../issues/099-fix-ansiviewer-recording-of-ansi-output-and-terminal-width.md) (P1, Bug): `ansiviewer --record` loses colour and wraps at the wrong width.
- [101](../issues/101-lean-sprint-leftovers-dead-treemap-block-gofmt-unproven-split-capture-test-thin-evidence.md) (P3, Hygiene): small, headless lean-sprint leftovers. Cheap, and it keeps the suite honest before the next conversion wave.
- [111](../issues/111-add-an-image-media-widget-rendered-with-cati.md) (P2, user-queued 2026-09-24): hostable Image/Media widget rendered with `../cati`; canary the cati v1 API first.
- [112](../issues/112-add-media-controls-play-pause-zoom-and-panning-for-the-image-media-widget.md) (P3, user-queued): play/pause, zoom +/-, panning. Depends on 111.

---

## Next: pointer UX, real-host migration, and verification

**Rationale.** 107/108 give a trustworthy hover and coordinate baseline plus a
black-box probe method, so pointer features can now be built and tested
without guessing offsets. 103 is the strongest proof of the product goal:
migrating a real Harnez view onto Loom.

- [089](../issues/089-add-mouse-cursor-position-hints-and-configurable-visual-effects.md) (P1): mouse cursor hints and effects. Build its PTY tests on the 107 `ptytest` colour-cell method ([`HoverTesting.md`](HoverTesting.md)).
- [091](../issues/091-add-double-click-interaction-for-filebrowser-and-path-trees.md) (P1): reusable double-click recognition.
- [106](../issues/106-make-scrollbar-enabled-by-default-for-panes.md) (P2): `scrollbar: auto` as the pane default.
- [065](../issues/065-ansi-styled-rows-widget-and-treemap-conversion.md) (P3): ANSI-styled rows and a hostable treemap. Moved up from Later because 034 and 081 shipped.
- [103](../issues/103-provide-a-hostable-migration-loop-for-coexisting-loom-views.md) (P2): hostable migration loop in a Harnez-usage clone.
- [095](../issues/095-add-human-observable-pty-test-view-mode-and-feedback-flow.md) (P1): human-observable PTY view mode. 107's colour-cell snapshots are a natural base for step feedback.
- [102](../issues/102-human-review-collection-manual-checks-for-lean-sprint-deliveries.md) (P2): manual review checklist for the lean-sprint deliveries.

---

## Later: docs, polish, extras

- [042](../issues/042-docs-tuiinput-md-referenced-by-5-code-comments-but-does-not-exist.md) (P3, Documentation): `docs/TuiInput.md`. More valuable now that 088 has settled the key decoding.
- [098](../issues/098-document-animatedbackground-ticker-initialization-pitfall.md) (P2, Documentation): `AnimatedBackground` ticker initialization pitfall.
- [090](../issues/090-support-image-backed-app-backgrounds-and-background-theme-switching.md) (P1): image-backed backgrounds. A visual extra, not needed by the reference dashboards.
- [100](../issues/100-investigate-full-width-ansi-top-bar-background-in-ansiviewer.md) (P3, Bug): full-width top-bar background in ansiviewer. Follows 099.
- [052](../issues/052-resolve-unused-choicestyle-border-contract.md) (P2, Refactor): unused `ChoiceStyle.Border`. Now informed by the shipped 037 `BoxStyle`.

---

## Close / deprioritize candidates

- [048](../issues/048-emoji-rune-width-discrepancy-causes-horizontal-border-drift.md) (P2, Bug): **reclassify to Documentation.** Emoji width variance across terminals cannot be fixed in library code.
- [039](../issues/039-graph-renderbar-subchar-boundary-glyph-shows-a-visible-seam-without-ansi-background-styling.md) (P3, Bug): **close as won't-fix-in-code.** The seam comes from terminal font rendering.
- [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md) (P2): **park / close as milestone.**
- [056](../issues/056-embed-real-applications-as-pty-hosted-widgets-tmux-screen-style.md) (P3, Aspirational): in-process hosting is delivered by 057 to 065. Full VT100 sub-process emulation is a separate product. `ptytest`'s VT colour-cell model (107) is a small reusable seed if it is ever revived.

---

## Parked: declarative data sources and simulated targets (014 to 019)

- [014](../issues/014-configurable-graph-colors-and-glyph-presentation.md) → [015](../issues/015-simulated-voxi-transcript-and-daemon-panels.md) → [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md)
- [017](../issues/017-external-file-and-socket-adapters-with-separate-producer-fixtures.md) → [018](../issues/018-explore-bounded-linux-and-daemon-source-adapters.md)
- [019](../issues/019-evaluate-declarative-source-and-action-wiring.md)

---

## Dependency map

```text
[057 058 059 060 061 062 034 081 107 108 shipped]
   │
   ├──► 063 ◄──► 105 ──► 103 (real-host migration)
   ├──► 064
   ├──► 065
   └──► 089, 091 (pointer UX on proven hover baseline)
099 ──► 100
107 ──► 095 (colour-cell snapshots)
088 ──► 042
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
| 14 — Hosted widgets conversion | [062](../issues/062-example-widget-factories-newwidget-for-split-and-tabs-hosted-loom-demo-mode-headless-bench-smoke.md), [063](../issues/063-convert-filebrowser-example-to-a-hostable-widget.md), [064](../issues/064-convert-splash-and-monitor-examples-to-hostable-widgets.md), [065](../issues/065-ansi-styled-rows-widget-and-treemap-conversion.md) |
| 15 — Compositor & overlays (Shipped) | [066](../issues/066-nested-help-pane-opens-a-second-pane-racing-the-outer-pane-tty-reader.md), [067](../issues/067-animated-loom-background-for-filebrowser-with-proper-compositing.md), [068](../issues/068-formalize-layered-compositor-semantics-and-background-inheritance.md), [069](../issues/069-render-help-as-a-root-level-modal-overlay.md), [070](../issues/070-prevent-text-overflow-in-framework-help-modals.md), [050](../issues/050-support-truecolor-rgb-values-in-theme-specs.md) |
| 16 — Terminal resize & stability (Shipped) | [071](../issues/071-reflow-stacked-dynamic-frames-within-narrow-terminal-heights.md), [072](../issues/072-ensure-stable-responsive-frame-layout-during-terminal-resize.md), [073](../issues/073-add-configurable-winch-resize-diagnostics-app-and-spec-backed-rendering-modes.md), [074](../issues/074-use-measured-winch-speed-for-adaptive-resize-width-guard.md) |
| 17 — Framework primitives (Shipped) | [075](../issues/075-position-cursor-on-previous-folder-when-navigating-up-in-file-browser.md), [076](../issues/076-add-first-class-split-widget-with-ratio-control-dividers-and-nested-focus-traversal.md), [077](../issues/077-support-dynamic-tab-lifecycle-operations-and-configurable-keybindings-in-tabs-widget.md), [078](../issues/078-add-concurrency-safe-metricstore-and-metric-bound-gauge-and-sparkline-widgets.md), [079](../issues/079-extract-reusable-directory-model-filebrowser-primitives-and-platform-file-opener.md), [080](../issues/080-add-declarative-startup-transition-runner-with-deterministic-completion-lifecycle.md) |
| 18 — Examples modernization & smoke (Shipped) | [082](../issues/082-adopt-new-loom-framework-features-across-examples-and-retire-obsolete-code.md), [083](../issues/083-configure-splash-and-treemap-demoargs-with-watch-mode-in-registry-to-prevent-instant-exit-in-loom-demo.md), [085](../issues/085-add-robust-smoke-pty-tests-for-all-example-apps.md), [086](../issues/086-fix-post-refactor-demo-bugs-in-background-split-and-treemap.md), [087](../issues/087-loom-demo-cannot-start-from-codex-terminal.md) |
| 19 — Interactive input & direct manipulation (088, 092–094 Shipped; rest Active) | [088](../issues/088-extend-loom-key-capture-coverage-and-sane-action-defaults.md), [089](../issues/089-add-mouse-cursor-position-hints-and-configurable-visual-effects.md), [090](../issues/090-support-image-backed-app-backgrounds-and-background-theme-switching.md), [091](../issues/091-add-double-click-interaction-for-filebrowser-and-path-trees.md), [092](../issues/092-add-mouse-drag-support-for-scrollbars.md), [093](../issues/093-restrict-filebrowser-mouse-interaction-to-item-content.md), [094](../issues/094-keep-filebrowser-help-modal-in-control-of-keyboard-input.md) |
| 20 — Developer feedback & verification tools (096, 107 Shipped; 095 Active) | [095](../issues/095-add-human-observable-pty-test-view-mode-and-feedback-flow.md), [096](../issues/096-follow-up-038-with-a-non-ascii-text-rendering-example-app.md) |
| 21 — Mouse correctness (Shipped) | [107](../issues/107-pty-mouse-test-system-detect-loomoji-hover-vs-highlight-offset-via-rendered-fg-bg-cells.md), [108](../issues/108-fix-loomoji-hover-highlight-one-row-above-pointer-dy-1.md) |
| Parked — Data sources & targets | [014](../issues/014-configurable-graph-colors-and-glyph-presentation.md), [015](../issues/015-simulated-voxi-transcript-and-daemon-panels.md), [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md), [017](../issues/017-external-file-and-socket-adapters-with-separate-producer-fixtures.md), [018](../issues/018-explore-bounded-linux-and-daemon-source-adapters.md), [019](../issues/019-evaluate-declarative-source-and-action-wiring.md), [056](../issues/056-embed-real-applications-as-pty-hosted-widgets-tmux-screen-style.md) |

## Long-term outcome

The completed target is a declarative Loom application whose document defines
the title/status chrome, responsive boxes, rows, bars, timelines, colors,
visibility controls, and data bindings. Go defines collectors, external
integration, state transitions, and action behavior. The renderer remains
independent of collection frequency and application-specific coordinates.
