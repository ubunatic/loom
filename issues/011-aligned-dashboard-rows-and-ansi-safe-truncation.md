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

## Implementation progress — 2026-09-10

First small increment: `TruncateText(text, width, marker)` implements bounded
terminal-cell truncation with caller-declared marker text, including budgets
0/1/2, combining clusters and complete wide glyphs. Raw ANSI is stripped under
the geometry policy; styles are supplied separately, not embedded in strings.
Tests use the independent emitted-cell oracle rather than production width
helpers for the budget assertion. Vet, full tests and race checks pass.

Second small increment: box `rows` now declares fixed-width columns, alignment,
optional bold style, gaps, ellipsis and fixed string values. Runtime validation
rejects invalid widths/alignment and mismatched value counts; schema covers all
fields. Leftmost columns have overflow priority; trailing columns clip or vanish.
The monitor shows three dummy usage rows and three dummy hardware rows, entirely
from YAML. Headless geometry tests cover wide/slim/tiny layouts; alignment tests
cover 99%/100%, long names/durations and styled columns.

Still open: graph-placeholder demonstration and final ticket-level review of
all acceptance criteria. Values are static, not collected system measurements.
