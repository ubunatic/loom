<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# loom — issues

**v0.1.0 is released** — tagged, pushed, and consumed by uzu through the public
module proxy. The extraction (copy, module path, import rewrite, green build) and
the release itself are done; what remains is the downstream cutover and doc
upkeep.

| # | File | Title | Status |
|---|------|-------|--------|
| 001 | [001-first-release-v0.1.0.md](001-first-release-v0.1.0.md) | Cut the first release (v0.1.0) | Done — v0.1.0 tagged and pushed; README refreshed 2026-08-04 |
| 002 | [002-pane-driver-helper.md](002-pane-driver-helper.md) | Ship an exported pane-driver helper | Done in loom (2026-07-04) — consumer cutover tracked in 004 |
| 003 | [003-public-api-audit.md](003-public-api-audit.md) | Audit and document the public API surface | Done (2026-07-04) |
| 004 | [004-migrate-uzu-to-shared-loom.md](004-migrate-uzu-to-shared-loom.md) | Migrate uzu to depend on the shared loom module | Closed — invalid |
| 005 | [005-license-clarification.md](005-license-clarification.md) | Confirm AGPL is the intended license for a shared library | Resolved — keep AGPL-3.0-or-later (2026-07-04) |
| 006 | [006-static-declarative-monitor-shell-with-a-minimal-validated-contract.md](006-static-declarative-monitor-shell-with-a-minimal-validated-contract.md) | Static declarative monitor shell with a minimal validated contract | Closed — static declarative shell implemented and verified |
| 007 | [007-clock-watch-mode-with-independent-collection-and-redraw.md](007-clock-watch-mode-with-independent-collection-and-redraw.md) | Clock watch mode with independent collection and redraw | Closed — watch clock with independent cadence and verified terminal restoration |
| 008 | [008-responsive-declared-box-layout.md](008-responsive-declared-box-layout.md) | Responsive declared box layout | Closed — responsive layout verified headlessly and through PTY resize |
| 009 | [009-declarative-box-visibility-controls.md](009-declarative-box-visibility-controls.md) | Declarative box visibility controls | Closed — declared visibility actions and hints verified |
| 010 | [010-geometry-and-visual-evidence-milestone-before-rich-content.md](010-geometry-and-visual-evidence-milestone-before-rich-content.md) | Geometry and visual evidence milestone before rich content | Closed — independent geometry gate and reproducible ANSI evidence pass; visual review requested |
| 011 | [011-aligned-dashboard-rows-and-ansi-safe-truncation.md](011-aligned-dashboard-rows-and-ansi-safe-truncation.md) | Aligned dashboard rows and ANSI-safe truncation | Closed — aligned dashboard rows, left/right truncation, graph placeholders, and dynamic value binding implemented and verified |
| 012 | [012-copy-harnez-rograph-with-verified-provenance-and-bounded-adapters.md](012-copy-harnez-rograph-with-verified-provenance-and-bounded-adapters.md) | Copy Harnez rograph with verified provenance and bounded adapters | Closed |
| 013 | [013-deterministic-live-snapshots-and-independent-rolling-histories.md](013-deterministic-live-snapshots-and-independent-rolling-histories.md) | Deterministic live snapshots and independent rolling histories | Closed |
| 014 | [014-configurable-graph-colors-and-glyph-presentation.md](014-configurable-graph-colors-and-glyph-presentation.md) | Configurable graph colors and glyph presentation | Open |
| 015 | [015-simulated-voxi-transcript-and-daemon-panels.md](015-simulated-voxi-transcript-and-daemon-panels.md) | Simulated Voxi transcript and daemon panels | Open |
| 016 | [016-complete-harnez-and-voxi-simulated-ui-milestone.md](016-complete-harnez-and-voxi-simulated-ui-milestone.md) | Complete Harnez and Voxi simulated UI milestone | Open |
| 017 | [017-external-file-and-socket-adapters-with-separate-producer-fixtures.md](017-external-file-and-socket-adapters-with-separate-producer-fixtures.md) | External file and socket adapters with separate producer fixtures | Open |
| 018 | [018-explore-bounded-linux-and-daemon-source-adapters.md](018-explore-bounded-linux-and-daemon-source-adapters.md) | Explore bounded Linux and daemon source adapters | Open |
| 019 | [019-evaluate-declarative-source-and-action-wiring.md](019-evaluate-declarative-source-and-action-wiring.md) | Evaluate declarative source and action wiring | Open |
| 020 | [020-review-ticket-011-rows-and-target-state-alignment.md](020-review-ticket-011-rows-and-target-state-alignment.md) | Review of Ticket 011 increments and target state alignment | Closed — proposed alignment changes implemented and verified across tickets 011, 021, 022, 023 |
| 021 | [021-cmd-help-tty-blocking-in-tests.md](021-cmd-help-tty-blocking-in-tests.md) | Fix :help command blocking on interactive /dev/tty in test environments | Closed — decoupled showHelp from /dev/tty via headless detection and test hook |
| 022 | [022-decouple-static-shell-golden-from-evolving-monitor-spec.md](022-decouple-static-shell-golden-from-evolving-monitor-spec.md) | Decouple static shell golden tests from evolving example monitor spec | Closed — extracted dedicated empty-shell fixture and removed in-memory mutation |
| 023 | [023-rows-schema-validation-and-negative-controls.md](023-rows-schema-validation-and-negative-controls.md) | Add negative schema controls and column-value validation for box rows | Closed — added comprehensive negative schema controls for rows in validate-spec.py |
| 024 | [024-port-harnez-rograph-primitives-with-provenance.md](024-port-harnez-rograph-primitives-with-provenance.md) | Port Harnez rograph primitives with provenance | Closed |
| 025 | [025-integrate-graph-renderers-into-declarative-monitor.md](025-integrate-graph-renderers-into-declarative-monitor.md) | Integrate graph renderers into declarative monitor | Closed — resolved in e944324 |
| 026 | [026-investigate-monitor-pty-smoke-test-idle-redraw-regression.md](026-investigate-monitor-pty-smoke-test-idle-redraw-regression.md) | Investigate monitor PTY smoke-test idle redraw regression | Closed — resolved in watch-pty verification target |
| 027 | [027-introduce-first-spec-driven-collector-prototype.md](027-introduce-first-spec-driven-collector-prototype.md) | Introduce first spec-driven collector prototype | Closed |
| 028 | [028-add-reusable-measurement-and-dynamic-box-layout-primitives.md](028-add-reusable-measurement-and-dynamic-box-layout-primitives.md) | Add reusable measurement and dynamic box layout primitives | Closed — resolved in 5e8cac5 |
| 029 | [029-declarative-centered-layout-and-viewport-alignment-primitives.md](029-declarative-centered-layout-and-viewport-alignment-primitives.md) | Declarative Centered Layout and Viewport Alignment Primitives | Closed |
| 030 | [030-braille-activity-spinner-and-bracketed-progress-bar-primitives.md](030-braille-activity-spinner-and-bracketed-progress-bar-primitives.md) | Braille Activity Spinner and Bracketed Progress Bar Primitives | Closed |
| 031 | [031-provider-status-pill-cluster-and-lifecycle-state-presentation.md](031-provider-status-pill-cluster-and-lifecycle-state-presentation.md) | Provider Status Pill Cluster and Lifecycle State Presentation | Closed |
| 032 | [032-splash-lifecycle-controller-async-provider-coordination-and-key-dismissal.md](032-splash-lifecycle-controller-async-provider-coordination-and-key-dismissal.md) | Splash Lifecycle Controller, Async Provider Coordination, and Key Dismissal | Closed |
| 033 | [033-harnez-target-splash-screen-integration-and-golden-tests.md](033-harnez-target-splash-screen-integration-and-golden-tests.md) | Harnez Target Splash Screen Integration and Golden Tests | Closed |
| 034 | [034-ansi-sgr-escape-sequence-parsing-and-writeansi-canvas-helper.md](034-ansi-sgr-escape-sequence-parsing-and-writeansi-canvas-helper.md) | ANSI SGR Escape Sequence Parsing and WriteANSI Canvas Helper | Open |
| 035 | [035-extend-loom-view-navigation-keybindings-and-export-terminalsize-helper.md](035-extend-loom-view-navigation-keybindings-and-export-terminalsize-helper.md) | Extend loom.View Navigation Keybindings and Export TerminalSize Helper | Open |
