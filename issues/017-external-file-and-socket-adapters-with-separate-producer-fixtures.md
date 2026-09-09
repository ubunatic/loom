<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 017 — External file and socket adapters with separate producer fixtures

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 10 — external boundary
**Depends on**: [016](016-complete-harnez-and-voxi-simulated-ui-milestone.md) (transitive prerequisites apply).

## Problem and findings

Exercise the real process boundary before building system adapters. Current [YamlElement.Source](../yaml.go) is not proof of a generic external streaming contract. Files and sockets must be fed by separate processes, not only in-memory fakes.

## Acceptance criteria

- [ ] First retain minimal standalone canaries for actual file replacement/tailing and a local socket connection, observing framing and lifecycle before feature integration.
- [ ] Provide separate runnable fixture producer processes for a test file and a local socket, with a documented bounded record format. UI process consumes snapshots/events through adapters outside Draw.
- [ ] Define file semantics (replacement versus append), socket framing, poll/read timing, size limits and malformed/partial-record handling. Show missing file, producer exit/disconnect, reconnect and stale/error status without blocking rendering.
- [ ] Demonstrate producer restart and a slow/bursty stream with independent collection/redraw; no duplicate history on repeated redraw or unchanged file reads.
- [ ] Use temporary owned paths/socket names with deterministic cleanup. Adapter shutdown cancels reads and joins work; one-shot has a bounded initial read/timeout.

## Verification

Run separate-process file/socket canaries and integration fixtures, including partial writes, file replacement, disconnect/reconnect, oversized records and cancellation. Assert snapshot/history counts and responsiveness with the geometry gate; run isolated vet/tests and race checks. Keep instructions runnable without live Harnez/Voxi services.

## Scope limits

Local test files/sockets only; no authenticated remote transport, daemon installation, production Linux collectors or declarative source DSL yet.
