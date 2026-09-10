<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 027 — Introduce first spec-driven collector prototype

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [013](013-deterministic-live-snapshots-and-independent-rolling-histories.md), [017](017-external-file-and-socket-adapters-with-separate-producer-fixtures.md), [018](018-explore-bounded-linux-and-daemon-source-adapters.md), [019](019-evaluate-declarative-source-and-action-wiring.md), [spec conventions](../docs/Spec.md), [roadmap context](../docs/RoadmapContext.md)

---

## 1. Problem & Motivation

Loom has established the need for independent collection and redraw, but it
does not yet have a small typed collector boundary that a validated
specification can describe. Introduce the first narrow prototype: a typed
collector abstraction and a collector that reads a file at a fixed rate. The
same boundary must represent changing CPU/GPU-style numeric data without
coupling sampling to rendering.

This is a deliberately bounded follow-up to issues [013], [017], [018] and
[019]. It proves the spec-to-collector seam before broader Linux, daemon,
source, or action wiring is attempted.

## 2. Technical Specification / Findings

- Define the smallest typed collector contract needed to produce timestamped,
  safely published values or snapshots, including cadence, lifecycle,
  cancellation and read/error state.
- Add a `read file at fixed rate` implementation with explicit path, interval,
  parsing/type rules, bounded reads and deterministic test seams. A changing
  numeric file fixture must demonstrate values changing across samples.
- Keep collection, snapshot/history ownership and rendering as separate
  boundaries. Rendering consumes a snapshot and never advances collection or
  history merely by drawing.
- Express the prototype configuration through the repository's spec/schema
  direction, without hardcoded duplicate spec values or prematurely claiming
  a stable public DSL.

## 3. Implementation & Verification Plan

1. Inspect the current YAML/spec loading surface and establish the minimal
   schema/config shape for a typed file collector.
2. Implement the collector interface, fixed-rate file adapter and bounded
   snapshot publication behind a narrow package boundary; use an owned
   fixture or injectable reader/clock for deterministic tests.
3. Connect one simulated CPU/GPU-style changing-data example to the existing
   rendering path through snapshots, documenting the independent collection
   and redraw rates.
4. Document lifecycle, malformed/missing/stale input, retention and shutdown
   behavior, then record the prototype's limits and next decision for issues
   018/019.

### Acceptance criteria

- [ ] A typed collector abstraction is defined and consumed by a fixed-rate
  file collector with configurable path, cadence and bounded parsing/read
  behavior.
- [ ] A deterministic fixture shows changing CPU/GPU-style values producing
  timestamped samples at the collector rate, including startup, malformed or
  missing input, stale/error status and cancellation behavior.
- [ ] A renderer can redraw faster or slower than collection while consuming
  the same published snapshot; redraws do not add samples, mutate collector
  state or change sample timestamps/counts.
- [ ] The minimal spec/schema declaration is validated and does not introduce
  duplicated hardcoded spec values or an unbounded source framework.
- [ ] The prototype's file semantics, lifecycle, limits and compatibility
  status are documented, including what remains for issues 018 and 019.

### Verification

Use a fake clock and changing temporary fixture/reader to test fixed cadence,
multiple numeric values, empty and malformed content, missing files, read
errors, stale data, bounded input and cancellation. Assert independent sample
counts/timestamps while rendering at a different cadence, repeated rendering
of one snapshot, and clean shutdown. Run schema validation, focused Go tests,
`GOWORK=off go test ./...`, `GOWORK=off go vet ./...`, and relevant race tests.

## Scope limits

No external eventing, socket/event-stream transport, daemon integration,
production Linux hardware discovery, action wiring, persistence, plugin
system, generalized reactive graph or wholesale collector port. The file
reader is a prototype mechanism and not yet a universal source contract.
