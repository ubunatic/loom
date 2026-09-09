<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 014 — Configurable graph colors and glyph presentation

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 8
**Depends on**: [013](013-deterministic-live-snapshots-and-independent-rolling-histories.md) (transitive prerequisites apply).

## Problem and findings

[Style](../style.go) already supports terminal colors and bold. The missing capability is declared visual semantics for graphs and status values, not a new ANSI styling engine. Copied graph options must receive resolved values without hardcoded application palettes.

## Acceptance criteria

- [ ] Declare and validate configurable level ranges, color stops, background, wrappers and glyph presentation; demonstrate at least two palettes and monochrome output.
- [ ] Apply equivalent level semantics to bars and timeline cells, including documented representative-value behavior for paired Braille samples.
- [ ] Support block, Braille and an explicitly configured alternative glyph set, with defined behavior for invalid ranges, unsupported width glyphs and missing styles.
- [ ] Apply bold/regular and status colors through the existing style boundary or a bounded adapter; prevent style leakage into labels, borders and subsequent rows.
- [ ] Changing colors or valid glyph presentation preserves allocated geometry and does not alter samples/history. Extend the consumed schema instead of adding hidden Go defaults.

## Verification

Test boundary/out-of-range levels and invalid declarations; compare independent final column positions for color on/off and all supported presentations. Replay colored frames for human review if available; otherwise record unattended ANSI/golden evidence. Run isolated vet/tests.

## Scope limits

No global palette registry, terminal-theme discovery, animation system or hardcoded Harnez-specific colors inside widgets.
