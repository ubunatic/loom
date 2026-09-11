# 030 — Braille Activity Spinner and Bracketed Progress Bar Primitives

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [docs/HarnezSplashTarget.md](../docs/HarnezSplashTarget.md), [docs/Graph.md](../docs/Graph.md), [issues/024-port-harnez-rograph-primitives-with-provenance.md](024-port-harnez-rograph-primitives-with-provenance.md)

---

## 1. Problem & Motivation

The splash screen features two animated/graphical elements:
1. An animated braille activity spinner (e.g. `⠙`, `⠋`, `⠓`, `⠚`, `⠞`, etc.) next to the title `harnez usage`.
2. A fixed-width bracketed bar `[...]` rendering braille/dot-matrix fill patterns (such as `[⣿⣿⣿⣿⡇                   ]` or `[::::::::::::::::::::::::::::::::]`).

These primitives must render deterministically with sub-cell precision and correct visual width measurements without breaking terminal geometry.

## 2. Technical Specification

1. **Braille Spinner**:
   - Provide a spinner sequence definition (Braille dots, rotation cycles) with step/frame query function: `SpinnerGlyph(frame int) rune`.
   - Ensure cell width is strictly measured as 1 column.
2. **Bracketed Progress / Pattern Bar**:
   - Extend `graph.RenderBar` or create a specialized bracketed bar renderer supporting braille block characters (`⣿`, `⣷`, `⣤`, `⡇`, `⠶`) and dot-matrix fill styles.
   - Configurable outer brackets (e.g. `[` and `]`) and fixed target column width.
   - Support both determinate (0.0 to 1.0 ratio) and animated/indeterminate pulse modes.

## 3. Implementation & Verification Plan

- [x] Implement braille spinner frame cycle in graph/widget package.
- [x] Implement bracketed bar rendering with braille/dot-matrix fills.
- [x] Unit tests for visual column width stability across 0%, 50%, 100% and pulse steps.
- [x] Golden text tests verifying ANSI/Unicode width consistency.

