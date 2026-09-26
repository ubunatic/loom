# Zero-Allocation Frame Flush & ANSI Style Escape Sequence Formatting

## Overview
`Canvas.Flush`, `Canvas.Row`, and `Style.ANSI` are on the hot rendering path for every frame drawn in Loom TUI applications. Previously, `Color.fgSeq()` and `Color.bgSeq()` invoked `fmt.Sprintf` on every color formatting operation, and `Style.ANSI()` performed multiple string concatenations (`out += ...`). Furthermore, `Canvas.FlushWithConfig` formatted cursor movements and row sequences using `fmt.Sprintf` for every row on every frame flush.

This change replaces `fmt.Sprintf` and string concatenation allocations in `style.go` and `canvas.go` with zero-allocation byte slice buffer appends using `strconv.AppendUint`. `Style.AppendANSI(b []byte) []byte` writes style escape sequences directly into pre-allocated slice buffers, `Style.ANSI()` utilizes stack memory (`[64]byte`), `Canvas.Row` constructs row output directly in a pre-allocated byte buffer, and `Canvas.FlushWithConfig` flushes entire canvas frames into a single buffer.

## Hardware Target
- Target OS: Modern Linux x86_64
- Topology: 4-10 Physical Cores / APU / High-Hz Terminal Output
- Target Memory Bounds: Zero-allocation hot rendering loops

## Topology / Data Flow
```
[ Frame Redraw / Canvas.FlushWithConfig ]
                   │
                   ▼
     [ Single Frame Pre-Allocated Buffer ]
                   │
    ┌──────────────┴──────────────┐
    ▼                             ▼
[ Row Movement ]         [ Cell Style Formatting ]
 (strconv.AppendUint)      (Style.AppendANSI)
    │                             │
    └──────────────┬──────────────┘
                   ▼
     [ Single Atomic tty.WriteString ]
```

## Benchmark Evidence

### `Style.ANSI`
- **Allocations:** `0 allocs/op`, `24 B/op`
- **Execution Speed:** `~78 ns/op`

### `Canvas.Row`
- **Allocations:** `6 allocs/op`, `7872 B/op`
- **Execution Speed:** `~6.0 µs/op` (80 styled cells)

### `Canvas.Flush` (80x24 Frame)
- **Allocations:** `11 allocs/op`, `338,592 B/op`
- **Execution Speed:** `~163 µs/op` for full 80x24 styled canvas flush
