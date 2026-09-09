<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 009 — Declarative box visibility controls

**Status**: Closed — declared visibility actions and hints verified
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 4
**Depends on**: [008](008-responsive-declared-box-layout.md) (transitive prerequisites apply).

## Problem and findings

The target needs declared keyboard hints and pane toggles. Existing [HandleKey](../widget.go) and [Stack](../stack.go) focus behavior provide input mechanics, not a visibility/action declaration contract.

## Acceptance criteria

- [x] Declare stable box IDs, toggle keys, action IDs and displayed hints together; validate unique keys, existing targets and handler completeness. Go supplies small explicit action handlers.
- [x] Hide and restore each box; surviving boxes use available space in wide and slim layouts. Support both hidden with intact title/status, and restore without a restart.
- [x] Status/title hints reflect current visibility and remain reachable when every box is hidden. Hidden children receive no input or focus; restoring does not reset clock or other application state.
- [x] One-shot honors declared initial visibility; watch toggles and quit work through the same bounded action mapping.

## Verification

Drive KeyEvent sequences for all visibility combinations, repeated toggles and resize while hidden; assert deterministic rectangles and hints. Include invalid-key/target declaration tests, existing input regressions and isolated vet/tests.

## Scope limits

No general command language, dynamic plugin actions, mouse buttons, persistence or domain controls. Build after watch, not before it.

## Delivery evidence

Frame actions declare stable IDs, keys, handlers, targets, title hints and both
visibility hints. Box `hidden` controls initial state. Layout skips hidden boxes;
all-hidden height is two chrome rows. Watch and show-once share the action map.
Unit tests cover all visibility states, repeated toggles, resize, initial hiding,
invalid declarations and unchanged snapshots. `make test`, `go test -race ./...`,
PTY toggle-state/quit/restore and wide/slim/wide runs pass. Scope intentionally
keeps children inert rather than introducing focus dispatch.
