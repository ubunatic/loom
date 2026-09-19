---
title: Loom Roadmap Context
weight: 21
---

# Loom Roadmap Context

This attachment preserves the context behind [`Roadmap.md`](Roadmap.md) for a
future, stronger agent that will derive concrete tickets. It should not need
to rediscover the current repository, target screenshots, or the decisions
already made in this conversation.

## Product vision

Loom is intended to become a Go-based terminal UI library for the age of
agentic software development. UIs should be declarative and simple rather than
built through builder patterns or fluent APIs. The library itself and the UI
declaration should be specced. Go should provide the application behavior and
data sources, while the UI structure, layout, widgets, presentation, and
event/data wiring should increasingly be declarative.

The desired capabilities include:

- flex layout;
- generic boxes with borders, padding, titles, and responsive composition;
- typical widgets without making the whole system application-hardcoded;
- labels, tables, bars, sparklines, timelines, and status/control bars;
- ANSI-aware width and truncation behavior;
- live data whose collection rate is independent of UI redraw rate;
- static show-once and live watch modes;
- deterministic rendering and visual tests.

## Current Loom state

> **Snapshot note (2026-09-19).** This section records the repository as it
> stood when the roadmap was first derived. It is preserved as historical
> framing, not as a current inventory — several gaps listed below have since
> been closed (embedded specs and JSON schemas, the generic `Box`/`Frame`
> responsive layout, the `graph/` renderers, the `collector/` package and its
> independent redraw scheduler, `measure/` and `layout/`, themes including
> truecolor values, and an explicit layered compositor with root-level
> overlays). For the live state of the backlog see
> [`Roadmap.md`](Roadmap.md); the vision, target documents, data/redraw
> decision and sibling-code notes in the rest of this file remain accurate.

Loom is currently a released `v0.1.0` inline TUI widget library extracted from
`uzu`. It has a working terminal pane/runtime, canvas renderer, keyboard and
mouse decoding, colors/styles, screenshots, and widgets including:

- `Choice`
- `Table`
- `Confirm`
- `TextInput`
- `TextArea`
- `Settings`
- `View`
- `Popup`
- `Notif`
- `Stack` and `Grid`
- `Tabs`
- `Router`

The current public runtime boundary is a procedural `Widget` interface with
`Draw`, `HandleKey`, and `HandleMouse`. `Stack` divides space equally and
`Grid` provides uniform cells. The YAML layer builds a limited set of
predefined widgets and routes between named views.

The repository has no implemented `spec/` directory, JSON schemas, embedded
specifications, generic responsive layout, generic box widget, graph widget,
data-binding model, or independent redraw scheduler. `docs/Spec.md` describes
the desired spec system but is currently a reference document rather than an
enforced implementation.

The existing YAML implementation also contains several fields whose behavior
is not yet fully wired, including style and change-handler concepts. This is
important context when deriving tickets: the roadmap is intended to move from
the current narrow YAML factory toward a real declarative document model.

The isolated module checks pass with:

```text
GOWORK=off go test ./...
GOWORK=off go vet ./...
```

The ordinary commands are affected by the parent `/home/uwe/projects/go.work`,
which does not currently include Loom.

## Target documents and assets

The visual targets are recorded in [`HarnezUsageTarget.md`](HarnezUsageTarget.md).
The saved assets are:

- `docs/assets/harnez-usage-wide.png`
- `docs/assets/harnez-usage-slim.png`
- `docs/assets/voxi-monitor-target.png`

The Harnez target has two titled boxes: All Usage and Load. The wide layout
places them side by side; the slim layout stacks them vertically. It includes
aligned labels, determinate usage bars, timestamps/durations, CPU/GPU/RAM
timelines, and a split VRAM/GTT timeline row.

The Voxi target adds:

- a title bar with application name, time, and history summary;
- titled panes with keyboard hints;
- a transcript feed;
- active daemon/health information;
- a bottom status bar with pane controls and quit;
- bold and regular text;
- colored status values;
- ANSI-safe ellipsis truncation;
- one-shot and watch operation.

## Data/redraw decision

Data collection and rendering must remain separate:

```text
collector frequency ──► snapshot/history ──► fixed-rate renderer ──► terminal
```

The redraw loop must not append timeline history merely because it rendered a
frame. A 1 Hz hardware collector and a 20 Hz UI may coexist. A faster source
such as an audio meter must not force hardware histories to advance faster.

Harnez issue 259 already records this invariant and its implementation. Harnez
issue 198 records fixed-width sparklines and preservation of the independent
VRAM/GTT split.

## Reusable sibling code

Harnez contains `internal/rograph`, a dependency-free renderer package with:

- determinate bars;
- partial/eighth-block precision;
- block and Braille sparklines;
- absolute-scale percentage timelines;
- fixed-width/shrink-to-fit behavior;
- configurable ANSI presentation.

Relevant files include:

```text
../harnez/internal/rograph/bar.go
../harnez/internal/rograph/sparkline.go
../harnez/internal/rograph/options.go
../harnez/internal/monitor/... (if present in later work)
```

Voxi has useful monitor rendering behavior in:

```text
../voxi/internal/monitor/render.go
```

That includes visible-width measurement, ANSI-preserving truncation, terminal
width caching, and one-shot/watch monitor behavior. `../pync` was not present
in the workspace when inspected.

The initial graph integration may copy Harnez's graph code into Loom. It should
not immediately force a cross-project extraction. The later architecture can
decide whether the renderer becomes a public shared package.

## Requested first application sequence

The intended order from the conversation is:

1. A tiny show-once app with a title bar, two empty boxes, and a status bar.
2. A basic watch mode, using a ticking clock as the first data source.
3. Responsive layout: boxes side by side when wide and vertically stacked when
   narrow.
4. Controls for hiding and showing boxes.
5. Stable box-size and border testing before adding complex content.
6. Harnez-style tables, aligned labels, bars, and multiple sparklines.
7. Simulated live data feeding the visuals.
8. Color ranges for bars and timelines, including Braille presentation.
9. A complete simulated Harnez/Voxi-style UI.
10. Real files, sockets, and Linux/system data sources.
11. A prototype for expressing data-source wiring and handlers declaratively.

The one apparent ordering correction is intentional: watch mode should be
implemented before visibility buttons, because it establishes the live-state
and redraw contract that later controls must operate within.

## Future ticket-writing guidance

When deriving tickets, keep the first tickets small and demonstrable. Each
stage should leave a runnable example and deterministic tests. Do not add all
future widget types or a generalized dataflow framework before the static
shell, watch loop, responsive layout, and visual geometry contract are proven.

The visual milestone should explicitly include a human confirmation gate when
someone is available. If no human is available, proceed only on the strength
of golden/screenshot output and geometry invariants.
