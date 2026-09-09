<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 011 — Aligned dashboard rows and ANSI-safe truncation

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 6 — rows
**Depends on**: [010](010-geometry-and-visual-evidence-milestone-before-rich-content.md) (transitive prerequisites apply).

## Problem and findings

[Table.computeWidths and padCol](../table.go) provide table-specific sizing and rune trimming, not reusable stable dashboard rows or ANSI-safe ellipsis. [Voxi TruncateLineANSI](../../voxi/internal/monitor/render.go) is a behavior reference, but returns three dots even for budgets below three and must not be copied as a correctness oracle.

## Acceptance criteria

- [ ] Declare reusable aligned label/value rows within the boxes; label and reserved value columns stay fixed as values change from 99% to 100% and durations/names lengthen.
- [ ] Fit rows to actual inner width with an explicit overflow priority and ellipsis policy. Widths 0, 1 and 2 never exceed budget; preserve valid style boundaries and do not split supported combining/wide text.
- [ ] Demonstrate static labels, timestamps/durations, multiple value columns and placeholders for later graphs using fixed fake data.
- [ ] Reuse the geometry contract for clipping and final positions; declaration properties have schema coverage and Go contains no target-specific column coordinates.

## Verification

Add table-driven alignment/truncation cases for ANSI, bold/regular, combining/CJK, tiny widths and long values; compare final display positions with the independent geometry gate. Run wide/slim goldens and isolated vet/tests.

## Scope limits

No graph algorithms, collectors, live histories or table-widget rewrite. Take Voxi behavior as inspiration without importing its monitor domain or mutating the sibling.
