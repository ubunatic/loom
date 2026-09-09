<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 019 — Evaluate declarative source and action wiring

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Architecture
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 11 — feasibility
**Depends on**: [018](018-explore-bounded-linux-and-daemon-source-adapters.md) (transitive prerequisites apply).

## Problem and findings

Evaluate data declarations only after collection boundaries are proven. [YamlElement.Source, OnChange and StyleName](../yaml.go) exist, but their presence does not establish a functioning generic source/action system. The first shell's validated document is the starting point, not grounds for a new framework prerequisite.

## Acceptance criteria

- [ ] Compare a small declarative prototype with equivalent explicit Go wiring using one proven file source and one socket/line source; declare frequency and a simple value-to-bar/timeline mapping.
- [ ] Prototype one event-to-named-action connection with a registered Go handler. Validate unknown source/field/action IDs, duplicate wiring, invalid frequency and incompatible value types before runtime.
- [ ] Extend the existing minimal schema/document only for consumed prototype fields. Keep Go responsible for collection mechanics, state transitions and action behavior; declarations do not execute arbitrary shell expressions.
- [ ] Measure authoring complexity, error quality, cancellation, stale data and testability for declaration versus Go versions using the same fixtures.
- [ ] Deliver a concise adopt/narrow/reject decision, working bounded evidence (or a reproducible failed experiment), and explicit concepts that remain Go. Document compatibility implications for current YAML source/on_change/style fields.
- [ ] Do not silently ship an experimental syntax as a stable public contract; label its status and avoid breaking current widget declarations.

## Verification

Run schema negative tests, handler completeness checks, separate producer fixtures and collection/redraw invariant tests. Demonstrate one event reaches its intended handler exactly once under the chosen event semantics. Run isolated vet/tests for any prototype code and retain commands/results with the decision.

## Scope limits

Feasibility only: no expression language, general reactive graph, plugin/transport ecosystem, automatic production migration or large dataflow framework.
