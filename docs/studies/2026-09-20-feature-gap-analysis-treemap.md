# Treemap Feature-Gap Analysis

## Scope

Reviewed `examples/treemap/` against the current `graph` and terminal APIs. The
example is already thin around the core treemap algorithm: it calls
`graph.AggregateTreemap` and `graph.RenderTreemap` from
`examples/treemap/treemap/treemap.go`.

## Current framework coverage

- Partitioning is in `graph/layoutTreemap` (`graph/treemap.go`), including
  integer-cell tiling and orientation selection.
- Data reduction is in `graph.AggregateTreemap` (`graph/treemap_aggregate.go`),
  including node limits, merging, and “other” aggregation.
- Cell drawing, border themes, ANSI styling, label fitting/centering, marker
  fallback, legends, and terminal-width-aware row construction are in
  `graph.RenderTreemap` and its helpers (`graph/treemap.go`).
- Clipping and cursor-safe redraw are already framework features via
  `loom.RawScreen.Draw` and `loom.ClipRow`; the example only calls them from
  `examples/treemap/treemap/treemap.go`.

## Gaps and opportunities

### 1. Pluggable layout algorithms

`layoutTreemap` is a private value-splitting layout. There is no public choice
between slice-and-dice, squarified, or ordered layouts. Add a reusable layout
contract and options:

```go
type TreemapLayout interface {
    Layout(values []float64, bounds Rect) []Rect
}

type TreemapLayoutKind int
const (
    TreemapLayoutSliceDice TreemapLayoutKind = iota
    TreemapLayoutSquarified
)

type TreemapOptions struct {
    Layout TreemapLayoutKind
    // existing fields...
}
```

`RenderTreemap` should use the selected strategy; custom implementations could
also be accepted through `TreemapLayout`.

### 2. First-class cell model and renderer hooks

The renderer currently writes directly into rune/style grids in
`drawTreemapBox`, `drawTreemapThinBox`, and `drawTreemapMarker`.
Expose layout cells before rendering so applications can add interaction,
icons, tooltips, or alternate drawing without copying the renderer:

```go
type TreemapCell struct {
    Segment TreemapSegment
    Bounds  Rect
    Depth   int
}

func LayoutTreemap(segments []TreemapSegment, opts TreemapLayoutOptions) []TreemapCell
func RenderTreemapCells(cells []TreemapCell, opts TreemapRenderOptions) []string
```

A callback such as `CellRenderer func(CellCanvas, TreemapCell)` should remain
optional; the existing themed renderer can be the default.

### 3. Configurable color scaling

`examples/treemap/treemap/style.go` hard-codes the eight ANSI background/
foreground pairs in `demoPalette`. `graph.TreemapOptions` cycles palettes by
segment index, so value-based heat maps are not available. Add a color scale:

```go
type TreemapColorScale func(value, min, max float64) CellStyle

type TreemapOptions struct {
    ColorScale TreemapColorScale
    // existing palette fields remain as the fallback
}
```

Built-ins could include `LinearANSIGradient`, `QuantileANSI`, and a
foreground-contrast helper.

### 4. Label placement policy

`drawTreemapBox` in `graph/treemap.go` centers one full label and otherwise
uses a marker/legend fallback. The example cannot select top-left, truncating,
multi-line, or depth-aware labels. Add a policy hook:

```go
type TreemapLabelPlacement int
const (
    TreemapLabelCenter TreemapLabelPlacement = iota
    TreemapLabelTopLeft
)

type TreemapOptions struct {
    LabelPlacement TreemapLabelPlacement
    LabelFunc func(TreemapSegment, Rect) string
}
```

The framework should continue to enforce display-width-safe placement and
clipping.

### 5. Terminal canvas sizing

`examples/treemap/treemap/terminal.go` duplicates terminal detection,
80x24 fallback behavior, prompt-row reservation, and dimension clamping.
This is broadly useful for terminal apps, not treemap-specific. Consider:

```go
func TerminalCanvas(out io.Writer, opts CanvasOptions) (width, height int)
func ClampCanvas(width, height, terminalWidth, terminalHeight int) (Rect, ClampReport)
```

`RawScreen` could optionally accept the same canvas policy so rendering and
redraw use one source of truth.

### 6. Watch loop and default quit keys

`examples/treemap/treemap/input.go` implements raw `/dev/tty` polling and
reuses `loom.SpeccedDefaults.FallbackQuitKeys`. A framework-level
`loom.RunRawScreen(ctx, interval, draw, options)` could standardize signal
cancellation, quit-key decoding, cleanup, and resize-aware redraws.

## Priorities

1. Expose layout cells and add squarified layout support.
2. Add value-based color scales and label-placement policies.
3. Extract terminal canvas sizing and the raw watch loop if other examples need
   them; process-tree collection in `process.go` should remain application
   code.
