<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 016 — Complete Harnez and Voxi simulated UI milestone

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 9 — complete simulated milestone, after color and before broader external sources (027 prototype excepted)
**Depends on**: [015](015-simulated-voxi-transcript-and-daemon-panels.md) (transitive prerequisites apply).

## Problem and findings

Completion must demonstrate BOTH reference UIs before moving to broader real-source integration. The narrow typed fixed-rate file prototype in [027](027-introduce-first-spec-driven-collector-prototype.md) was explicitly pulled forward and is the scoped exception. [RoadmapContext](../docs/RoadmapContext.md#requested-first-application-sequence) preserves the original sequence; a live Harnez dashboard alone is not completion.

## Acceptance criteria

- [ ] Publish reproducible commands and a coverage checklist linking each Harnez and Voxi target requirement to an example, fixture and assertion; neither target may be substituted by a generic two-box demo.
- [ ] Harnez covers All Usage/Load, paired bars with percentages/durations, CPU/GPU/RAM, separate VRAM/GTT timelines, short histories and wide/slim layouts.
- [ ] Voxi covers voice/speed, hardware, bounded transcript feed, daemon/health, title/time/history/hidden summary, pane/all-pane/quit hints, bold/regular, status colors and ellipsis.
- [ ] Both run in show-once and simulated watch modes, survive resize/hide/restore, and preserve collection/redraw independence under mismatched rates.
- [ ] Capture a deterministic matrix for wide/slim, startup/mature history, long values, monochrome/color and hidden panels. Human visual confirmation when available; otherwise record unattended golden/ANSI replay evidence, independent position checks, limitations and proceed on passing results.
- [ ] Treat passing this milestone as required before broader external file/socket or real-source work under 017/018, except for the narrow 027 file prototype. Close only with evidence, not because component tickets are closed.

## Verification

Execute both documented examples and the target matrix; run GOWORK=off go vet ./..., GOWORK=off go test ./... and relevant race checks. Record commands, dimensions, simulation seed/time, evidence paths and review mode. Existing ScreenshotScript is ANSI replay, not raster capture.

## Audit — 2026-09-10

Remains Open: the example inventory contains only `examples/monitor`, with
usage/load boxes. There is no second Voxi declaration, transcript/daemon
simulation or combined target matrix; 012–015 still have acceptance gaps.
The sequencing correction above reflects 027's existing scope and shipped
file reader/YAML wiring (`b4d8d02`, `9914fba`); it does not waive either target
or the 017/018 gate. Fresh Loom unit/race, schema, geometry and PTY checks pass,
but no combined-target run is possible with the current implementation.

## Scope limits

Integration and missing bounded simulation glue only; no real collectors or new generalized framework. Reference images are targets, not exact platform-independent pixel goldens.
