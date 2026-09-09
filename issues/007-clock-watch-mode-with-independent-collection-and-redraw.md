<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 007 — Clock watch mode with independent collection and redraw

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 2
**Depends on**: [006](006-static-declarative-monitor-shell-with-a-minimal-validated-contract.md) (transitive prerequisites apply).

## Problem and findings

[Pane.Run](../pane.go) redraws around terminal input/resize events and has no fixed-rate redraw scheduler. [Widget.Draw](../widget.go) is the rendering boundary; it must not become the clock collector.

## Acceptance criteria

- [ ] Extend the shell with a documented --watch path showing a ticking clock; preserve the same one-shot layout and exit behavior.
- [ ] Collect time independently into the latest snapshot; redraw on an explicitly configured refresh cadence. Rates and presentation come from validated declarations/configuration, while Go owns collection and lifecycle.
- [ ] Demonstrate a 1 Hz clock producer and 20 Hz redraw with an injected test clock: twenty draws between sample ticks do not call the producer or mutate state.
- [ ] Serialize UI updates or publish safe immutable snapshots; stop timers and producers on quit, cancellation and input failure, and restore the terminal. Invalid/nonpositive intervals fail clearly.
- [ ] Keep runtime changes additive and bounded to independent collection/redraw and shutdown; document the selected extension to Pane rather than assuming an existing ticker API.

## Verification

Use fake time/counters to assert independent tick counts, unchanged snapshots across redraws and prompt cancellation. Run GOWORK=off go vet ./..., go test ./... and go test -race ./... with GOWORK=off for each. Probe any new PTY/timer mechanism separately per Canary.md, then exercise idle-input watch, resize, quit and one-shot output.

## Scope limits

Clock only; no histories, hardware collectors, controls beyond quit, adaptive FPS or generalized scheduling/dataflow framework.
