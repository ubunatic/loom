<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 015 — Simulated Voxi transcript and daemon panels

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 9 — Voxi target preparation
**Depends on**: [014](014-configurable-graph-colors-and-glyph-presentation.md) (transitive prerequisites apply).

## Problem and findings

A Harnez-only dashboard misses the Voxi target. The [Voxi target](../docs/HarnezUsageTarget.md#voxi-monitor-target) adds transcript and daemon/health content beyond Harnez graphs. Existing [TextArea](../textarea.go) is an editor, not an established transcript-feed API.

## Acceptance criteria

- [ ] Provide a second simulated declaration using shared chrome/layout/row/graph primitives for voice & speed, hardware load, transcript feed and active daemons & health.
- [ ] Use deterministic fake transcript events and daemon state transitions. Bound retained transcript data and define visible tail, empty state, overflow and truncation behavior.
- [ ] Include application/time/history summary and hidden-pane hint in title chrome; show active pane controls and quit in status. Declare pane hints and all-panels action; Go wires explicit handlers.
- [ ] Show bold/regular labels, colored health/status, speed/meter data and independent producer rates. Reuse existing watch/snapshot behavior rather than a Voxi-specific render loop.
- [ ] Keep transcript and daemon domain names in example data/declarations; document generic primitives added only where existing View/TextArea/rows cannot serve the target.

## Verification

Replay deterministic transcript arrival, health transitions, long lines and all visibility combinations at wide/slim widths. Assert bounded memory and no history advancement on redraw; run independent geometry checks, isolated vet/tests and race tests for new concurrent code.

## Scope limits

Simulated data only; no speech recognition, microphone, daemon management, remote control, transcript persistence or changes to Voxi.
