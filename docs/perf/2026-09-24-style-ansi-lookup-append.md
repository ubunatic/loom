# Zero-Alloc Style ANSI Generation & Pre-Allocated Canvas.Row Buffering

## Overview
This optimization eliminates hot-path heap allocations and string concatenation churn during terminal cell style formatting and frame row rendering (`Canvas.Row`).

By introducing:
1. Pre-computed 256-color indexed ANSI lookup tables (`fgIndexTable` and `bgIndexTable`).
2. Zero-allocation byte slice appenders (`Color.appendFG`, `Color.appendBG`, `Style.AppendANSI`).
3. Pre-allocated byte buffer slice allocation in `Canvas.Row`.
4. Environment variable toggle via `LOOM_FAST_ANSI` (defaulting to fast path, falling back to legacy path when set to `"0"`, `"false"`, or `"off"`).

`Style.ANSI()` allocs dropped from **6** to **1** (or **0** when using `AppendANSI`), and `Canvas.Row` allocs dropped from **610** to **4** per row.

## Hardware Target
- **Topology:** 4–10 Cores, x86_64 / AVX-512, 32GB RAM.
- **Focus:** Eliminate GC churn during high-framerate TUI rendering and row flushes over standard stdout.

## Topology / Data Flow

```
[Canvas Cells] ---> [Canvas.Row] ---> [Style.AppendANSI([]byte)] ---> [Pre-Allocated []byte Buffer] ---> [Terminal Output]
                         |                         |
                         v                         v
               [Index Lookup Table]      [strconv.AppendUint (RGB)]
```

## Benchmark Evidence

### `Style.ANSI` Benchmark Delta
```
name                 old time/op    new time/op    delta
StyleANSI-4            697ns ± 0%     111ns ± 0%  -84.07%
StyleAppendANSI-4        N/A           53.3ns        N/A

name                 old alloc/op   new alloc/op   delta
StyleANSI-4            144B ± 0%       48B ± 0%  -66.67%
StyleAppendANSI-4        N/A             0B          N/A

name                 old allocs/op  new allocs/op  delta
StyleANSI-4            6.00 ± 0%      1.00 ± 0%  -83.33%
StyleAppendANSI-4        N/A           0.00          N/A
```

### `Canvas.Row` Benchmark Delta (120x40 Canvas, Styled Cells)
```
name                 old time/op    new time/op    delta
CanvasRow-4           92.5µs ± 0%    16.7µs ± 0%  -81.94%

name                 old alloc/op   new alloc/op   delta
CanvasRow-4          31.3kB ± 0%    18.9kB ± 0%  -39.61%

name                 old allocs/op  new allocs/op  delta
CanvasRow-4            610 ± 0%          4 ± 0%   -99.34%
```
