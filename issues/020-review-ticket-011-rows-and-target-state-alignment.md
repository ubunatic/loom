<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 020 — Review of Ticket 011 increments and target state alignment

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Planning / Review
**Related**: [011](011-aligned-dashboard-rows-and-ansi-safe-truncation.md), [012](012-copy-harnez-rograph-with-verified-provenance-and-bounded-adapters.md), [013](013-deterministic-live-snapshots-and-independent-rolling-histories.md), [Harnez target](../docs/HarnezUsageTarget.md), [Roadmap](../docs/Roadmap.md), [Spec](../docs/Spec.md)
**Roadmap stage**: 6 — rows
**Depends on**: [011](011-aligned-dashboard-rows-and-ansi-safe-truncation.md)

## Problem and findings

Work on [ticket 011](011-aligned-dashboard-rows-and-ansi-safe-truncation.md) is currently mid-flight (commits `94c208e` and `c19f226`). A thorough architectural review against the envisioned target state documented in [HarnezUsageTarget.md](../docs/HarnezUsageTarget.md), [Roadmap.md](../docs/Roadmap.md), and [Spec.md](../docs/Spec.md) reveals key successes, design trade-offs, and critical gaps to address before closing ticket 011 and moving to [ticket 012](012-copy-harnez-rograph-with-verified-provenance-and-bounded-adapters.md):

1. **Successful foundation**:
   - `TruncateText` in [truncate.go](../truncate.go) provides bounded terminal-cell budget enforcement, handling combined clusters and wide CJK/Braille runes without split glyphs, tested against independent cell-width checks in [truncate_test.go](../truncate_test.go).
   - `Rows` in [rows.go](../rows.go) introduces declarative column width reservation, left/right alignment, gap handling, and leftmost column priority under space constraints, isolated within child canvas bounds.

2. **Gaps against Ticket 011 Acceptance Criteria and Target State**:
   - **Missing graph placeholders**: Acceptance criteria 3 specifies: *"Demonstrate static labels, timestamps/durations, multiple value columns and placeholders for later graphs using fixed fake data."* Current [examples/monitor/spec/monitor.yaml](../examples/monitor/spec/monitor.yaml) defines only 3 columns (`name`, `percent`, `duration` / `detail`). The reference targets in [HarnezUsageTarget.md](../docs/HarnezUsageTarget.md) contain two determinate usage bars per row in All Usage (`[⣿⣿  ]` and `[    ]`), rolling timelines in Load (`[⣀⣀⣀⣀⣀⣀⣀⣀⣀⣀]`), and split dual timelines for VRAM/GTT (`[⣿⣿⣿⣿][⣀⣀⣀⣀]`). The current example omits placeholder columns entirely.
   - **Vertical height and row count mismatch**: In [monitor.yaml](../examples/monitor/spec/monitor.yaml), the boxes are defined with `height: 7` and `padding: 1`. With top/bottom borders (2) and vertical padding (2), the inner height is exactly 3 rows. The reference target in [HarnezUsageTarget.md](../docs/HarnezUsageTarget.md) has 4 rows in each box (Usage: `Claude Code`, `Gemini`, `Claude/GPT`, `OpenAI Codex`; Load: `cpu`, `ram`, `gpu`, `gpu vram/gtt`). Adding the 4th row causes silent vertical truncation under `paintClipped`. Box height must be increased to `8` (or vertical padding adjusted) and total frame height adjusted (e.g. to 10) to accommodate all 4 target rows.
   - **Right-alignment truncation orientation**: In [rows.go](../rows.go#L42-L47), `offset := width - StringWidth(text)` right-aligns text, but `TruncateText` always truncates the right side of the string. For numeric metrics or timestamps in right-aligned columns, truncating the trailing digits (e.g. `12345` -> `12...`) obscures the least significant values or unit suffixes.
   - **Styling and ANSI boundaries**: Ticket 011 specifies *"preserve valid style boundaries"*. `TruncateText` strips all ANSI escape sequences, delegating styling entirely to `RowColumn.Bold`. While appropriate for monochrome geometry validation, it does not support colors or per-cell styles needed for status colors (green/yellow/red) in Stages 8–9.
   - **Static YAML values vs dynamic data binding**: Hardcoding `values: [][]string` directly into `monitor.yaml` validates layout declaration, but the long-term target requires data snapshots to be supplied by Go code. The design needs an interface or method on `Rows` (e.g. `SetValues` or a dynamic snapshot provider) so [ticket 013](013-deterministic-live-snapshots-and-independent-rolling-histories.md) can update rows without rewriting YAML documents at runtime.

## Acceptance criteria

- [ ] Complete Ticket 011 scope by adding graph placeholder columns to [examples/monitor/spec/monitor.yaml](../examples/monitor/spec/monitor.yaml) with static Braille/block dummy data matching [HarnezUsageTarget.md](../docs/HarnezUsageTarget.md).
- [ ] Adjust box dimensions and inner layout capacity to fit all 4 canonical target rows in both the Usage and Load panels without vertical overflow.
- [ ] Ensure right-aligned columns handle truncation consistently without displacing column alignment boundaries.
- [ ] Expose an update/binding interface on `Rows` so Go can supply row values programmatically while YAML preserves layout, column widths, and alignment structure.
- [ ] Complete Ticket 011 delivery evidence and verification before transitioning to Ticket 012 rograph porting.

## Verification

Run `go test ./...`, `make test`, and `go run ./examples/monitor` to verify that the 4-row layout with graph placeholders renders accurately, aligns across all columns, and respects responsive terminal widths (64-col wide, 40-col slim).

## Scope limits

Graph algorithms, Braille interpolation, live data collection, and palette-based color mapping belong to tickets 012, 013, and 014. This review coordinates the completion of ticket 011 and structural alignment with subsequent tickets.
