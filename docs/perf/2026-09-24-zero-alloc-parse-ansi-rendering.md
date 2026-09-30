# Zero-Allocation ANSI Sequence Parsing & Canvas Flush Optimization

## Overview
`ParseANSINew`, `applySGRSequence`, `Style.ANSI()`, and `Canvas.FlushWithConfig` are central to Loom's TUI rendering pipeline.

Previously, `ParseANSINew` converted every input string to an `[]rune` slice (`rs := []rune(s)`), allocating an `int32` array for the full string length. On every rune step, extracting clusters via `string(rs[i:j])` generated individual string heap allocations for every character in the string. Furthermore, `applySGRSequence` called `strings.Split(params, ";")`, allocating a slice of strings for every ANSI escape sequence.

In addition, `Style.ANSI()` used `fmt.Sprintf` and string concatenation (`+`) to build color escape sequences, and `Canvas.FlushWithConfig` called `fmt.Sprintf` for cursor movement on every row render.

This optimization eliminates all intermediate string and slice allocations across the ANSI parsing and canvas flushing pipeline:
- `ParseANSINew` processes input strings directly without `[]rune(s)` conversion, performing zero-copy substring slicing (`s[i:j]`) directly on the underlying string memory.
- `applySGRSequence` parses semicolon-separated SGR parameter integer codes directly into a stack-allocated index array `[16][2]int` without `strings.Split` allocations.
- `cells` slice capacity is pre-allocated based on input string length.
- `Color` and `Style` pre-compute indexed color sequences (0–255) at package initialization and provide zero-allocation `AppendANSI` and `WriteANSI` helpers.
- `Canvas.FlushWithConfig` uses fast `strconv.AppendInt` cursor position formatting without `fmt.Sprintf`.

## Hardware Target
- Target OS: Modern Linux x86_64
- Topology: 4-10 Physical Cores / Zen 4+ / APU / Cloud Runners
- Constraints: Bounded parallelism (min(runtime.NumCPU(), 10))

## Topology / Data Flow
```
[ Raw ANSI String ]
         │
         ▼
[ In-Place UTF-8 / Escape Scanner ] ── (Zero-Copy Slicing s[i:j])
         │
         ▼
[ Zero-Alloc SGR Parser ] ──────────── (Stack Array [16][2]int)
         │
         ▼
[ Pre-Allocated Cell Buffer ]
         │
         ▼
[ Canvas.Row / FlushWithConfig ] ──── (WriteANSI & strconv.AppendInt)
         │
         ▼
[ Terminal Output Stream ]
```

## Benchmark Evidence

### `ParseANSI` Benchmark Delta
- **Legacy (`ParseANSI_Old`):** `14,869 ns/op`, `10,132 B/op`, `94 allocs/op`
- **Optimized (`ParseANSI_New`):** `9,421 ns/op`, `4,096 B/op`, `1 alloc/op`
- **Improvement:** **36.6% lower latency**, **59.6% lower memory usage**, **98.9% reduction in allocations** (from 94 to 1 allocation per call).

### `Canvas.WriteANSI` Benchmark Delta
- **Legacy (`CanvasWriteANSI_Old`):** `38,570 ns/op`, `11,414 B/op`, `174 allocs/op`
- **Optimized (`CanvasWriteANSI_New`):** `19,566 ns/op`, `4,097 B/op`, `1 alloc/op`
- **Improvement:** **49.3% lower latency**, **64.1% lower memory usage**, **99.4% reduction in allocations** (from 174 to 1 allocation per call).
