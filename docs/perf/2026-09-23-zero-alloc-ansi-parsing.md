# ANSI Escape Sequence Parsing & Canvas Cell Writing Optimization

## Overview
`ParseANSI` and `Canvas.WriteANSI` are critical paths in Loom when rendering ANSI terminal streams and styled text buffers. Previously, `ParseANSI` called `measure.Clusters(string(rs[i:]))` on every single character iteration. This allocated new string slices of the remaining text tail at every rune position. In addition, `Canvas.set` invoked full unicode cluster analysis for every cell, even when painting simple 1-byte ASCII characters or spaces.

This optimization refactors `ParseANSI` to perform direct in-place scanning for base runes and attached combining marks over `[]rune`, pre-allocating the `cells` buffer capacity. Furthermore, `Canvas.set` introduces a zero-allocation fast-path for single-byte printable ASCII characters.

## Hardware Target
- Target OS: Modern Linux x86_64
- Topology: 4-10 Physical Cores / APU / Cloud Runners
- Constraints: Multi-core bounds (min(runtime.NumCPU(), 10))

## Topology / Data Flow
```
[ ANSI String Stream ]
          │
          ▼
 [ In-Place Rune Scan ]  ◄── Pre-allocated []Cell buffer (cap = len(s))
          │
          ▼
  [ Cluster Grouping ]   ◄── Combines base rune + unicode.Mn/Me combining marks without string(rs[i:]) tail allocs
          │
          ▼
   [ Canvas.set() ]      ◄── Fast-path for 1-byte ASCII (0x20..0x7e) bypasses textClusters()
          │
          ▼
 [ Terminal Canvas Grid ]
```

## Benchmark Evidence

### `ParseANSI` Benchmark Delta
- **Before:** `506,688 ns/op`, `222,400 B/op`, `4,721 allocs/op`
- **After:** `10,906 ns/op`, `5,104 B/op`, `99 allocs/op`
- **Improvement:** **~46x faster execution**, **97.7% reduction in allocated memory**, **97.9% reduction in allocations**.

### `Canvas.WriteANSI` Benchmark Delta
- **Before:** `520,890 ns/op`, `226,272 B/op`, `5,121 allocs/op`
- **After:** `22,042 ns/op`, `5,745 B/op`, `179 allocs/op`
- **Improvement:** **~23.6x faster execution**, **97.4% reduction in allocated memory**, **96.5% reduction in allocations**.
