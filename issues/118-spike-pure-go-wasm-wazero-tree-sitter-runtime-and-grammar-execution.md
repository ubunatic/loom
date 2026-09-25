# 118 — Spike pure Go / Wasm (wazero) Tree-Sitter runtime and grammar execution

**Status**: Closed — Canary executed in internal/canary/treesitter; measured wazero Tree-Sitter load, parse, and query metrics in docs/studies/2026-09-treesitter-syntax-engine.md; reviewed by terra:med
**Priority**: P1
**Severity**: Moderate
**Category**: Research / Architecture / Canary
**Related**: `#119`, `#120`, `docs/Canary.md`

---

## Goal

`/goal`: Validate the feasibility, latency, memory footprint, and ergonomics of running Tree-Sitter and compiled language grammars in a pure-Go WebAssembly runtime (`wazero`) without requiring CGO or native C toolchains.

## 1. Context & Motivation

Loom aims to provide AST-driven syntax highlighting and code intelligence for its editor components (`TextArea`, `examples/textedit`) while maintaining pure-Go zero-CGO cross-compilation.

Tree-Sitter core and its language grammars can be compiled to standard WebAssembly (`.wasm`). Running them inside `github.com/tetratelabs/wazero` (a zero-dependency, pure-Go Wasm runtime with JIT compilation on amd64/arm64 and interpreter fallback) could provide full Tree-Sitter capabilities without CGO.

## 2. Canary & Spike Tasks

1. **Wasm Runtime Setup**:
   - Evaluate `wazero` configuration (`wazero.NewRuntimeWithConfig`, memory limits, instantiation overhead).
   - Test loading pre-compiled `tree-sitter.wasm` or an embedded grammar wasm module (e.g. `tree-sitter-go.wasm`).
2. **Benchmark & Feasibility Measurements**:
   - Initial module instantiation time and memory overhead.
   - Initial parse latency on 100-line, 1,000-line, and 10,000-line Go/JSON source files.
   - Incremental re-parse latency upon single-character edits (target: < 2ms per turn).
   - Execution of S-expression highlight queries (`highlights.scm`).
3. **CGO vs Wasm Tradeoff Assessment**:
   - Document any performance gaps, memory constraints, or build ergonomics for compiling grammars to Wasm.
   - Deliver findings in a canary study document (`docs/studies/2026-09-tree-sitter-wazero-spike.md`).

## 3. Acceptance Criteria

- A runnable canary test or benchmark in `internal/canary/treesitter/` exercises tree-sitter operations under `wazero`.
- Metrics recorded for load time, incremental parse latency, and memory allocation.
- A concise findings report documenting viability, risks, and recommended module design for ticket #119 and #120.
