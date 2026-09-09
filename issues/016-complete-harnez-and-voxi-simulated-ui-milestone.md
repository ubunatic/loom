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
**Roadmap stage**: 9 — complete simulated milestone, after color and before external sources
**Depends on**: [015](015-simulated-voxi-transcript-and-daemon-panels.md) (transitive prerequisites apply).

## Problem and findings

Completion must demonstrate BOTH reference UIs before moving to real sources. [RoadmapContext](../docs/RoadmapContext.md#requested-first-application-sequence) and roadmap stage 9 place this after colors; a live Harnez dashboard alone is not completion.

## Acceptance criteria

- [ ] Publish reproducible commands and a coverage checklist linking each Harnez and Voxi target requirement to an example, fixture and assertion; neither target may be substituted by a generic two-box demo.
- [ ] Harnez covers All Usage/Load, paired bars with percentages/durations, CPU/GPU/RAM, separate VRAM/GTT timelines, short histories and wide/slim layouts.
- [ ] Voxi covers voice/speed, hardware, bounded transcript feed, daemon/health, title/time/history/hidden summary, pane/all-pane/quit hints, bold/regular, status colors and ellipsis.
- [ ] Both run in show-once and simulated watch modes, survive resize/hide/restore, and preserve collection/redraw independence under mismatched rates.
- [ ] Capture a deterministic matrix for wide/slim, startup/mature history, long values, monochrome/color and hidden panels. Human visual confirmation when available; otherwise record unattended golden/ANSI replay evidence, independent position checks, limitations and proceed on passing results.
- [ ] Treat passing this milestone as required before external file/socket or real-source work. Close only with evidence, not because component tickets are closed.

## Verification

Execute both documented examples and the target matrix; run GOWORK=off go vet ./..., GOWORK=off go test ./... and relevant race checks. Record commands, dimensions, simulation seed/time, evidence paths and review mode. Existing ScreenshotScript is ANSI replay, not raster capture.

## Scope limits

Integration and missing bounded simulation glue only; no real collectors or new generalized framework. Reference images are targets, not exact platform-independent pixel goldens.
