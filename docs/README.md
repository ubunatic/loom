<!-- managed by harnez index: only the studies table is generated; evergreen documentation and other tables are hand-written -->
# Loom Documentation

- [Roadmap](Roadmap.md): assessed stages and implementation tickets.
- [Roadmap context](RoadmapContext.md): product vision and session summary.
- [UI targets (Usage Dashboard)](HarnezUsageTarget.md): Harnez and Voxi reference layouts (tables, boxes, sparklines).
- [UI targets (Splash Screen)](HarnezSplashTarget.md): Harnez startup splash & loading screen reference.
- [Graph primitives](Graph.md): determinate bars, rolling sparklines, 2D treemaps, and width padding invariants.
- [Monitor example](../examples/monitor/README.md): show-once/watch, responsive boxes and controls.
- [Geometry gate](Geometry.md): supported text policy, independent checks and ANSI replay evidence.
- [Hover testing](HoverTesting.md): black-box PTY mouse-hover probes, colour-cell diffing and offset reporting.
- [Lean sprints](LeanSprints.md): loom field notes for lean sprints with cheap developer agents (escalation ladder, fake-test review, evidence convention).
- [Editor](Editor.md): approved editor decisions, ANSI validation, widget composition, input safety, and PR review lessons.
- [.tui markup](TuiMarkup.md): bracket markup grammar, colors, escapes, strict errors, `loom edit`/`loom markup` support.
- [loom session](Session.md): PTY wrapper with tabs, bar, help, mouse modes, terminal restore and scrollback search.
- [Terminal safety](TerminalSafety.md): the auto-wrap corruption trap, `loom.RawScreen`/`WriteRows`/`ClipRow`, and `x/term` coverage rules for any raw-ANSI terminal writer.
- [Terminal repaint probe](TerminalRepaintProbe.md): standalone `loom-repaint-probe` benchmark modes, tuning, and timing semantics.
- [Terminal colors](TerminalColors.md): authoritative theme colors, shade glyphs, terminal dimming, and scrollbar experiments.
- [Themes](Themes.md): spec-driven palettes, semantic roles, widget adapters, runtime switching, and known boundaries.
- [Animated backgrounds](AnimatedBackgrounds.md): Astra-style deterministic Braille star fields, protected-cell rendering, and the custom-effect contract.
- [Root overlays](RootOverlays.md): the root-level-overlay hook pattern (`paneHelpRequest`) for modals that must draw over an entire split layout, not just a `paintClipped` child.
- [Key defaults](KeyDefaults.md): decoded keys, default actions per library widget, terminal limitations.
- [Terminal input](TuiInput.md): controlling tty ownership, interruptible input reads, and key decoding compatibility, modified Home/End/Insert/Delete, CSI-u, and VTE-reserved Shift keys.
- [Progress evidence](progress/): `.ansi` frames per ticket; view with `for f in docs/progress/*/*.ansi; do echo "== $f"; cat "$f"; done`.
- [Emoji & Unicode measurement](EmojiWidth.md): authoritative spec system, VTE render modes (pad-1), cluster parsing, and PTY width invariants.
- [Upgrading](Upgrading.md): breaking changes by release, the `ubunatic.com/loom` module path, switching dependents, bool → `EventResult` mapping.
- [Widgets & framework primitives](Widgets.md): split layout, dynamic tabs, metric stores, directory navigation, and startup transition runners, progress bar, and the input widget inventory.

## Case Studies

**`docs/studies/`** — case studies, indexed by `harnez index`.

<!-- harnez:begin docs-studies-index -->
| File | Topic |
|------|-------|
| [studies/2026-09-09-from-monitor-screenshots-to-a-declarative-ui-backlog.md](studies/2026-09-09-from-monitor-screenshots-to-a-declarative-ui-backlog.md) | From Monitor Screenshots to a Declarative UI Backlog |
| [studies/2026-09-10-agentic-sprint-from-mid-flight-rows-to-stable-graph-milestone.md](studies/2026-09-10-agentic-sprint-from-mid-flight-rows-to-stable-graph-milestone.md) | Agentic Sprint from Mid-Flight Rows to Stable Graph Milestone |
| [studies/2026-09-11-splash-screen-architecture-and-runtime-safeguards.md](studies/2026-09-11-splash-screen-architecture-and-runtime-safeguards.md) | Splash Screen Architecture and Runtime Safeguards |
| [studies/2026-09-14-terminal-safety-hardening-and-treemap-theme-iteration.md](studies/2026-09-14-terminal-safety-hardening-and-treemap-theme-iteration.md) | Terminal Safety Hardening and Treemap Theme Iteration |
| [studies/2026-09-14-x-term-coverage-gap-in-pane-termsize.md](studies/2026-09-14-x-term-coverage-gap-in-pane-termsize.md) | x/term Coverage Gap in Pane.termSize |
| [studies/2026-09-17-remove-deprecated-uzu-brand-and-home-global-command.md](studies/2026-09-17-remove-deprecated-uzu-brand-and-home-global-command.md) | Remove Deprecated `uzu` Brand and `:home` Global Command |
| [studies/2026-09-19-compositor-layering-and-root-modal-overlays.md](studies/2026-09-19-compositor-layering-and-root-modal-overlays.md) | Compositor Layering and Root Modal Overlays |
| [studies/2026-09-20-examples-modernization-and-framework-primitives-audit.md](studies/2026-09-20-examples-modernization-and-framework-primitives-audit.md) | Examples Modernization & Framework Primitives Adoption |
| [studies/2026-09-20-feature-gap-analysis-background.md](studies/2026-09-20-feature-gap-analysis-background.md) | Background Example Feature-Gap Analysis |
| [studies/2026-09-20-feature-gap-analysis-filebrowser.md](studies/2026-09-20-feature-gap-analysis-filebrowser.md) | Filebrowser Feature Gap Analysis |
| [studies/2026-09-20-feature-gap-analysis-monitor.md](studies/2026-09-20-feature-gap-analysis-monitor.md) | Monitor Example Feature Gap Analysis |
| [studies/2026-09-20-feature-gap-analysis-screens.md](studies/2026-09-20-feature-gap-analysis-screens.md) | Screens Example Feature-Gap Analysis |
| [studies/2026-09-20-feature-gap-analysis-splash.md](studies/2026-09-20-feature-gap-analysis-splash.md) | Splash Example Feature Gap Analysis |
| [studies/2026-09-20-feature-gap-analysis-split.md](studies/2026-09-20-feature-gap-analysis-split.md) | Split Example Feature Gap Analysis |
| [studies/2026-09-20-feature-gap-analysis-tabs.md](studies/2026-09-20-feature-gap-analysis-tabs.md) | Tabs Example Feature-Gap Analysis |
| [studies/2026-09-20-feature-gap-analysis-treemap.md](studies/2026-09-20-feature-gap-analysis-treemap.md) | Treemap Feature-Gap Analysis |
| [studies/2026-09-20-feature-gap-analysis-winch.md](studies/2026-09-20-feature-gap-analysis-winch.md) | Winch Feature Gap Analysis |
| [studies/2026-09-20-harnez-agent-lean-sprint-orchestration-and-pitfalls.md](studies/2026-09-20-harnez-agent-lean-sprint-orchestration-and-pitfalls.md) | Practical Orchestration and Pitfall Analysis of `harnez agent` in Lean Sprints |
| [studies/2026-09-20-luna-lean-sprint-candidates.md](studies/2026-09-20-luna-lean-sprint-candidates.md) | Roadmap Items Suited to `codex:luna:low` Lean Sprints |
| [studies/2026-09-20-session-token-usage-and-cost-analysis.md](studies/2026-09-20-session-token-usage-and-cost-analysis.md) | Session Token Usage and Cost Analysis |
| [studies/2026-09-21-lean-sprint-session-report.md](studies/2026-09-21-lean-sprint-session-report.md) | Lean Sprint Session Report: Roadmap Items with Cheap Developer Agents |
| [studies/2026-09-24-emoji-width-measurement-consolidation-and-vte-modes.md](studies/2026-09-24-emoji-width-measurement-consolidation-and-vte-modes.md) | Emoji Width Measurement Consolidation and VTE Modes |
| [studies/2026-09-24-roadmap-now-sprint-099-101-105.md](studies/2026-09-24-roadmap-now-sprint-099-101-105.md) | Roadmap "Now" sprint: 099, 101, 105 |
| [studies/2026-09-26-textedit-syntax-and-agentic-failover.md](studies/2026-09-26-textedit-syntax-and-agentic-failover.md) | `examples/textedit`, `codeberg.org/ubunatic/loom/syntax`, `loom.TextArea`, multi-provider agent orchestration |
| [studies/2026-09-treesitter-syntax-engine.md](studies/2026-09-treesitter-syntax-engine.md) | Pure-Go / Wasm (wazero) Tree-Sitter Syntax Engine for Loom |
| [studies/2026-10-01-loom-module-migration.md](studies/2026-10-01-loom-module-migration.md) | Loom module migration reports (issue 235 M3) |
| [studies/2026-10-04-inline-tui-library-agentic-build-story.md](studies/2026-10-04-inline-tui-library-agentic-build-story.md) | Loom: building a terminal UI library in fast, testable slices |
| [studies/2026-10-09-tui-markup-format-lean-sprint-story.md](studies/2026-10-09-tui-markup-format-lean-sprint-story.md) | Rapid Evolution of the .tui Markup Format via Harnez Lean Sprints |
| [studies/2026-10-10-two-agent-lean-sprint-issue-315-windows-measure.md](studies/2026-10-10-two-agent-lean-sprint-issue-315-windows-measure.md) | 2026-10-10 Lean Sprint Study: Windows Builds in Terminal Measurement (Issue 315) |
| [studies/2026-10-10-two-agent-lean-sprint-issue-316-popover-semantics.md](studies/2026-10-10-two-agent-lean-sprint-issue-316-popover-semantics.md) | 2026-10-10 Lean Sprint Study: Configurable Popover Menu Semantics in loom edit (Issue 316) |
| [studies/2026-10-10-two-agent-lean-sprint-issue-317-directional-popover-focus.md](studies/2026-10-10-two-agent-lean-sprint-issue-317-directional-popover-focus.md) | 2026-10-10 Lean Sprint Study: Directional Popover Focus and Tab/Ctrl+I Decoupling (Issue 317) |
| [studies/2026-10-10-two-agent-lean-sprint-issue-318-cancel-focused-popover.md](studies/2026-10-10-two-agent-lean-sprint-issue-318-cancel-focused-popover.md) | 2026-10-10 Lean Sprint Study: Cancel Focused Popover on Arrow Away from Menu Bar (Issue 318) |
| [studies/2026-10-20-two-agent-lean-sprint-issue-342-popover-toggle-mousegrab.md](studies/2026-10-20-two-agent-lean-sprint-issue-342-popover-toggle-mousegrab.md) | Two-Agent Lean Sprint: Issue 342 (Ctrl+Space popover toggle, mousegrab default) |
| [studies/2026-10-21-two-agent-lean-sprint-issue-346-colorpicker.md](studies/2026-10-21-two-agent-lean-sprint-issue-346-colorpicker.md) | Two-Agent Lean Sprint: Issue 346 (ColorPicker widget, 🎨 popover item) |
| [studies/2026-10-21-two-agent-lean-sprint-issue-347-colorpicker-redesign.md](studies/2026-10-21-two-agent-lean-sprint-issue-347-colorpicker-redesign.md) | Two-Agent Lean Sprint: Issue 347 (ColorPicker redesign) |
| [studies/2026-10-21-two-agent-lean-sprint-issue-353-cspace-menu.md](studies/2026-10-21-two-agent-lean-sprint-issue-353-cspace-menu.md) | Two-Agent Lean Sprint: Issue 353 (Ctrl+Space menu labels and hotkeys) |
| [studies/2026-10-21-two-agent-lean-sprint-issues-348-325-340-341-343.md](studies/2026-10-21-two-agent-lean-sprint-issues-348-325-340-341-343.md) | Two-Agent Lean Sprint: Issues 348, 325, 340, 341, 343 |
| [studies/2026-10-21-two-agent-lean-sprint-issues-349-344-345-350.md](studies/2026-10-21-two-agent-lean-sprint-issues-349-344-345-350.md) | Two-Agent Lean Sprint: Issues 349, 344, 345, 350 |
| [studies/2026-10-21-two-agent-lean-sprint-issues-355-356-358-357.md](studies/2026-10-21-two-agent-lean-sprint-issues-355-356-358-357.md) | Two-Agent Lean Sprint: Issues 355, 356, 358, 357 |
| [studies/2026-10-22-lean-sprints-spec-shadows-keybindings-v040.md](studies/2026-10-22-lean-sprints-spec-shadows-keybindings-v040.md) | Lean sprints 2026-10-22 (second half): reference crops, PTY flake, spec shadows, key bindings, v0.4.0 |
| [studies/2026-10-23-honing-test-tooling-for-fast-feedback.md](studies/2026-10-23-honing-test-tooling-for-fast-feedback.md) | Honing Test Tooling for Fast Feedback |
| [studies/2026-10-23-loom-session-tabs-search-and-modularization-v050.md](studies/2026-10-23-loom-session-tabs-search-and-modularization-v050.md) | loom session: tabs, search and modularization (v0.5.0) |
| [studies/2026-10-23-session-retro-editor-colorpicker-session.md](studies/2026-10-23-session-retro-editor-colorpicker-session.md) | Session retro: editor bar, color picker, loom session (2026-10-20 to 2026-10-23) |
| [studies/2026-10-23-test-optimization-and-root-modularization.md](studies/2026-10-23-test-optimization-and-root-modularization.md) | Test Optimization and Root Modularization |
| [studies/2026-10-24-multiplexer-ansi-leakage-and-diagnostic-recovery.md](studies/2026-10-24-multiplexer-ansi-leakage-and-diagnostic-recovery.md) | Multiplexer ANSI Leakage, VT State Machines, and Diagnostic Recovery |
| [studies/2026-10-24-multiplexer-output-framing-and-vt-streaming-continuity.md](studies/2026-10-24-multiplexer-output-framing-and-vt-streaming-continuity.md) | Issue #382 (VT parser streaming continuity and safe discard for split/unsupported escape sequences) |
| [studies/2026-10-24-two-agent-lean-sprint-issue-382-vt-streaming-continuity.md](studies/2026-10-24-two-agent-lean-sprint-issue-382-vt-streaming-continuity.md) | Two-Agent Lean Sprint Study: Issue #382 VT Parser Streaming Continuity and Safe Discard |
<!-- harnez:end docs-studies-index -->


Study files are the source of truth for this table.
