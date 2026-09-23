# ANSI Escape Sequence Parsing & Canvas Cell Writing Optimization

## Overview
`ParseANSI` and `Canvas.WriteANSI` are critical paths in Loom when rendering ANSI terminal streams and styled text buffers. Previously, `ParseANSI` called `measure.Clusters(string(rs[i:]))` on every single character iteration. This allocated new string slices of the remaining text tail at every rune position. In addition, `Canvas.set` invoked full unicode cluster analysis for every cell, even when painting simple 1-byte ASCII characters or spaces.

This change preserves the legacy `ParseANSIOld` and `Canvas.setOld` functions for reference and comparative testing, while introducing optimized `ParseANSINew` and `Canvas.setNew` functions. The default runtime mode uses the optimized path, and setting the environment variable `LOOM_FAST_ANSI` to `"0"`, `"false"`, or `"off"` toggles back to the legacy implementation.

## Hardware Target
- Target OS: Modern Linux x86_64
- Topology: 4-10 Physical Cores / APU / Cloud Runners
- Constraints: Multi-core bounds (min(runtime.NumCPU(), 10))

## Topology / Data Flow
```
[ ANSI String Stream / LOOM_FAST_ANSI ]
               │
      ┌────────┴────────┐
      ▼                 ▼
[ Fast Path ]     [ Legacy Path ] (LOOM_FAST_ANSI=0)
  (Default)             │
      │                 │
      ▼                 ▼
[ In-Place Scan ] [ Suffix Slicing ]
      │                 │
      ▼                 ▼
[ Canvas.setNew ] [ Canvas.setOld ]
```

## Benchmark Evidence

### `ParseANSI` Benchmark Delta
- **Old Path (`ParseANSIOld`):** `461,119 ns/op`, `222,401 B/op`, `4,721 allocs/op`
- **New Path (`ParseANSINew`):** `10,712 ns/op`, `5,104 B/op`, `99 allocs/op`
- **Improvement:** **~43x faster execution**, **97.7% reduction in allocated memory**, **97.9% reduction in allocations**.

### `Canvas.WriteANSI` Benchmark Delta
- **Old Path (`CanvasWriteANSI_Old`):** `512,724 ns/op`, `226,271 B/op`, `5,121 allocs/op`
- **New Path (`CanvasWriteANSI_New`):** `28,403 ns/op`, `5,745 B/op`, `179 allocs/op`
- **Improvement:** **~18x faster execution**, **97.4% reduction in allocated memory**, **96.5% reduction in allocations**.
