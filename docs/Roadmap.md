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

## Shipped since the last roadmap pass (2026-09-28)

- Media example and widget: [141](../issues/141-move-media-widget-labels-colors-and-timing-defaults-into-specs.md) spec-driven media defaults, [143](../issues/143-make-the-media-example-full-width-and-play-video.md)/[144](../issues/144-fix-examples-media-pane-initialization-to-take-full-terminal-size.md)/[146](../issues/146-set-pane-maxcols-0-in-examples-media-to-uncap-terminal-width.md) full-width video example, [145](../issues/145-show-measured-screen-dimensions-and-debug-data-in-a-status-bar-below-the-media-demo.md) status bar, [149](../issues/149-convert-the-media-example-to-a-cobra-cli-app.md) Cobra CLI.
- ansiviewer as a desktop app: [148](../issues/148-open-ansiviewer-directly-on-a-file-argument.md) open file arguments, [150](../issues/150-register-ansiviewer-as-a-linux-desktop-viewing-tool.md) desktop entry, [151](../issues/151-add-and-install-desktop-icon-for-ansiviewer.md) icon, [152](../issues/152-run-update-desktop-database-and-gtk-update-icon-cache-in-make-install.md) desktop cache hooks.
- [142](../issues/142-document-theme-selection-and-listing-in-the-go-library.md) theme API docs, [153](../issues/153-hide-the-filebrowser-search-ui-until-search-is-active.md) hidden idle filebrowser search, [159](../issues/159-set-up-go-release-pipeline-for-loom-and-cli-tools.md) release pipeline (GoReleaser, minisign).
- [160](../issues/160-media-widget-playback-controls-and-poster-frame-support.md) M1 landed (`fbe08d1`: poster frame, `Play`/`Pause`/`Restart`); ticket stays open pending verification.

**Closed by decision** (the previous Close/Park list was executed): 016, 039, 056, 015, 017, 018, 019.

Earlier passes: 138, 140; 089, 091, 106, 134; 065, 042, 111, 103; 064, 124,
128, 130, 131; 063, 099, 101, 105, 129; 117 to 123; 034, 037, 051, 060 to 062,
081, 088, 092 to 097, 107, 108; 036, 038, 057 to 059, 085. See the stage table.

---

## Themes in the open backlog

The 24 open tickets fall into six themes:

1. **Stability and rendering correctness** (136, 135, 048, 100): a live resize bug, an environment-leaking test, two ANSI-width/background quirks.
2. **Diagnosability and visual testing** (133, 095, 157): ruler overlay, human-observable PTY tests, and a golden-ANSI canvas comparator.
3. **App-building primitives from loom-games** (154, 155, 156, 158): ticker cadence control, dialog overlay, SubCanvas/Blit, KeyMap. New this pass; all filed from the snake v0 build, i.e. a second real host beyond Harnez.
4. **Media** (160, 161, 112, 090): playback API verification, the redesign-005 example, zoom/pan, image backgrounds.
5. **CLI asset tools** (137, 147, 139): feedback triage, unclosed-box detection, `loom play`.
6. **Internals, adoption, docs** (116, 102, 052, 098, 014, 120).

---

## Now: fix live bugs, finish in-flight media, confirm the real host

**Rationale.** 136 is a user-facing regression and 135 a misreporting test;
133 makes this bug class visible. 160 is half-done (M1 committed) and 161 is a
small, already-designed follow-up that consumes 160's API, so finishing both
avoids stale work in flight. 102 gates harnez 607.

- [136](../issues/136-pane-misses-a-resize-that-happens-during-startup-foot-app-draws-at-initial-size.md) (P2, Bug): pane misses a resize during startup (SIGWINCH before `signal.Notify`).
- [135](../issues/135-testbrowserusesanimatedbackground-fails-when-loom-evidence-1.md) (P3, Bug): `LOOM_EVIDENCE=1` breaks a filebrowser test via ambient env state.
- [133](../issues/133-pane-debug-mode-with-ruler-overlay-shift-f12-loom-debug.md) (P2): ruler overlay (Shift-F12, `LOOM_DEBUG`); exposes 136, 048, 100.
- [160](../issues/160-media-widget-playback-controls-and-poster-frame-support.md) (P2): **new, in flight.** Verify M1 (poster frame, playback API) and close.
- [161](../issues/161-align-media-example-with-redesign-005.md) (P2): **new.** Media example to redesign 005; depends on 160.
- [102](../issues/102-human-review-collection-manual-checks-for-lean-sprint-deliveries.md) (P2): manual checks for 103, 111, 124, 064; gates harnez 607.

---

## Next: primitives for a second host, visual testing, ANSI core

**Rationale.** 154, 155, 156 and 158 each came from building a real app
(loom-games snake) and are small, additive API. 156 (SubCanvas/Blit) goes
first because 155 (dialog) can build on it. 157 pairs with 095: both make
rendering regressions catchable, and 157 directly serves the golden-mockup
workflow used for 161. 116 is library-wide and safer after the Now fixes.
147 is the concrete output of 137's feedback; do them together.

- [156](../issues/156-canvas-subcanvas-and-blit-layer-compositing.md) (P2): **new.** `Canvas.SubCanvas` / `Blit`; unblocks 155.
- [155](../issues/155-built-in-modal-and-dialog-overlay-primitive.md) (P2): **new.** Dialog/overlay primitive; builds on 156.
- [154](../issues/154-dynamic-cadence-and-timer-control-on-ticker.md) (P2): **new.** Dynamic ticker cadence and reset.
- [158](../issues/158-keymap-and-action-key-aliasing-helper.md) (P2): **new.** `KeyMap` action aliasing; fits spec-declared keybindings.
- [157](../issues/157-ansi-golden-mockup-visual-test-comparator.md) (P2): **new.** Golden ANSI canvas comparator (`loomtest`).
- [095](../issues/095-add-human-observable-pty-test-view-mode-and-feedback-flow.md) (P1): human-observable PTY test view mode.
- [116](../issues/116-extract-zero-alloc-ansi-styling-and-parsing-into-dedicated-ansi-subpackage.md) (P2): zero-alloc `ansi` subpackage.
- [137](../issues/137-feedback-cli-asset-tools-workflow-and-multi-box-ansi-validation.md) (P3) + [147](../issues/147-detect-unclosed-boxes-and-prioritize-box-validation-errors.md) (P2, **new**): CLI asset-tool feedback triage and unclosed-box detection.

---

## Later: docs, polish, extras

- [112](../issues/112-add-media-controls-play-pause-zoom-and-panning-for-the-image-media-widget.md) (P3): media zoom/panning; **moved down from Next** because 160 now covers play/pause and 138 shipped pane panning, leaving mainly zoom. Video parts wait for cati 059.
- [048](../issues/048-emoji-rune-width-discrepancy-causes-horizontal-border-drift.md) (P2, Bug): **reclassify to Documentation**; emoji width variance is a terminal property (see [`EmojiWidth.md`](EmojiWidth.md)).
- [100](../issues/100-investigate-full-width-ansi-top-bar-background-in-ansiviewer.md) (P3): full-width top-bar background in ansiviewer; may fold into 116.
- [098](../issues/098-document-animatedbackground-ticker-initialization-pitfall.md) (P2): `AnimatedBackground` ticker pitfall doc; consider bundling with 154.
- [052](../issues/052-resolve-unused-choicestyle-border-contract.md) (P2): unused `ChoiceStyle.Border` contract.
- [014](../issues/014-configurable-graph-colors-and-glyph-presentation.md) (P2): configurable graph colors/glyphs; **kept open** while its data-source siblings closed, since it is pure presentation.
- [090](../issues/090-support-image-backed-app-backgrounds-and-background-theme-switching.md) (P1): image-backed backgrounds; reuse the 111 cati rendering. No host has asked.
- [120](../issues/120-implement-wazero-backed-tree-sitter-syntax-engine-with-embedded-grammars-and-queries.md) (P2): wazero Tree-Sitter engine; large, dependency-heavy, no dashboard use case.
- [139](../issues/139-add-loom-play-for-interactive-tui-commands-and-ansi-capture.md) (P2): `loom play`; needs a reuse-vs-build decision (ansiviewer `--record`, Reelang, tmux, asciinema).

---

## Close / Park

- Nothing new to close. The prior list (016, 039, 056, 015, 017 to 019) was closed.
- Candidate: if 157 lands as a reusable package, re-check whether 095 still needs its own view mode or can shrink to a feedback flow on top of 157.

---

## Dependency map

```text
[103 shipped] ──► 102 (manual check) ──► harnez 607
[160 M1 committed] ──► 160 verify ──► 161 (redesign 005)
[160, 138 shipped panning] ──► 112 (zoom; video waits on cati 059)
156 (SubCanvas/Blit) ──► 155 (dialog)
[128-131 shipped] ──► 116 (ansi subpackage) ──► 100?
[133 debug overlay] ──► exposes 136, 048, 100
137 (feedback) ──► 147 (unclosed boxes)
157 (golden comparator) ~ 095 (PTY view)
[111 shipped] ──► 090
[119 121 shipped] ──► 120
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
| 20 — Developer feedback & verification tools (096, 107 Shipped; 095 Active) | [095](../issues/095-add-human-observable-pty-test-view-mode-and-feedback-flow.md), [096](../issues/096-follow-up-038-with-a-non-ascii-text-rendering-example-app.md) |
| 21 — Mouse correctness (Shipped) | [107](../issues/107-pty-mouse-test-system-detect-loomoji-hover-vs-highlight-offset-via-rendered-fg-bg-cells.md), [108](../issues/108-fix-loomoji-hover-highlight-one-row-above-pointer-dy-1.md) |
| 22 — loom CLI and asset validation (Shipped) | [128](../issues/128-add-box-frame-geometry-and-line-width-validator-for-ansi-assets.md), [129](../issues/129-build-loom-cli-tool-for-tui-asset-validation-measurement-and-interactive-viewing.md), [130](../issues/130-add-loom-format-command-to-auto-align-re-pad-and-normalize-ansi-files.md), [131](../issues/131-add-loom-frame-command-to-wrap-ansi-text-in-styled-box-borders.md) |
| 23 — Real-host proof and media (Shipped) | [042](../issues/042-docs-tuiinput-md-referenced-by-5-code-comments-but-does-not-exist.md), [103](../issues/103-provide-a-hostable-migration-loop-for-coexisting-loom-views.md), [111](../issues/111-add-an-image-media-widget-rendered-with-cati.md) |
| 24 — 2D panning and lazy media (Shipped) | [138](../issues/138-allow-panning-in-loom-panes-and-make-ansiviewer-preview-pane-pannable.md), [140](../issues/140-lazy-load-media-widget-images-and-video-previews.md) |
| 25 — Media example, ansiviewer desktop, release (Shipped) | 141 to 146, 148 to 153, 159 |
| Closed — Data sources & targets (parked) | [014](../issues/014-configurable-graph-colors-and-glyph-presentation.md), [015](../issues/015-simulated-voxi-transcript-and-daemon-panels.md), [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md), [017](../issues/017-external-file-and-socket-adapters-with-separate-producer-fixtures.md), [018](../issues/018-explore-bounded-linux-and-daemon-source-adapters.md), [019](../issues/019-evaluate-declarative-source-and-action-wiring.md), [056](../issues/056-embed-real-applications-as-pty-hosted-widgets-tmux-screen-style.md) |

## Long-term outcome

The completed target is a declarative Loom application whose document defines
the title/status chrome, responsive boxes, rows, bars, timelines, colors,
visibility controls, and data bindings. Go defines collectors, external
integration, state transitions, and action behavior. The renderer remains
independent of collection frequency and application-specific coordinates.
</content>
