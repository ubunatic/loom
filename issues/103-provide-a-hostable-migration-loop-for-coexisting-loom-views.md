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

### M1 delivered — standalone Loom rebuild

Added `examples/usage`: a hostable `Frame` with colored All Usage and Load boxes,
deterministic quota samples, and asynchronous local CPU/memory sampling from
procfs. The source is injectable, the pane redraw cadence is independent, and
the example is registered for `loom-demo`/`loom-bench`. Tests cover initial
collection, expected content, narrow and resized layouts, and silent hosted
Draw/Tick calls.

Verification: `make test-q1` passed; output is
`/tmp/loom-103-m1-test-q1.log` (no `--- FAIL` lines).

### M2 delivered — async repaint and terminal evidence

Successful background collection now invokes the Pane invalidation callback.
The test keeps the redraw interval at one hour and proves collection completion
repaints without waiting for a tick. A PTY test reads Loom's rendered color
cells and checks the All Usage title RGB. The PTY snapshot is
`/tmp/loom-103-usage.ansi`.

Verification: `make test-q1` passed on rerun after changing the PTY test to wait
for stable title/content rather than the transient collecting message. Output:
`/tmp/loom-103-m2-test-q1.log` (no `--- FAIL` lines). A direct PTY capture of
the built example also confirmed the RGB SGR and produced the reported snapshot.

### M3 delivered — selectable plain and Loom views

`--view=loom|plain` selects the initial view; `v` switches views while the
example is running. Both views render the same collected snapshot. The plain
view uses Loom `StyledRows` for ANSI styling and clipping, with the panels
side-by-side at wide sizes and stacked at narrow sizes. The example README and
help text document the switch.

Verification: `make test-q1` passed; output is
`/tmp/loom-103-m3-test-q1.log` (no `--- FAIL` lines). The PTY color-cell
snapshot remains at `/tmp/loom-103-usage.ansi`.

### Reference behavior for M1–M3

Source: `../harnez/internal/usage/watch.go`, `load.go`, `indicatorsspec.go`,
`../harnez/spec/indicators.yaml`, and `colors.yaml` (read-only; reference only).

- **Compact arrangement:** `compactWatchSections` enables All Usage and Load
  (as well as Tokens and Remote Load in the full Harnez view); this Loom example
  focuses on the requested local All Usage and Load pair (`watch.go:411-413`).
- **All Usage contents:** one row per model group when present, otherwise the
  agent row uses its weekly and session quota windows. Each window shows used
  percentage and a short time-to-reset; two windows show two bars/percentages,
  one window leaves the second bar slot blank. Agents with no usage data are
  skipped, while stale values are dimmed and collectors with status but no
  windows can be marked (`watch.go:733-840, 862-930`). Empty summary text is
  “no quota windows available” (`watch.go:692-705`).
- **All Usage colors:** heat mode colors bars and percentages cool blue for
  0–25%, green above 25–50%, yellow above 50–75%, and warm red above 75–100%;
  the boundaries stay in the lower band. These colors are SGR 34, 32, 33, and
  31 respectively (`indicatorsspec.go:561-566, 584-625`; `spec/indicators.yaml:103-116`;
  `spec/colors.yaml:19-34`).
- **Load contents:** local CPU row reports core count, recent utilization trend,
  current utilization percent (or `n/a` before a valid delta), and temperature
  when available (`watch.go:1112-1147`). RAM reports used/total GiB and percent
  (`watch.go:1167-1180`). Up to two GPUs report utilization and optional
  temperature; available VRAM/GTT is shown as used/total GiB and percent
  (`watch.go:1080-1095, 1183-1262`). Unavailable GPU data is `gpu n/a`; an
  otherwise empty local snapshot is `load data unavailable`.
- **Load data and chart colors:** the local CPU reader gathers core count,
  delta-based utilization, CPU temperature, and memory, while its result also
  carries load averages (`load.go:16-50, 62-82, 97-106`). CPU/RAM/GPU chart presentation defaults to Braille time
  series; heat-colored percentage bands share the thresholds and palette above
  (`spec/indicators.yaml:93-116, 123-127`).
- **Panel treatment:** both boxes use titled sharp-corner borders; Harnez's
  compact renderer draws titled box outlines and clips/pads inner rows
  (`watch.go:197-229, 692-705, 1017-1023`). Loom should express this through
  Loom widgets and styles, not by carrying over that renderer.

**Delivered:** reference behavior (a537019). **M1 (8d78564):** `examples/usage` with All Usage (deterministic fake quotas) and Load (live CPU and RAM) boxes on a Loom `Frame`, using `StyledRows` and heat colors 34/32/33/31 at the Harnez thresholds; hostable and registered. **M2 (be9f193):** asynchronous collection with repaint through `InvalidationAware`. **M3 (c2c90d3):** a flag and a key switch between the plain rebuilt view and the Loom view, sharing one model. Host review: the snapshot `/tmp/loom-103-usage.ansi` shows correct heat bands (27% green, 38% green, 64% yellow, 81% red), titled boxes and the footer hint.
**Simplified vs Harnez (acceptable for a minimal port):** one quota window per agent, and no CPU temperature, GPU/VRAM or Braille trend. **Visual check** is parked in 102. **Harnez-side integration** is tracked in the harnez tracker.
