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
[`HarnezUsageTarget.md`](HarnezUsageTarget.md), including the Harnez usage
dashboard and the Voxi monitor.

## Assessment against code and backlog

Reconciled 2026-09-10 against the open tickets returned by
`harnez find -d . issues status:open`, code/tests at `6b80ced`, and recent commits.
Issue 004 is closed as invalid because uzu is deprecated and was never an active
downstream consumer; no other complete issue was verified.
The initial 2026-09-09 assessment and sequence in
[`RoadmapContext.md`](RoadmapContext.md) remain historical context; statements
there and in the root README about missing schemas/watch support predate the
implementation below.

- The monitor now has embedded YAML and consumed schemas, title/status chrome,
  bounded child rendering, breakpoint stacking, visibility actions, independent
  collection/redraw scheduling, and deterministic geometry checks. Existing
  `Stack`/`Grid` behavior and legacy YAML fields do not establish a general
  content-measurement or data-binding API.
- [`graph/`](../graph/) contains the adapted Harnez renderers and provenance.
  Numerical snapshots drive bars, CPU/RAM/GPU timelines, and separate VRAM/GTT
  timelines. [`state.go`](../examples/monitor/state.go) supplies copied,
  bounded simulated histories; all simulated series currently advance together.
  Independent slow/fast source rates and timestamp/count coverage still belong
  to 013.
- [`collector/`](../collector/) now provides `Collector`, explicit `type: file`,
  bounded whole-file reads, immediate/periodic sampling, and in-memory history
  with a default 15-minute retention window. The monitor loads its embedded
  [`watch.yaml`](../examples/monitor/spec/watch.yaml) and starts a real
  `/proc/stat` source at 1 Hz on worker goroutines, separate from redraw.
  This loads YAML embedded in the binary; it does not add runtime config-file
  discovery or reload.
- Collector records are still raw bytes retained by the worker; they do not
  feed displayed metrics. Show-once uses a static numerical snapshot; watch
  graphs remain simulated. Source errors currently end watch, and retention
  trims relative to appended timestamps. `316159a` shipped synchronized
  `History.Append`/`Snapshot` with copied byte slices and concurrent-access
  tests. Parsed numeric snapshots, wiring those snapshots to the UI,
  stale/error presentation, and lifecycle/rate tests remain acceptance work
  for 027. No disk persistence or GPU collector shipped. The Load footer's
  `(real collector data)` currently mislabels simulated graphs, including
  show-once output where no source runs; correct it in 027's display proof.
- Show-once uses terminal width up to the declared 80-column cap, with an
  80-column fallback for redirected output. Plain provenance footers and a
  spacer now appear inside boxes. These improvements still rely on fixed
  declared box dimensions: [`frame.go`](../frame.go) clips preferred rectangles
  and switches at a breakpoint; it does not measure content or distribute
  min/preferred/max sizes. The footer is a special box field, not a general
  measured vertical stack.
- 028's first measurement slice shipped in `fc1bbeb` and `3ed92c7`:
  [`measure/`](../measure/) exports renderer-independent cell widths, clusters,
  multiline bounds and left/right truncation; Canvas delegates its width policy.
  ANSI is stripped and emoji-ZWJ clusters are explicitly unsupported. Exact
  fit/padding, root truncation deduplication, constrained content sizing and
  dynamic allocation remain open. Its sibling findings identify Harnez's `internal/uix` min/pref/max,
  wrap/stretch planner and usage measurement pass, plus Voxi's visible-width,
  ANSI-preserving truncation and box sizing. Consolidate Loom's existing width
  helpers against an explicit cell policy; rune-counting sibling code is not
  a Unicode oracle. Keep domain assembly outside the library and record reuse
  provenance without changing either sibling.

### Shipped and partial progress

- Release/API foundations: 001–003 done; 005 resolved (AGPL retained).
- Stages 1–5: 006–010 closed, covering shell, schemas, watch, responsive
  placement, controls and the geometry gate.
- Stage 6: rows/review 011 and 020 closed; graph primitives 024 and numerical
  monitor integration 025 closed. Parent 012 remains incomplete: the second
  usage bar is a literal placeholder, provenance omits the verified metadata
  gap, and full paired-graph target geometry is not established.
- Test/schema hygiene: 021–023 closed. Newly reconciled 026 is closed:
  `make watch-pty` builds the monitor and supplies `--watch`, fixing the smoke
  invocation rather than changing the scheduler.
- Since the previous roadmap pass: 027's typed file reader, duration retention
  and YAML-to-watch wiring landed (`b4d8d02`, `8e87868`, `9914fba`); terminal
  sizing, the 80-column limit, and plain footers/spacer landed through `2c2b782`.
  These are shipped slices of open work, not closure of 027 or 028.
- 013's first deterministic sampling slice (`b376e0a`) remains partial.
  Since the previous roadmap pass, 028's shared measurement policy and 027's
  synchronized history publication shipped. Neither completes its ticket.

### Verified backlog gaps

Every open ticket was read against the implemented surface and tests. This
table records why none can close; detailed audit notes accompany 004, 012,
013, 016, 027 and 028. Unfinished acceptance checkboxes remain unchecked.

| Issue | Evidence and remaining acceptance |
|---|---|
| 004 | Closed as invalid: uzu is deprecated and was never an active downstream consumer. |
| 012 | `applySnapshot` updates one usage bar; the second is literal YAML. Graph port/tests exist, but paired target coverage and accurate source-license/authorization evidence remain. |
| 013 | `monitorState.Sample` advances all series together without timestamps. One-producer cadence tests do not prove independent slow/fast simulations. |
| 014 | `graph/options.go` offers Go colors/glyphs; monitor schema has no declared palettes/ranges/glyph contract or two-palette target matrix. |
| 015 | Only the usage/load monitor example exists; no Voxi declaration, bounded transcript events or daemon-state simulation. |
| 016 | Both target implementations and the combined colored/show-once/watch matrix are incomplete. Only 027 is excepted from its broader-source gate. |
| 017 | The raw file reader has unit fixtures, but no separate-process file/socket producers, reconnect/framing or partial-write lifecycle evidence. |
| 018 | `/proc/stat` is read as bytes; no metric parsing/units or daemon-source probe/adapter and no evidence-backed go/no-go report. |
| 019 | Typed file/cadence YAML is consumed; socket/value mapping, named event handlers, comparison with Go and adopt/narrow/reject decision are absent. |
| 027 | Raw history publication is synchronized; numeric display integration, error/stale state and rate/in-flight cancellation tests remain. |
| 028 | Shared cell policy and line bounds exist; fit/padding, constrained content sizing, dynamic layout and full policy/migration evidence remain. |

### Close / Park

- [004 — uzu migration](../issues/004-migrate-uzu-to-shared-loom.md): closed as
  invalid because uzu is deprecated and was never an active downstream
  consumer.
- [012 — graph parent](../issues/012-copy-harnez-rograph-with-verified-provenance-and-bounded-adapters.md)
  is no longer a closure candidate: the audit found concrete paired-bar,
  provenance and target-geometry work. Move that bounded remainder into Now
  alongside 028, before dependent 013 acceptance; retain the existing graph port.
- [019 — wider declarative wiring](../issues/019-evaluate-declarative-source-and-action-wiring.md):
  park its socket/event/action and comparative feasibility work until 017/018
  supply proven boundaries. Its narrow file/cadence declaration is already
  being proven by 027, so avoid a second competing syntax; keep 019 open for
  mappings, named handlers and the adopt/narrow/reject decision.
- Keep cross-project package extraction and generalized dataflow parked.
  Reusing code inside Loom is supported by the sibling findings; it does not
  justify a shared multi-repository framework or a claimed savings percentage.

### Ticket sequencing

The immediate sequence is **finish 028 measurement → 028 dynamic layout → remaining
013 timing contract → 027 displayed file-data proof**, with 012's verified
acceptance gaps resolved alongside 028. 028 does not depend on unfinished collectors;
013's timing evidence supports finishing 027. These are bounded increments,
not a requirement to build a general layout or source framework first.

**Now — reusable sizing and trustworthy live data.** Retain
[028](../issues/028-add-reusable-measurement-and-dynamic-box-layout-primitives.md)
because the footer/spacer work exposed repeated manual dimension tuning.
Finish fit/padding, truncation consolidation and constrained content sizing
on the shipped `measure` policy, then add a pure
deterministic allocator and Frame/Box adapters. Migrate the monitor through
validated declarations, including rows, plain text, spacer, border and padding
in its measured height; preserve fixed-size compatibility and graph widths.
This directly improves every application composing Loom widgets.

Keep [013](../issues/013-deterministic-live-snapshots-and-independent-rolling-histories.md)
in Now: finish independent slow/fast simulated producers, bounded histories,
timestamps, shutdown and mismatched redraw-rate tests. Complete
[027](../issues/027-introduce-first-spec-driven-collector-prototype.md) with a
changing numeric file fixture feeding safely published display snapshots,
explicit parsing/error/stale semantics, and retention/cancellation evidence.
Document raw whole-file polling as the initial source contract; no external
event notification is required. Check provenance text against actual displayed
values so the real-source label cannot imply that simulated graphs are real.
Preserve the 15-minute live-only default and keep parsing outside Draw.

**Next — expressive, proven target UIs.** Continue
[014](../issues/014-configurable-graph-colors-and-glyph-presentation.md) →
[015](../issues/015-simulated-voxi-transcript-and-daemon-panels.md) →
[016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md): declared
colors/glyphs, bounded transcript and daemon panels, then the complete two-target
matrix. This sequence remains valuable: it proves reusable composition across
metrics, text and status instead of equating a two-box demo with both targets.
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

**Ordering change:** the user explicitly pulled the narrow fixed-rate file
prototype forward via 027, ahead of the original 016 → 017 → 018 → 019 gate.
That is the scoped exception now reflected here; 016 remains the complete
simulated-target milestone and the gate for broader file/socket/system work.
016 now records this exception explicitly; the context document retains the
historical original ordering. This audit changes no issue status or priority.
012 moves from speculative closure to concrete Now work because its acceptance
audit found unfinished requirements; the other buckets retain their value order.

Each stage should leave a runnable example and deterministic checks. Keep examples in the root module so root test discovery includes them; the target document's nested `go.mod` is illustrative, not a requirement.

| Stage | Tickets |
|---|---|
| 1–5 — Shipped Foundation | [006](../issues/006-static-declarative-monitor-shell-with-a-minimal-validated-contract.md), [007](../issues/007-clock-watch-mode-with-independent-collection-and-redraw.md), [008](../issues/008-responsive-declared-box-layout.md), [009](../issues/009-declarative-box-visibility-controls.md), [010](../issues/010-geometry-and-visual-evidence-milestone-before-rich-content.md) |
| Hygiene & Fixes (Shipped) | [021](../issues/021-cmd-help-tty-blocking-in-tests.md), [022](../issues/022-decouple-static-shell-golden-from-evolving-monitor-spec.md), [023](../issues/023-rows-schema-validation-and-negative-controls.md), [026](../issues/026-investigate-monitor-pty-smoke-test-idle-redraw-regression.md) |
| 6 — Rows & Graphs | [011](../issues/011-aligned-dashboard-rows-and-ansi-safe-truncation.md) (done), [020](../issues/020-review-ticket-011-rows-and-target-state-alignment.md) (done), [024](../issues/024-port-harnez-rograph-primitives-with-provenance.md) (done), [025](../issues/025-integrate-graph-renderers-into-declarative-monitor.md) (done); parent [012](../issues/012-copy-harnez-rograph-with-verified-provenance-and-bounded-adapters.md) (now; acceptance gaps) |
| 7 — Simulated live data | [013](../issues/013-deterministic-live-snapshots-and-independent-rolling-histories.md) (now) |
| Reusable measurement and dynamic layout | [028](../issues/028-add-reusable-measurement-and-dynamic-box-layout-primitives.md) (now; extends shipped Stage 3) |
| Typed fixed-rate file prototype | [027](../issues/027-introduce-first-spec-driven-collector-prototype.md) (now; partial slice pulled forward from Stages 10–11) |
| 8 — Color and glyphs | [014](../issues/014-configurable-graph-colors-and-glyph-presentation.md) |
| 9 — Complete simulated targets | [015](../issues/015-simulated-voxi-transcript-and-daemon-panels.md), [016](../issues/016-complete-harnez-and-voxi-simulated-ui-milestone.md) |
| 10 — External sources | [017](../issues/017-external-file-and-socket-adapters-with-separate-producer-fixtures.md), [018](../issues/018-explore-bounded-linux-and-daemon-source-adapters.md) |
| 11 — Wider declarative source/action evaluation | [019](../issues/019-evaluate-declarative-source-and-action-wiring.md) (parked beyond 027's file prototype) |

### Verification gates for the next increments

028 extends existing known-column canaries with measurement/fit tests, then tests
content height, min/preferred/max allocation, fixed sizing, visibility, wrapping,
tiny terminals and ANSI geometry independently of its own measurement helper.
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
needed no row changes because all statuses and paths remain unchanged. Existing checks
validate implemented slices, not the absent acceptance matrices above;
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

## Stage 10 — Real external data sources

027 has pulled forward raw fixed-rate file reads and live retention. The
separate-process file/socket and source-specific work in 017/018 remains later.

Replace simulated producers with small real data-source adapters.

Use Harnez and Voxi as reference implementations for querying Linux metrics,
daemon state, transcript feeds, files, and sockets. Begin with test files and
test sockets filled by a separate process so the source boundary behaves like
the real environment.

The UI must continue to consume snapshots/events rather than knowing how the
data was collected.

## Stage 11 — Declarative data-source and action prototype

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
