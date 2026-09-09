<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 018 — Explore bounded Linux and daemon source adapters

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Architecture
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 10 — exploratory real adapters
**Depends on**: [017](017-external-file-and-socket-adapters-with-separate-producer-fixtures.md) (transitive prerequisites apply).

## Problem and findings

Actual Linux metrics and daemon/transcript formats require fresh inspection and mechanism probes. The [Harnez/Voxi target](../docs/HarnezUsageTarget.md#combined-porting-target) identifies reference domains but does not specify a portable producer protocol or supported host matrix.

## Acceptance criteria

- [ ] Inspect current Harnez/Voxi collectors read-only and record concrete paths, protocols, units, availability and licensing before proposing any reuse; do not assume monitor render.go contains the collectors.
- [ ] Probe one read-only Linux metric source and one daemon/state or transcript source available on the host. If unavailable, retain a separate-process fixture for the observed/documented format and record the missing host capability.
- [ ] Implement at most one small adapter for each selected source behind the existing snapshot/event boundary, or deliver a documented no-go finding with evidence when feasibility fails.
- [ ] Specify sample cadence, timestamps, missing/stale/error behavior and bounded cancellation. Show that a stalled real source leaves UI refresh and independent producers working.
- [ ] Record a concrete recommendation and remaining source-specific follow-ups; observations may refine later scope without turning this into an all-system-monitoring project.

## Verification

Keep standalone canary commands/results, fixture parsing and lifecycle tests, plus a bounded live smoke check when available. Run isolated vet/tests and relevant race checks if code is added. Preserve the complete simulated examples as regressions.

## Scope limits

Exploratory, not a promise of universal GPU/daemon support. No credentials, service reconfiguration, install/start/stop actions, remote hosts, sibling mutation or wholesale collector port.
