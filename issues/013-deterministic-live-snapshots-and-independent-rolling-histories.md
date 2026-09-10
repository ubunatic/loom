<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 013 — Deterministic live snapshots and independent rolling histories

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 7
**Depends on**: [012](012-copy-harnez-rograph-with-verified-provenance-and-bounded-adapters.md) (transitive prerequisites apply).

## Problem and findings

The static graph target must gain deterministic live data without coupling history advancement to Draw. The existing [Widget](../widget.go) contract contains no snapshot/history model; build on the clock separation rather than inventing an existing binding API. [Harnez 259](../../harnez/issues/259-decouple-hardware-load-timeline-sampling-from-high-fps-tui-redraw-cadence.md) records the same invariant.

## Acceptance criteria

- [ ] Use seeded/scripted simulated producers and immutable or safely published snapshots; Go owns bounded history storage and only producer samples append to it.
- [ ] Run independent slow hardware and faster meter sources at configurable rates while rendering at a separate fixed cadence. Changing render rate does not change sample timestamps/counts.
- [ ] Keep CPU, GPU, RAM, VRAM and GTT histories independent, including separate VRAM/GTT data. Padding during startup does not fabricate stored measurements.
- [ ] Define startup, retention, latest-value and shutdown behavior; repeated rendering of the same snapshot is byte/geometry stable and does not mutate histories.
- [ ] Provide a runnable simulated Harnez watch example and deterministic one-shot snapshot path. Full Voxi parity remains explicitly gated by the later combined milestone.

## Verification

Use fake time and seeded sequences for empty/short/full/rolling histories; assert sample counts at 1 Hz and a faster rate while drawing at 20 Hz, plus rapid repeated draws and producer shutdown. Run race tests, isolated vet/tests and wide/slim final-output geometry checks.

## Audit — 2026-09-10

Remains Open. `b376e0a` supplies deterministic sampling, copied snapshots and
bounded per-series slices in [state.go](../examples/monitor/state.go).
`TestMonitorStateSamplesIndependentlyFromSnapshot` verifies CPU retention and
copy isolation. `TestCadenceIndependentTicks` in [cadence_test.go](../cadence_test.go)
proves 20 redraw ticks do not advance one producer's count/timestamp and checks
cancellation. These are partial acceptance evidence, not a multi-source test.

`monitorState.Sample` still advances CPU/RAM/GPU/VRAM/GTT together, and
`monitorSnapshot` stores no sample timestamps. There is no independently
scheduled faster simulated meter or slow/fast producer matrix with both faster
and slower redraw. Keep the remaining startup/history/render-stability and
joined producer-shutdown matrix open. The watch path also starts the real
027 file source; it is not yet an isolated simulated-source demonstration.
Fresh `GOWORK=off make test`, `GOWORK=off go test -race ./...` and
`GOWORK=off make watch-pty` pass without establishing those missing contracts.

## Scope limits

No real I/O, unbounded history, source framework, persistence or claim that Harnez alone completes the two-target milestone.
