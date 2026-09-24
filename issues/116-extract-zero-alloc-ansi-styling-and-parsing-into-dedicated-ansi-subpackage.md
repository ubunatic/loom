# 116 — Extract zero-alloc ANSI styling and parsing into dedicated ansi subpackage

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Refactor
**Related**: `style.go`, `canvas.go`, `parse_ansi.go`, GitHub PR #6

## Goal

Create a dedicated zero-dependency `codeberg.org/ubunatic/loom/ansi` subpackage that encapsulates low-level ANSI SGR escape sequence generation, pre-computed 256-color lookup tables, 24-bit RGB formatting, stack-buffered zero-allocation styling, and SGR parameter parsing. Integrate the subpackage into `loom` using backwards-compatible type aliases (`loom.Style = ansi.Style`, `loom.Color = ansi.Color`), and optimize `Canvas.Row` to eliminate GC allocation overhead during frame rendering (incorporating the learnings and benchmarks from PR #6).

## Acceptance

- A dedicated `codeberg.org/ubunatic/loom/ansi` package is created alongside `measure/`, `layout/`, and `graph/`.
- `ansi` provides zero-allocation buffer appending methods (`(s Style) AppendANSI(b []byte) []byte`, `(c Color) AppendFG(b []byte) []byte`, `(c Color) AppendBG(b []byte) []byte`), precomputed 256-color index tables, and stack-buffered `Style.ANSI()`.
- SGR parsing logic in `parse_ansi.go` is cleanly extracted or backed by `ansi`.
- Root package `loom` re-exports `Style`, `Color`, `Reset`, `ColorReset`, `ColorIndex`, and `ColorRGB` as seamless aliases, preserving 100% public API compatibility.
- `Canvas.Row` in `canvas.go` uses `AppendANSI` into pre-allocated row buffers.
- Benchmarks demonstrate parity with PR #6 results (~99% reduction in `Canvas.Row` allocations and ~80% reduction in `Style.ANSI` latency).
- Existing unit, integration, and PTY tests pass without regressions.
