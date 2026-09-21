# Terminal Graph Primitives

<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

Loom provides dependency-free, fixed-width terminal graph rendering primitives under package `codeberg.org/ubunatic/loom/graph` ([`graph/`](../graph/)). These primitives render determinate progress bars and absolute/relative rolling sparklines designed specifically for terminal dashboards and box layouts.

Graph glyphs and padding glyphs are required to occupy one terminal cell. A
custom glyph with another measured width is replaced by a one-cell fallback so
exact-width output remains true. This is a bounded cell policy, not a claim of
full grapheme-cluster or terminal-emulator conformance.

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

## 2. Stacked Bars and 2D Treemaps (`RenderStackedBar`, `RenderTreemap`, `AggregateTreemap`)

Two distinct chart shapes exist for showing relative proportions, deliberately kept separate:

- **`RenderStackedBar`** (`graph/stackedbar.go`) is the original single-row proportional bar — glyphs filling one line, each segment's share of the row proportional to its value. It is not a treemap; it was briefly built and named as one before being corrected and given its own file/name.
- **`RenderTreemap`** (`graph/treemap.go`) is a real 2D treemap: rectangular boxes tiling a `Width x Height` character grid, gap-free, non-overlapping, each box's area proportional to its segment's value.

### Layout algorithm

`layoutTreemap` is a recursive binary split (not squarified): segments sort by value descending, partition as evenly as possible by cumulative value, and the rectangle divides between the two partitions along whichever axis keeps its children **landscape** (wider than tall) rather than aiming for a square. `treemapLandscapeBias` (3.0) biases this deliberately above what "square in cell count" would need, because terminal cells are themselves roughly twice as tall as wide and it's horizontal text that needs to fit inside a box.

### Label / marker / legend rule: full label or a number, never a fragment

A box's full label (name, plus value if `ShowValues`) is shown only when it fits without truncation. A box too small to describe gets a superscript number (`¹²³...`) instead — **never a truncated fragment of its name** ("fir…" is never acceptable; the number is the fallback for "cannot be described," not a size hint). Numbered boxes are listed on one legend row below the grid, ranked biggest-value-first, hard-truncated only whole-entry-at-a-time with a trailing `…` when the legend itself runs out of room.

This rule has bitten this codebase twice in fixable ways worth knowing about if touching this code again:

1. A theme's own pre-pass must be the thing that decides "does this box get a number instead of a label" — if a theme's draw path can reach its label-placement code without that decision having been made for it first (e.g. by leaving a shared "does it need a marker" slice unpopulated), it will silently fall through to truncating instead of using the fallback it already has available.
2. `truncateLabel` (an ellipsis-truncation helper) turned out to be permanently unreachable dead code once this invariant is actually upheld everywhere — deleted rather than left as an attractive nuisance for a future change to accidentally call.

### `AggregateTreemap` (`graph/treemap_aggregate.go`)

Turns an arbitrary tree (e.g. a process tree) into a flat `[]TreemapSegment` bounded by a node budget (`MaxTreemapNodes = 100`): merges same-name siblings (browser tabs, worker processes), then expands the heaviest nodes first up to the budget. Expansion is **partial, not all-or-nothing** — a node whose full expansion would overflow the budget gets its top children kept plus the rest folded into an "other" child, rather than being locked out of expanding at all. (The all-or-nothing version was a real bug: a single bushy node like `systemd` with 60+ children could never expand, becoming one oversized, uninformative blob.)

### Theming (`TreemapOptions.Theme`)

- **`TreemapThemeClassic`** (zero value): a box-drawing border (`┌─┐│└─┘`) when it fits, else a solid glyph/ANSI-background fill.
- **`TreemapThemeNumbered`**: full-bleed (no border cell — labels get the whole box), with a seven-eighths block glyph (`▇` top row, `▉` right column) softening just those two edges instead of a hard border line — the glyph's thin unfilled sliver carries no background SGR at all, so the ambient terminal background shows through it, purely locally to each box's own rectangle (no neighbor lookup). Every box gets a superscript corner marker, ranked by value across the whole treemap, not just the ones needing one in place of a label; the legend still only explains the latter. Requires `ANSI`/`BackgroundANSI`; falls back to `TreemapThemeClassic`+`NoBorder` without them.

See issue 040 for the full design-iteration history (two earlier theme designs — half-block neighbor-blended edges, and a filled-corner-number variant — were built, live-reviewed against real data, and dropped) and [`docs/studies/2026-09-14-terminal-safety-hardening-and-treemap-theme-iteration.md`](studies/2026-09-14-terminal-safety-hardening-and-treemap-theme-iteration.md) for the broader lessons from that iteration.

### `examples/treemap` and terminal safety

`RenderTreemap` has no way to know the real terminal width — every row it returns is *exactly* `opts.Width` columns, and if that's wider than the terminal actually printing to, the terminal's own auto-wrap corrupts every subsequent row (see `RenderTreemap`'s doc comment, and [`docs/TerminalSafety.md`](TerminalSafety.md) for the general pattern and the `loom.RawScreen`/`loom.WriteRows`/`loom.ClipRow` primitives that guard against it). `examples/treemap` is the worked example of a caller doing this correctly, including a `--watch` mode and a `--theme` flag.

---

## 3. Architectural Invariants for Loom

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

## 4. Geometry Gate Alignment

The output of `graph` primitives adheres to Loom's display-width invariants verified in [`docs/Geometry.md`](Geometry.md):
- Standard ASCII brackets (`[` and `]`) and eighth-block characters (`\u2588`–`\u258f`) consume exactly 1 terminal display cell.
- The entire Unicode Braille Patterns block (`\u2800`–`\u28ff`) consumes exactly 1 terminal cell per glyph.
- ANSI SGR sequences do not contribute to measured cell width.

## 5. Treemap Cells and Color Scales (`LayoutTreemap`, `ColorScale`)

`graph.LayoutTreemap(segments, w, h, layout)` returns integral `TreemapCell`s. The layout is
slice-dice or squarified (float squarify plus cumulative rounding so cells tile the area
exactly); squarified gives visibly better aspect ratios for skewed data. `ColorScale` maps a
value in a range to a cell style; `LinearColorScale(from, to)` interpolates RGB. Delivered by
ticket 081; frames in `docs/progress/081/`.
