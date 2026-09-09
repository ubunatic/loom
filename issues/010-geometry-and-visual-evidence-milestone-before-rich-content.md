<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 010 — Geometry and visual evidence milestone before rich content

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 5
**Depends on**: [009](009-declarative-box-visibility-controls.md) (transitive prerequisites apply).

## Problem and findings

The rich-content gate must detect actual terminal geometry failures. [Canvas.Write](../canvas.go) clips only to canvas bounds, not child Rect, advances combining runes by one, and uses coarse RuneWidth ranges; StringWidth counts ANSI bytes/runes as visible. Cell-only assertions or expectations calculated with those same helpers can bless broken output. [Render and ScreenshotScript](../screenshot.go) produce headless ANSI rows and replay scripts, not raster screenshots.

## Acceptance criteria

- [ ] Establish and enforce child content clipping, including left/right/top/bottom boundaries, so oversized text or a deliberately over-writing child cannot damage its box border, sibling or chrome.
- [ ] Validate final emitted display positions independently of production RuneWidth/StringWidth/Canvas.Write. Use explicit known terminal-column fixtures and an independent ANSI/terminal interpretation or equivalent oracle; internal cell counts alone are insufficient.
- [ ] Cover ASCII, combining marks, CJK/wide glyphs, block/Braille, styled text, wide glyphs at the right edge, zero/tiny child regions and repeated resize/visibility cycles. Define the supported Unicode/ANSI policy and address defects needed for those fixtures.
- [ ] Preserve exact border rectangles, row widths, padding and chrome across wide/slim/tiny layouts. Include a deliberately broken border/width/child-clipping case that the gate demonstrably rejects.
- [ ] Save deterministic golden ANSI output plus replay evidence with dimensions, commands and expected positions. Seek human visual confirmation when available; otherwise record unattended evidence, oracle results and limitations and proceed after passing. Never label ANSI replay as a raster screenshot.
- [ ] Make this a blocking dependency for rich rows/graphs and later visual work; use a standalone canary before relying on a terminal emulator or raster capture mechanism.

## Verification

Run independent final-output geometry checks, negative controls, golden/replay comparisons and isolated vet/tests. When a human is available record reviewer/result; otherwise record 'unattended', artifacts, dimensions and passed assertions. A new capture stack is optional, not a reason to stall on missing desktop access.

## Scope limits

Bound to shell geometry and representative text fixtures. No graphs or realistic dashboards before this gate, no exhaustive terminal conformance suite, and no assumption that Voxi width helpers are an authoritative Unicode oracle.
