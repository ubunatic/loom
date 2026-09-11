# Graph Package Provenance

<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

## Origin and Attribution

- **Upstream Project**: Harnez (`github.com/ubunatic/harnez`)
- **Source Revision**: `01e59b331c9d85699e55fbbb946c579e19901083`
- **Original Path**: `internal/rograph/`
- **Author**: Uwe Jugel
- **Authorization & Metadata Findings**:
  - The user explicitly authorized copying Harnez rograph into Loom while retaining it in Harnez; no renewed copy permission is needed.
  - Inspection of upstream Harnez at revision `01e59b331c9d85699e55fbbb946c579e19901083` confirmed no root `LICENSE` or `REUSE.toml` file and no in-file license notices or SPDX tags in `internal/rograph/{bar,sparkline,options}.go`. Upstream git commit history identifies Uwe Jugel as author.
  - Ported files are licensed under `AGPL-3.0-or-later` in Loom consistent with repository licensing ([REUSE.toml](../REUSE.toml)).

## Ported Files

The following files from `harnez/internal/rograph/` were ported into `codeberg.org/ubunatic/loom/graph`:

1. `bar.go` -> `graph/bar.go` (`RenderProgressBar`)
2. `sparkline.go` -> `graph/sparkline.go` (`PercentSparkline`, `percentSparkChars`)
3. `options.go` -> `graph/options.go` (`RenderBar`, `RenderSparkline`, `RenderPercentSparkline`, `FormatPercent`, `BarOptions`, `SparklineOptions`, Braille and sub-character eighth-block calculations)
4. `rograph_test.go` & `options_test.go` -> `graph/graph_test.go` (table-driven unit tests for bars, sparklines, Braille glyph levels, sub-character quantization, ANSI escaping, and exact-width padding)

## Architectural Adaptations for Loom

1. **Package Scope and Isolation**:
   - Ported under `codeberg.org/ubunatic/loom/graph`.
   - Preserved zero external dependencies (standard library only: `fmt`, `math`, `strings`, `unicode/utf8`).

2. **Elimination of Mutable Global State**:
   - Harnez defined `var DefaultBackgroundANSI = "100"`, which was modified at runtime by spec loaders and caused mutable global coupling.
   - Loom defines `const DefaultBackgroundANSI = "100"` as an immutable constant.
   - Callers configure `BackgroundANSI` per invocation via `BarOptions` or `SparklineOptions`. A value of `"none"` cleanly suppresses the background SGR code without requiring global state mutation.

3. **Guaranteed Exact Column Width Padding**:
   - Harnez's `PercentSparkline` emitted only `len(pcts)` runes when fewer samples than `maxWidth` were available, and `RenderSparkline` returned `""` on empty input. In a fixed-column terminal dashboard or box layout, this caused line lengths to collapse or misalign.
   - Loom enforces exact-width padding:
     - `PercentSparkline` always renders exactly `maxWidth` (minimum 1) glyphs, left-padding short histories and empty inputs with the idle glyph (`'▁'`).
     - `RenderSparkline` pads short histories on the left and empty inputs across all `width` cells with the presentation's idle glyph (`'▁'` for blocks, `'⣀'` for Braille) or `opts.PadRune`.
     - Callers requiring raw, unpadded lengths can opt out via `opts.NoPad = true`.
