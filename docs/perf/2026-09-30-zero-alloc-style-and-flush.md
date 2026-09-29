# Zero-Allocation Style ANSI Formatting & Canvas Frame Serialization Optimization

## Overview
`Style.ANSI()` and `Canvas.FlushWithConfig` are on the hottest rendering path in Loom during every frame redraw. Previously, `Style.ANSI()` relied on `fmt.Sprintf` calls in `Color.fgSeq()` and `Color.bgSeq()` as well as multiple string concatenations (`+`). On a modest 120x40 canvas frame, style transitions during row serialization generated over 17,000 heap allocations per frame flush, placing heavy pressure on Go garbage collection.

This optimization implements `Style.AppendANSI(dst []byte) []byte` and `Style.WriteANSI(b *strings.Builder)` using `strconv.AppendUint` to stream ANSI SGR escape sequences directly into pre-allocated buffer slices without `fmt.Sprintf` or string heap allocations. `Canvas.Row` now pre-sizes its builder and streams style sequences via `WriteANSI`, while `Canvas.FlushWithConfig` uses `strconv.Itoa` to format row positioning escape sequences.

## Hardware Target
- Target OS: Modern Linux x86_64
- Topology: 4-10 Physical Cores / Zen 4+ / APU / Cloud Runners
- Constraints: Multi-core bounds (max 10 workers/cores)

## Topology / Data Flow
```
[ Canvas Cell Matrix (cols x rows) ]
                 │
                 ▼
      [ Canvas.Row Serialization ]
                 │
                 ▼
  [ cell.Style.WriteANSI(builder) ]
                 │
        ┌────────┴────────┐
        ▼                 ▼
 [ AppendANSI ]    [ strconv.AppendUint ] (Zero-alloc SGR)
        │
        ▼
 [ Canvas.FlushWithConfig ] (strconv.Itoa row positioning)
        │
        ▼
 [ Terminal Stdout / Writer ]
```

## Benchmark Evidence

### `BenchmarkStyleANSI` Delta
- **Before:** `693.1 ns/op`, `128 B/op`, `6 allocs/op`
- **After:** `98.96 ns/op`, `48 B/op`, `1 allocs/op`
- **Improvement:** **7.0x faster execution**, **62.5% memory reduction**, **83.3% allocation reduction**.

### `BenchmarkCanvasRow` Delta
- **Before:** `101,913 ns/op`, `17,977 B/op`, `429 allocs/op`
- **After:** `14,127 ns/op`, `7,938 B/op`, `4 allocs/op`
- **Improvement:** **7.2x faster execution**, **55.8% memory reduction**, **99.1% allocation reduction** (429 -> 4 allocs/op).

### `BenchmarkCanvasFlush` Delta (120x40 Frame)
- **Before:** `2,833,102 ns/op`, `1,429,858 B/op`, `17,238 allocs/op`
- **After:** `893,575 ns/op`, `1,028,033 B/op`, `176 allocs/op`
- **Improvement:** **3.17x faster frame flush**, **28.1% memory reduction**, **99.0% allocation reduction** (17,238 -> 176 allocs/op).
