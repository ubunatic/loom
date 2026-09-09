---
title: Loom Roadmap
weight: 20
---

# Loom Roadmap

This is an aspirational product roadmap for turning Loom into a simple,
spec-driven terminal UI system. The stages below are tracked by tickets
006–019 in the [issue index](../issues/README.md).

The value axis is simple: an application should declare its UI structure and
presentation, while Go supplies behavior and changing data. The UI should be
easy to maintain, responsive to terminal size, testable without a live
terminal, and free from application-specific coordinate/layout code.

The primary reference targets are documented in
[`HarnezUsageTarget.md`](HarnezUsageTarget.md), including the Harnez usage
dashboard and the Voxi monitor.

## Assessment against code and backlog

Assessed 2026-09-09. The findings below describe the planning baseline before
ticket 006. Stages 1–5 (tickets 006–010) have now completely shipped, establishing
the static shell, initial schemas, watch mode, responsive layout, controls, and
geometry gate. Later monitor capabilities and active bug fixes remain open.

- [`yaml.go`](../yaml.go) already builds widgets and routes views without a
  terminal, but is a narrow widget factory: `StyleName` and `OnChange` are
  declared without wiring. There is no implemented `spec/` or JSON Schema
  coverage; [`Spec.md`](Spec.md) describes the intended contract.
- [`stack.go`](../stack.go) divides space equally; [`grid.go`](../grid.go)
  uses uniform cells. Neither supplies responsive boxes, chrome, padding, or
  visibility reflow. [`pane.go`](../pane.go) redraws on input/resize, without
  an independent periodic redraw scheduler.
- Reuse [`canvas.go`](../canvas.go), [`style.go`](../style.go), and
  [`screenshot.go`](../screenshot.go), but strengthen their geometry contract:
  `RuneWidth` counts combining marks as one, `StringWidth` counts ANSI sequences
  as printable runes, and `Write` clips to the canvas rather than the child
  rectangle. [`view.go`](../view.go) strips ANSI and truncates by rune count.
  `Render` produces ANSI rows; `ScreenshotScript` replays ANSI, rather than
  providing a terminal-emulator or image oracle. Existing render tests are a
  foundation, not evidence that the target layout is correct.
- Harnez's `../harnez/internal/rograph/{bar,sparkline,options}.go` supplies
  bars, fixed-width timelines, Braille, and configurable presentation. Copy
  or adapt the renderer initially (its `internal` path cannot be imported
  directly by Loom). Voxi's `../voxi/internal/monitor/render.go` supplies
  box and ANSI truncation behavior worth adapting and testing. Neither proves
  integration with Loom's cell/style model; keep domain collectors outside it.
- At the initial assessment, `harnez find` reported only [004, uzu migration](../issues/004-migrate-uzu-to-shared-loom.md)
  open, with no appended implementation plan. Issues 001–003 are marked done;
  005 is resolved. Release and driver work are foundations already shipped,
  not new roadmap dependencies. Tickets 011–023 cover the monitor capabilities,
  hygiene, and test reliability work.

### Close / Park

004 is a closure candidate: the README says uzu consumes the released module,
while the ticket remains open. A sibling uzu checkout is unavailable here, so
verify its imports, dependency, removed copy, and checks before closing it.
This reconciliation need not block the monitor shell. Park generalized dataflow
and cross-project graph extraction until the small examples establish need.
Reuse is evidenced above; no measured cost-savings percentage is established.

### Ticket sequencing

**Shipped:** Stages 1–5 (tickets 006–010), Stage 6 rows (tickets 011 and 020), and test/schema hygiene (tickets 021, 022, 023).
**Now:** Ticket 012 (copy Harnez rograph with verified provenance and bounded adapters) to introduce graph primitives.
**Next:** Stages 7–9 (tickets 013–016), introducing simulated live snapshots/histories, color/glyphs, and completing the combined milestone.
**Later:** Stages 10–11 (tickets 017–019), introducing external sources and declarative wiring.
Each stage should leave a runnable example and deterministic checks. Keep examples in the root module so root test discovery includes them; the target document's nested `go.mod` is illustrative, not a requirement.

| Stage | Tickets |
|---|---|
| 1–5 — Shipped Foundation | [006](../issues/006-static-declarative-monitor-shell-with-a-minimal-validated-contract.md), [007](../issues/007-clock-watch-mode-with-independent-collection-and-redraw.md), [008](../issues/008-responsive-declared-box-layout.md), [009](../issues/009-declarative-box-visibility-controls.md), [010](../issues/010-geometry-and-visual-evidence-milestone-before-rich-content.md) |
| Hygiene & Fixes (Shipped) | [021](../issues/021-cmd-help-tty-blocking-in-tests.md), [022](../issues/022-decouple-static-shell-golden-from-evolving-monitor-spec.md), [023](../issues/023-rows-schema-validation-and-negative-controls.md) |
| 6 — Rows (Shipped) & Graphs | [011](../issues/011-aligned-dashboard-rows-and-ansi-safe-truncation.md) (done), [020](../issues/020-review-ticket-011-rows-and-target-state-alignment.md) (done), [012](../issues/012-copy-harnez-rograph-with-verified-provenance-and-bounded-adapters.md) (now; sub-tickets [024](../issues/024-port-harnez-rograph-primitives-with-provenance.md), [025](../issues/025-integrate-graph-renderers-into-declarative-monitor.md)) |
| 7 — Simulated live data | [013](../issues/013-deterministic-live-snapshots-and-independent-rolling-histories.md) |
| 8 — Color and glyphs | [014](../issues/014-configurable-graph-colors-and-glyph-presentation.md) |
| 9 — Complete simulated targets | [015](../issues/015-simulated-voxi-transcript-and-daemon-panels.md), [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md) |
| 10 — External sources | [017](../issues/017-external-file-and-socket-adapters-with-separate-producer-fixtures.md), [018](../issues/018-explore-bounded-linux-and-daemon-source-adapters.md) |
| 11 — Declarative source prototype | [019](../issues/019-evaluate-declarative-source-and-action-wiring.md) |

## Stage 1 — Static monitor shell

Implemented in ticket 006. Run `GOWORK=off go run ./examples/monitor`;
see the [example and contract](../examples/monitor/README.md). The shell uses
declared dimensions, printable ASCII chrome/text, and spec-defined borders.
Schema validation, existing Go tests, baseline geometry tests, and a no-TTY
smoke run pass. This does not complete the broader Stage 5 visual gate.

Build the smallest useful Loom application in show-once mode.

It should render:

- a title bar;
- two empty titled boxes;
- a bottom status bar.

There is no `--watch` mode, live data, graph, or button behavior yet. The goal
is to establish the application frame, generic box rendering, basic spacing,
and a small declarative UI example with a minimal Go entrypoint.

Include only the shell's declarative vocabulary in an embedded YAML spec with
a companion JSON Schema and validation/loader-fidelity tests. Titles, spacing,
and other declared values must be consumed rather than duplicated in Go. Grow
schema coverage with each stage; a comprehensive framework is not a prerequisite.

Start geometry coverage here: fixed terminal-cell fixtures for chrome, borders,
padding, and child clipping, including tiny bounds. Check terminal positions
independently of the renderer's own width helpers. Expand this baseline through
responsive layout and controls before the full Stage 5 milestone.

## Stage 2 — Basic watch mode and clock

Add a simple watch mode using the current time as the first data source.

The clock is collected independently from rendering. The UI redraws at an
explicit refresh rate, while the time source and redraw loop remain separate
concerns. Show-once mode must continue to work.

This stage establishes the fundamental live-data contract before controls or
domain-specific content are introduced.

## Stage 3 — Responsive box layout

Make the two boxes responsive to terminal width.

- At sufficient width, place the boxes next to each other.
- Below the layout breakpoint, place them in a vertical stack.
- Preserve titles, borders, spacing, and usable inner dimensions in both modes.

The same declared layout should produce both arrangements. It should not
require separate wide and slim UI implementations.

## Stage 4 — Box visibility controls

Add keyboard controls for hiding and showing the boxes, with visible hints in
the title or status bar.

The control model should be declarable, but the first implementation can keep
the action wiring small and explicit. The layout must recompute correctly when
a box is hidden or restored.

## Stage 5 — Visual layout contract and testing milestone

Before adding substantial content, establish a stable way to test geometry.

Promote the early baseline to a canonical terminal-cell fixture for wide, narrow,
and tiny terminals, resize, and hidden/restored boxes. Reference screenshots
vary with data, time, and font, and clipboard copies may be imperfect; fix inputs
and document intentional deviations rather than treating those images as exact
goldens. Independently check terminal positions, including ANSI, combining and
wide characters, child clipping, and border integrity; ANSI replay alone is not
the oracle.

The test system should detect when content damages box borders, causes width
drift, or misaligns rows. The implementation may use a combination of:

- robust ANSI-aware display-width measurement;
- explicit border/rectangle invariants;
- deterministic canvas rendering;
- golden or screenshot-style output comparisons.

This is a milestone requiring visual confirmation when a user is available.
If no user is available, automated geometry and screenshot tests should provide
the gate to the next stage.

## Stage 6 — Static dashboard content and graph primitives

Add the first realistic content using simulated, fixed snapshots.

The target is the Harnez-style dashboard:

- aligned label/value rows;
- determinate bars;
- fixed-width rolling sparklines;
- multiple timelines in one box;
- a split VRAM/GTT row;
- truncation for values that do not fit.

Reuse or copy Harnez's dependency-free `rograph` implementation initially.
Keep collection and application-specific metric names outside the Loom graph
renderer.

## Stage 7 — Simulated live data

Connect deterministic simulated data to the dashboard.

The simulation should update snapshots and histories independently from the
render loop. Tests should cover changing values, rolling histories, startup
with short histories, stable widths, and repeated redraws without accidental
history advancement.

This proves changing monochrome content in watch mode. Complete target
acceptance follows color and the Voxi content integration in Stage 9.

## Stage 8 — Color and visual semantics

Add configurable colors after the monochrome geometry is stable.

- Map bar levels through a color range such as green to red.
- Apply equivalent level-aware color to sparkline/timeline output.
- Support Braille and other configured glyph presentations.
- Ensure ANSI styling never changes measured geometry.
- Keep visual values configurable rather than embedding application-specific
  palettes in widgets.

## Stage 9 — Complete simulated target milestone

Demonstrate both Harnez and Voxi targets in show-once and watch modes after
color. Include Harnez's independent VRAM/GTT timelines and Voxi's transcript
feed, active daemon/health content, voice/speed values, history summary,
title/status chrome, visibility hints, bold text, and ANSI-safe ellipsis.

Use deterministic snapshots/events for changing feed and daemon states as well
as metrics. Verify wide/narrow layouts, hidden/restored panes, stable colored
geometry, short histories, and redraws that do not advance sampling. Compare
against the canonical fixtures and document intentional target deviations;
seek visual confirmation when available, with automated geometry/golden checks
as the gate otherwise. This is the complete simulated UI gate before collectors.

## Stage 10 — Real external data sources

Replace simulated producers with small real data-source adapters.

Use Harnez and Voxi as reference implementations for querying Linux metrics,
daemon state, transcript feeds, files, and sockets. Begin with test files and
test sockets filled by a separate process so the source boundary behaves like
the real environment.

The UI must continue to consume snapshots/events rather than knowing how the
data was collected.

## Stage 11 — Declarative data-source and action prototype

Evaluate how far data wiring can be expressed in the UI specification.

Possible first experiments include declaring:

- a file watched at a configured frequency;
- a socket or line-oriented source;
- a mapping from source values to a bar or timeline;
- an event source connected to a UI action handler.

This is a feasibility prototype, not a commitment to a large framework. The
goal is to learn which data-source and action concepts belong in Loom's
specification and which should remain Go application code.

## Long-term outcome

The completed target is a declarative Loom application whose document defines
the title/status chrome, responsive boxes, rows, bars, timelines, colors,
visibility controls, and data bindings. Go defines collectors, external
integration, state transitions, and action behavior. The renderer remains
independent of collection frequency and application-specific coordinates.
