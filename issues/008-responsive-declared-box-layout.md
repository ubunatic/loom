<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 008 — Responsive declared box layout

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 3
**Depends on**: [007](007-clock-watch-mode-with-independent-collection-and-redraw.md) (transitive prerequisites apply).

## Problem and findings

[Stack.childRect](../stack.go) splits a fixed direction equally and [Grid.Draw](../grid.go) uses uniform cells. Neither implements a declared width breakpoint or chrome-aware responsive box allocation.

## Acceptance criteria

- [ ] One validated declaration places both boxes side by side at/above a declared breakpoint and stacks them below it, preserving declaration order.
- [ ] Allocate title/status chrome, borders, padding and gaps before child content; no negative rectangles or overlap. Define deterministic behavior when terminal width or height cannot fit minimum dimensions.
- [ ] Resize across the breakpoint in both directions during watch without losing clock state; one-shot and watch use the same layout calculation.
- [ ] Keep terminal coordinate arithmetic inside generic layout code, not the example or its producer; preserve existing Stack/Grid consumers.

## Verification

Table-test breakpoint minus one, exactly breakpoint, plus one, wide/slim, and tiny width/height with explicit expected rectangles. Render headless fixtures and exercise watch resize. Run isolated vet and tests; document the example invocation and supported size policy.

## Scope limits

Two-box responsive row/column composition only; no CSS/flexbox engine, weighted grids, rich content or visibility controls yet.
