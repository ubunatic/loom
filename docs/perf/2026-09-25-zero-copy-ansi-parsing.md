# Zero-Copy ANSI Escape Sequence Parsing

## Overview
`ParseANSINew` decodes ANSI-formatted terminal text into Loom canvas `[]Cell` slices. Previously, it performed a heap-allocating `[]rune(s)` conversion for the entire input string and relied on `strings.Split` and `strconv.Atoi` inside `applySGRSequence` when decoding SGR color/style parameters. On typical ANSI-styled strings, this produced 99 allocations per call and ~10.4µs latency.

The optimization replaces `[]rune(s)` with zero-copy string indexing and `utf8.DecodeRuneInString`, scans SGR parameters using a stack-buffered `[16]int` array with `parseSGRDecimal` (eliminating `strings.Split` and `strconv.Atoi`), and introduces a fast-path for plain printable ASCII characters. This drops allocations from 99 to 1 (only the pre-allocated output `Cell` slice) and increases throughput by ~3.7x (~2.8µs latency).

## Hardware Target
- **Target Topology:** 4-10 Physical Cores, x86_64 / AVX-512 (AMD Zen 4+ / Intel Xeon).
- **RAM & Footprint:** 32GB UMA buffer sharing; optimized for throughput and GC pressure elimination in high-frequency TUI render loops.

## Topology / Data Flow

```
[ANSI Input String]
        │
        ▼
[Fast ASCII Range Check (0x20..0x7e)] ──► [Direct Cell Slice Output]
        │
        ├─► [CSI ESC [ Scanner] ──► [Stack-Buffered SGR Decoder [16]int]
        │
        └─► [utf8.DecodeRuneInString] ──► [In-Place Combining Mark Slicer]
```

## Benchmark Evidence

Command:
```bash
go test -bench=BenchmarkParseANSI|BenchmarkCanvasWriteANSI -benchmem -count=5 .
```

### `ParseANSINew` Delta

| Metric | Before | After | Change |
| :--- | :--- | :--- | :--- |
| **Latency (ns/op)** | 10,420 ns/op | 2,824 ns/op | **~3.7x faster (-72.9%)** |
| **Memory (B/op)** | 5,104 B/op | 4,096 B/op | **-19.7%** |
| **Allocations (allocs/op)** | 99 allocs/op | 1 allocs/op | **-99.0%** |

### `Canvas.WriteANSI` Delta

| Metric | Before | After | Change |
| :--- | :--- | :--- | :--- |
| **Latency (ns/op)** | 23,445 ns/op | 15,385 ns/op | **~1.52x faster (-34.4%)** |
| **Memory (B/op)** | 5,105 B/op | 4,096 B/op | **-19.8%** |
| **Allocations (allocs/op)** | 99 allocs/op | 1 allocs/op | **-99.0%** |
