---
title: Loom Roadmap
weight: 20
---

# Loom Roadmap

This is an aspirational product roadmap for turning Loom into a simple,
spec-driven terminal UI system. The original stages and their follow-ups are
tracked in the [issue index](../issues/README.md). Stage numbers preserve the
original plan; the Now/Next/Later sequence below reflects current priorities.

Loom's value is reusable, dependency-light inline terminal UI: applications
such as the Harnez/Voxi reference dashboards, should share widgets
without taking over the full terminal. An application should declare its UI
structure and presentation, while Go supplies behavior and changing data. The
UI should be easy to maintain, responsive to terminal size, testable without a live
terminal, and free from application-specific coordinate/layout code.

The primary reference targets are documented in
[`HarnezUsageTarget.md`](HarnezUsageTarget.md) (usage dashboard and Voxi monitor)
and [`HarnezSplashTarget.md`](HarnezSplashTarget.md) (startup loading and splash
screen).

## Assessment against code and backlog

Reconciled 2026-09-11 against the open tickets returned by
`harnez find -d . issues status:open` and current code/tests.
Issue 004 is closed as invalid because uzu is deprecated and was never an active
downstream consumer.

- The monitor has embedded YAML and consumed schemas, title/status chrome,
  bounded child rendering, breakpoint stacking, visibility actions, independent
  collection/redraw scheduling, and deterministic geometry checks.
- [`graph/`](../graph/) contains the adapted Harnez renderers and verified
  provenance. Numerical snapshots drive paired bars, CPU/RAM/GPU timelines,
  and separate VRAM/GTT timelines. [`state.go`](../examples/monitor/state.go)
  supplies copied, bounded simulated histories with independent slow hardware
  and fast meter sampling cadences and timestamps (012, 013 closed).
- [`collector/`](../collector/) provides `Collector`, explicit `type: file`,
  bounded whole-file reads, immediate/periodic sampling via `Run`, duration-based
  retention, and safe history publication. The monitor loads its embedded
  [`watch.yaml`](../examples/monitor/spec/watch.yaml) and runs a real
  `/proc/stat` source at 1 Hz on worker goroutines, parsing CPU percentage into
  live display snapshots with correct simulated vs real provenance footers (027 closed).
- [`measure/`](../measure/) and [`layout/`](../layout/) export renderer-independent
  cell widths, clusters, multiline bounds, left/right truncation, and min/pref/max
  space allocation for responsive and dynamic boxes (028 closed).

### Shipped and partial progress

- Release/API foundations: 001–003 done; 005 resolved (AGPL retained).
- Stages 1–5: 006–010 closed, covering shell, schemas, watch, responsive
  placement, controls and the geometry gate.
- Stage 6: rows/review 011 and 020 closed; graph primitives 024 and numerical
  monitor integration 025 closed; parent 012 closed with verified provenance,
  paired usage bars, and tiny-terminal graph rendering tests.
- Test/schema hygiene: 021–023 closed. Newly reconciled 026 is closed:
  `make watch-pty` builds the monitor and supplies `--watch`, fixing the smoke
  invocation rather than changing the scheduler.
- Stage 7: deterministic live snapshots, independent slow/fast producer sampling, and decoupled redraw timing 013 closed.
- Typed fixed-rate file prototype: 027 closed with validated collector spec, bounded file reader, lifecycle and changing-fixture tests, `/proc/stat` parsing, and live snapshot feeding.
- Reusable measurement and dynamic layout: 028 closed with shared cell policy, line bounds, and min/preferred/max dynamic layout.

### Verified backlog gaps

Every open ticket was read against the implemented surface and tests. This
table records why none can close; detailed audit notes accompany 004, 016,
and 028. Unfinished acceptance checkboxes remain unchecked.

| Issue | Evidence and remaining acceptance |
|---|---|
| 004 | Closed as invalid: uzu is deprecated and was never an active downstream consumer. |
| 014 | `graph/options.go` offers Go colors/glyphs; monitor schema has no declared palettes/ranges/glyph contract or two-palette target matrix. |
| 015 | Only the usage/load monitor example exists; no Voxi declaration, bounded transcript events or daemon-state simulation. |
| 016 | Both target implementations and the combined colored/show-once/watch matrix are incomplete. Only 027 is excepted from its broader-source gate. |
| 017 | The raw file reader has unit fixtures, but no separate-process file/socket producers, reconnect/framing or partial-write lifecycle evidence. |
| 018 | `/proc/stat` is read as bytes; no metric parsing/units or daemon-source probe/adapter and no evidence-backed go/no-go report. |
| 019 | Typed file/cadence YAML is consumed; socket/value mapping, named event handlers, comparison with Go and adopt/narrow/reject decision are absent. |

### Close / Park

- [004 — uzu migration](../issues/004-migrate-uzu-to-shared-loom.md): closed as
  invalid because uzu is deprecated and was never an active downstream
  consumer.
- [019 — wider declarative wiring](../issues/019-evaluate-declarative-source-and-action-wiring.md):
  park its socket/event/action and comparative feasibility work until 017/018
  supply proven boundaries. Its narrow file/cadence declaration is already
  being proven by 027, so avoid a second competing syntax; keep 019 open for
  mappings, named handlers and the adopt/narrow/reject decision.
- Keep cross-project package extraction and generalized dataflow parked.
  Reusing code inside Loom is supported by the sibling findings; it does not
  justify a shared multi-repository framework or a claimed savings percentage.

### Ticket sequencing

The immediate sequence is **014 configurable colors/glyphs → 015 Voxi panels → 016 complete simulated target milestone**, alongside the Harnez splash sequence (**029 centered layout → 030 braille spinner/bar → 031 status pills → 032 lifecycle coordinator → 033 integration**).
012, 013, 027, and 028 are shipped and provide stable graph, timing, measurement, and collector foundations.

**Now — expressive target UIs & splash startup.** Continue
[014](../issues/014-configurable-graph-colors-and-glyph-presentation.md) →
[015](../issues/015-simulated-voxi-transcript-and-daemon-panels.md) →
[016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md): declared
colors/glyphs, bounded transcript and daemon panels, then the complete two-target
matrix. In parallel or following these, implement the Harnez splash screen target
sequence ([029](../issues/029-declarative-centered-layout-and-viewport-alignment-primitives.md) →
[030](../issues/030-braille-activity-spinner-and-bracketed-progress-bar-primitives.md) →
[031](../issues/031-provider-status-pill-cluster-and-lifecycle-state-presentation.md) →
[032](../issues/032-splash-lifecycle-controller-async-provider-coordination-and-key-dismissal.md) →
[033](../issues/033-harnez-target-splash-screen-integration-and-golden-tests.md)) to prove
centered viewport layouts, braille spinner/bar animations, and view transition lifecycles.
028 now precedes these because shared measurement reduces sizing and ANSI drift
as richer content arrives; 013 remains 014's declared prerequisite.

**Later — broader external boundaries and binding decisions.** Retain
[017](../issues/017-external-file-and-socket-adapters-with-separate-producer-fixtures.md)
→ [018](../issues/018-explore-bounded-linux-and-daemon-source-adapters.md):
separate-process file/socket fixtures and lifecycle canaries, then bounded
Linux/daemon probes. Reuse 027's file collector, adding replacement/partial-write,
disconnect/reconnect and stale-data evidence rather than rebuilding it. Real
CPU percentage parsing, GPU paths/units/availability, and daemon/transcript
formats still need source-specific evidence under 018. Revisit parked 019 only
after these boundaries justify wider declarations. Deferral preserves attention
for reusable UI value while host-specific integration remains uncertain.

Each stage should leave a runnable example and deterministic checks. Keep examples in the root module so root test discovery includes them; the target document's nested `go.mod` is illustrative, not a requirement.

| Stage | Tickets |
|---|---|
| 1–5 — Shipped Foundation | [006](../issues/006-static-declarative-monitor-shell-with-a-minimal-validated-contract.md), [007](../issues/007-clock-watch-mode-with-independent-collection-and-redraw.md), [008](../issues/008-responsive-declared-box-layout.md), [009](../issues/009-declarative-box-visibility-controls.md), [010](../issues/010-geometry-and-visual-evidence-milestone-before-rich-content.md) |
| Hygiene & Fixes (Shipped) | [021](../issues/021-cmd-help-tty-blocking-in-tests.md), [022](../issues/022-decouple-static-shell-golden-from-evolving-monitor-spec.md), [023](../issues/023-rows-schema-validation-and-negative-controls.md), [026](../issues/026-investigate-monitor-pty-smoke-test-idle-redraw-regression.md) |
| 6 — Rows & Graphs | [011](../issues/011-aligned-dashboard-rows-and-ansi-safe-truncation.md) (done), [020](../issues/020-review-ticket-011-rows-and-target-state-alignment.md) (done), [024](../issues/024-port-harnez-rograph-primitives-with-provenance.md) (done), [025](../issues/025-integrate-graph-renderers-into-declarative-monitor.md) (done), [012](../issues/012-copy-harnez-rograph-with-verified-provenance-and-bounded-adapters.md) (done) |
| 7 — Simulated live data | [013](../issues/013-deterministic-live-snapshots-and-independent-rolling-histories.md) (done) |
| Reusable measurement and dynamic layout (Shipped) | [028](../issues/028-add-reusable-measurement-and-dynamic-box-layout-primitives.md) (closed; bounded measurement, planner and dynamic Box/Stack integration) |
| Typed fixed-rate file prototype | [027](../issues/027-introduce-first-spec-driven-collector-prototype.md) (done) |
| 8 — Color and glyphs | [014](../issues/014-configurable-graph-colors-and-glyph-presentation.md) |
| 9 — Complete simulated targets | [015](../issues/015-simulated-voxi-transcript-and-daemon-panels.md), [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md) |
| 10 — Splash & startup screen | [029](../issues/029-declarative-centered-layout-and-viewport-alignment-primitives.md), [030](../issues/030-braille-activity-spinner-and-bracketed-progress-bar-primitives.md), [031](../issues/031-provider-status-pill-cluster-and-lifecycle-state-presentation.md), [032](../issues/032-splash-lifecycle-controller-async-provider-coordination-and-key-dismissal.md), [033](../issues/033-harnez-target-splash-screen-integration-and-golden-tests.md) |
| 11 — External sources | [017](../issues/017-external-file-and-socket-adapters-with-separate-producer-fixtures.md), [018](../issues/018-explore-bounded-linux-and-daemon-source-adapters.md) |
| 12 — Wider declarative source/action evaluation | [019](../issues/019-evaluate-declarative-source-and-action-wiring.md) (parked beyond 027's file prototype) |

### Verification gates for the next increments

The shipped 028 measurement/layout foundation is covered by known-column
canaries, content-height, min/preferred/max allocation, fixed sizing, visibility,
wrapping, tiny-terminal and ANSI geometry tests.
013/027 require deterministic sample-count/timestamp matrices with slower and
faster redraw, changing file fixtures, stale/error cases, bounded retention and
joined shutdown. Retain `make test`, relevant race checks and `make watch-pty`
for implementation delivery.

This audit freshly passed `GOWORK=off make test` (schema negative controls,
64x9/40x18 saved geometry replay, vet and all Go packages),
`GOWORK=off go test -race ./...`, `GOWORK=off make watch-pty` (redraw,
resize, toggles and terminal restoration), and the no-controlling-TTY canary.
`harnez status` reports all 28 tracker entries consistent. `harnez index` and
`harnez index --check` pass after converting the existing studies list in
[`docs/README.md`](README.md) to the required anchored table. The issue index
was regenerated for the 028 closure. Existing checks validate the bounded
implementation and its acceptance matrix;
visual verification was automated, with no human screenshot review.

The stage descriptions below retain the original scope as historical acceptance
context. Their order is superseded by the sequence above where explicitly noted.

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

At this initial slice there was no `--watch` mode, live data, graph, or button
behavior. The goal is to establish the application frame, generic box rendering, basic spacing,
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

Shipped in 008 as fixed preferred boxes with breakpoint stacking. Content-driven
measurement and min/preferred/max allocation are the new 028 follow-up.

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

Partially implemented in 013; the independent producer-rate matrix remains open.

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
as the gate otherwise. This remains the complete simulated UI gate before
broader external adapters, with the scoped 027 file prototype exception above.

## Stage 10 — Splash & startup loading target

Demonstrate the Harnez splash and startup loading screen target
([`HarnezSplashTarget.md`](HarnezSplashTarget.md)) in the example app suite.

- Implement declarative vertical and horizontal viewport centering (029).
- Add braille animation spinners and bracketed dot-matrix progress bars (030).
- Render multi-state provider status pill clusters (`● mic`, `✳ claude`, `֍ codex`, `Λ agy`) with stable spacing and ANSI styling (031).
- Orchestrate asynchronous provider startup tasks, step messaging (`fetching claude...` → `agy done`), and interactive `Esc` skip (032).
- Assemble the integrated example with deterministic golden frame tests (033).

This proves centered viewport composition, interactive early exit, and
smooth transition lifecycles from startup screens into main dashboard views.

## Stage 11 — Real external data sources

027 has pulled forward raw fixed-rate file reads and live retention. The
separate-process file/socket and source-specific work in 017/018 remains later.

Replace simulated producers with small real data-source adapters.

Use Harnez and Voxi as reference implementations for querying Linux metrics,
daemon state, transcript feeds, files, and sockets. Begin with test files and
test sockets filled by a separate process so the source boundary behaves like
the real environment.

The UI must continue to consume snapshots/events rather than knowing how the
data was collected.

## Stage 12 — Declarative data-source and action prototype

027 already consumes embedded typed file/cadence declarations. The remaining
019 evaluation covers wider source mappings and named actions after 017/018;
it must build on that seam rather than introduce a second file declaration.

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
