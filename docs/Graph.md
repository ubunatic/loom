# Terminal Graph Primitives

<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

Loom provides dependency-free, fixed-width terminal graph rendering primitives under package `codeberg.org/ubunatic/loom/graph` ([`graph/`](../graph/)). These primitives render determinate progress bars and absolute/relative rolling sparklines designed specifically for terminal dashboards and box layouts.

The implementations are ported and adapted from Harnez's `internal/rograph` (see [`graph/PROVENANCE.md`](../graph/PROVENANCE.md)).

---

## 1. Core Primitives

### Determinate Fill Bars (`RenderBar`, `RenderProgressBar`)

- **Shrink-to-Fit**: Renders at requested character width, clamping to a minimum of 1 character to prevent negative-width strings or panics under narrow constraints.
- **Sub-Character Precision (`SubChar`)**:
  - Horizontal eighth-blocks (`▏▎▍▌▋▊▉`) achieve 8 subdivisions per character cell along the fill boundary rather than snapping to whole cells.
  - Braille sub-character fill (`SubCharacterGlyphs: []rune{'⡇'}`) achieves 2 subdivisions per cell (empty `⠀`, half `⡇`, full `⣿`).
  - Cells before the boundary remain fully filled (`█` or `opts.Fill`); cells after remain empty (`░`, space ` `, or `opts.Empty`).
- **ANSI Styling & Wrappers**:
  - Optional bracket wrappers (`[` and `]`, or custom `opts.Left`/`opts.Right`), suppressible via `opts.NoWrapper: true`.
  - ANSI background wrapping applies strictly to the inner graph glyphs, leaving brackets and trailing percent labels outside the SGR sequence to prevent escape leakage across layout columns.

### Rolling Time-Series Sparklines (`PercentSparkline`, `RenderSparkline`, `RenderPercentSparkline`)

- **Chronological Flow**: Time flows left to right (oldest sample at left, newest observation at right).
- **Presentation Modes**:
  - `SparklineBlocks`: 8-level vertical block sequence (`▁▂▃▄▅▆▇█`). Each terminal cell represents one sample pair, displaying the newer sample so latest observations are always represented.
  - `SparklineBraille`: Double-resolution btop-style sparkline where each character cell encodes two chronological samples across two 4-dot bottom-aligned Braille columns (left: dots 7,3,2,1; right: dots 8,6,5,4).
- **Scale Modes**:
  - `FixedRange: true`: Absolute scale (e.g. 0–100 for percentages), ensuring a 5% reading appears idle rather than filling the chart.
  - Relative: Derives dynamic minimum and maximum from the visible sample window. Flat series render at the middle glyph.

---

## 2. Architectural Invariants for Loom

### Invariant 1: Guaranteed Exact-Width Padding

In terminal box layouts (such as Loom's `rows` widget in [`widget_rows.go`](../widget_rows.go)), row alignment depends on each column consuming an exact, predictable display width.

- **The Problem in Harnez**: Upstream `PercentSparkline` emitted only `len(pcts)` glyphs when fewer samples than `maxWidth` were provided, and `RenderSparkline` returned `""` on empty slices. In a multi-column row, short or empty sample histories caused lines to collapse or drift out of alignment.
- **The Loom Contract**:
  - `PercentSparkline` always outputs exactly `maxWidth` characters (wrapped in ANSI if enabled). When `len(pcts) < maxWidth`, older timestamps on the left are padded with the idle glyph (`'▁'`).
  - `RenderSparkline` always outputs exactly `opts.Width` cells. When sample history is shorter than `2 * Width`, or when `len(values) == 0`, missing cells on the left are padded with the presentation's idle glyph (`'▁'` for blocks, `'⣀'` for Braille) or `opts.PadRune`.
  - Callers requiring raw, unpadded lengths can opt out explicitly by setting `opts.NoPad: true`.

### Invariant 2: Zero Mutable Global State

In accordance with Loom's Go conventions ([`docs/Go.md`](Go.md)):

- **No Mutable Package Globals**: Harnez defined `var DefaultBackgroundANSI = "100"`, which was modified at runtime by spec loaders and introduced hidden cross-package coupling.
- **Immutable Constants with Per-Call Overrides**: Loom defines `const DefaultBackgroundANSI = "100"` as an immutable string constant. Callers customize background colors per call using `BarOptions.BackgroundANSI` or `SparklineOptions.BackgroundANSI`.
- **Clean Background Suppression**: Passing `BackgroundANSI: "none"` or `"-"` suppresses the background escape completely, allowing transparent/terminal-native backgrounds without mutating global configuration.

### Invariant 3: Clean Layer Isolation

- The `graph` package depends **only** on the Go standard library (`fmt`, `math`, `strings`, `unicode/utf8`).
- It has no awareness of YAML specs, spec loaders, file systems, metrics collectors, or OS APIs.
- Callers (such as declarative widgets or live collectors) are responsible for resolving spec values and passing plain numbers and options to the graph functions.

---

## 3. Geometry Gate Alignment

The output of `graph` primitives adheres to Loom's display-width invariants verified in [`docs/Geometry.md`](Geometry.md):
- Standard ASCII brackets (`[` and `]`) and eighth-block characters (`\u2588`–`\u258f`) consume exactly 1 terminal display cell.
- The entire Unicode Braille Patterns block (`\u2800`–`\u28ff`) consumes exactly 1 terminal cell per glyph.
- ANSI SGR sequences do not contribute to measured cell width.
