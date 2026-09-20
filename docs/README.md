# Loom Documentation

- [Roadmap](Roadmap.md): assessed stages and implementation tickets.
- [Roadmap context](RoadmapContext.md): product vision and session summary.
- [UI targets (Usage Dashboard)](HarnezUsageTarget.md): Harnez and Voxi reference layouts (tables, boxes, sparklines).
- [UI targets (Splash Screen)](HarnezSplashTarget.md): Harnez startup splash & loading screen reference.
- [Graph primitives](Graph.md): determinate bars, rolling sparklines, 2D treemaps, and width padding invariants.
- [Monitor example](../examples/monitor/README.md): show-once/watch, responsive boxes and controls.
- [Geometry gate](Geometry.md): supported text policy, independent checks and ANSI replay evidence.
- [Terminal safety](TerminalSafety.md): the auto-wrap corruption trap, `loom.RawScreen`/`WriteRows`/`ClipRow`, and `x/term` coverage rules for any raw-ANSI terminal writer.
- [Terminal colors](TerminalColors.md): authoritative theme colors, shade glyphs, terminal dimming, and scrollbar experiments.
- [Themes](Themes.md): spec-driven palettes, semantic roles, widget adapters, runtime switching, and known boundaries.
- [Animated backgrounds](AnimatedBackgrounds.md): Astra-style deterministic Braille star fields, protected-cell rendering, and the custom-effect contract.
- [Root overlays](RootOverlays.md): the root-level-overlay hook pattern (`paneHelpRequest`) for modals that must draw over an entire split layout, not just a `paintClipped` child.
- [Widgets & framework primitives](Widgets.md): split layout, dynamic tabs, metric stores, directory navigation, and startup transition runners.

## Case Studies

**`docs/studies/`** — case studies, indexed by `harnez index`.

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

Study files are the source of truth for this table.
