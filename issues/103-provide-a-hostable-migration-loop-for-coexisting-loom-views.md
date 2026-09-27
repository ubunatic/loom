# 103 — Port minimal colored `harnez usage --compact --watch`

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature

## Goal

Port the complete `../harnez` `harnez usage --compact --watch` execution path
into a minimal, standalone Loom example. It must be runnable and fully colored,
with the two All Usage and Load boxes, asynchronous data collection, an
independent redraw loop, terminal resize/input handling, and realistic local
data for Load plus fake All Usage data.

The port should preserve the useful Harnez behavior rather than replacing it
with a placeholder renderer. Once the colored clone is working and covered by
tests/ANSI evidence, use it as the host application for the Loom migration:
introduce Loom views incrementally, gate old/new views behind CLI flags, allow
both views to coexist, and provide an interactive switch between them.

The SDK changes required by the migration should be kept small and generic.
File separate follow-up issues for Loom capabilities that are discovered but
are not necessary to complete this port.

## Clarification (user, 2026-09-27)

"Port" here means **rebuild**, not copy: re-implement the behavior on Loom widgets, or use the
strangler pattern, where old and new views coexist behind flags and the Loom views gradually
replace the old ones. Do not copy Harnez's rendering code across. Reuse only its data
collection and its behavior as the reference.

## Sprint Log

**Plan (dev-103, 2026-09-27), adjusted by host:**
- **M1: `examples/usage` rebuild.** Two colored boxes, All Usage (deterministic fake quota data) and Load (real local readings), built on Loom `Frame`/`Box` and hostable like `examples/monitor` (064): `NewWidget(args)`, `Ticker`, `Close()`, `examplesreg` registration. Collection is injectable and asynchronous. Tests cover layout, narrow widths and resize, and a Draw/Tick capture test that stays silent.
- **M2: redraw and evidence.** Repaint after collection through `InvalidationAware` (`widget.go`, added in 065 M4). The PTY colour-cell check follows the 107 method. The `.ansi` snapshot goes to `/tmp` for the user.
- **M3: strangler switch inside the example.** A flag selects between the plain view (a minimal rebuild of the Harnez layout as colored rows via `StyledRows`) and the Loom-widget view, with an interactive key to switch between them. Both share one data model.
- **Out of scope here:** wiring this into the `../harnez` repo. That needs its own harnez ticket once M3 exists.

**Pre-Work (host):**
1. The plan says `StyledRows` and `InvalidationAware` were not found. They exist: `styledrows.go` and `widget.go` (065). Use them rather than reinventing them.
2. Before coding, write into this ticket the reference behavior taken from `../harnez` (what each box shows: fields, units, colors and thresholds), with file:line references, so the review can check parity.
