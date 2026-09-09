<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 024 — Port Harnez rograph primitives with provenance

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [012](012-copy-harnez-rograph-with-verified-provenance-and-bounded-adapters.md), [Harnez target](../docs/HarnezUsageTarget.md), [Roadmap](../docs/Roadmap.md)
**Roadmap stage**: 6 — graphs (sub-ticket of 012)
**Depends on**: [011](011-aligned-dashboard-rows-and-ansi-safe-truncation.md)

## Problem and findings

The user authorized copying the dependency-free `rograph` renderer from Harnez (`../harnez/internal/rograph/`) into Loom.
The source files (`bar.go`, `sparkline.go`, `options.go`) contain:
- `RenderBar` for determinate bars with sub-character Braille/block precision.
- `PercentSparkline` and `RenderSparkline` for absolute-scale rolling percentage timelines.
- Sub-character Braille dot mappings.

Currently, `PercentSparkline` emits one glyph per sample without automatic full-width padding for short or empty histories, and Harnez uses a mutable global `DefaultBackgroundANSI`. Loom requires:
1. Copying the files into an isolated package (`codeberg.org/ubunatic/loom/graph` or `graph.go`) with full provenance, source git revision, and license attribution documented.
2. Adapting the rendering functions to Loom's cell and styling conventions without mutable globals.
3. Enforcing exact width padding for empty and short histories.

## Acceptance criteria

- [x] Record the source revision (`01e59b331c9d85699e55fbbb946c579e19901083`), copied file list, and verified provenance.
- [x] Copy and adapt `RenderBar`, `PercentSparkline`, and `RenderSparkline` to Loom's cell/style conventions without external dependencies or global state.
- [x] Guarantee exact allocated column widths even when sample histories are empty or shorter than the requested width.
- [x] Port focused unit tests from Harnez for boundary values, empty histories, 2-sample Braille cells, and sub-character precision.
- [x] Ensure all tests pass under `go test -race ./...` and `make test`.

## Verification

- `graph/PROVENANCE.md` records upstream commit `01e59b331c9d85699e55fbbb946c579e19901083`, source files, and architectural adaptations.
- Zero mutable package globals: `DefaultBackgroundANSI` is an immutable constant; background codes are resolved per options call with `"none"` support.
- Short and empty sample arrays are automatically padded on the left to guarantee exact `Width`/`maxWidth` rune counts, preventing dashboard row misalignments. `NoPad: true` remains available for callers requesting raw slice lengths.
- `go test -race ./...` and `make test` pass cleanly.

## Scope limits

Limited to pure graph rendering primitives and tests. Integrating these renderers into the declarative monitor rows widget is handled in [025](025-integrate-graph-renderers-into-declarative-monitor.md).
