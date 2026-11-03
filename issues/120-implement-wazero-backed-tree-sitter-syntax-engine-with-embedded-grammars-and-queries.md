# 120 — Implement wazero-backed Tree-Sitter syntax engine with embedded grammars and queries

**Status**: Open
**Priority**: P2
**Severity**: Moderate
**Category**: Feature / Parser Engine
**Related**: `#118`, `#119`, `#121`

---

## Goal

`/goal`: Implement the `syntax.Engine` interface using pure-Go `wazero` running Tree-Sitter with embedded Wasm grammars (Go, Markdown, YAML, JSON) and Scheme highlight queries (`highlights.scm`).

## 1. Context & Motivation

Following the successful spike in #118 and interface definition in #119, this issue builds the concrete production engine that embeds pre-compiled WebAssembly grammar assets and runs S-expression pattern queries to produce syntax highlight spans.

## 2. Technical Specification

1. **Embedded Assets**:
   - Store compiled `.wasm` grammars under `syntax/grammars/` (or a dedicated assets subpackage).
   - Embed standard `highlights.scm` queries for supported languages (Go, Markdown, JSON, YAML).
2. **Engine Implementation (`wazeroEngine`)**:
   - Manages a shared `wazero.Runtime` with compiled module caching.
   - Per-document parser instance maintaining the active CST (`ts_tree`).
   - Translates `syntax.Edit` structs into Tree-Sitter incremental edit calls (`ts_tree_edit`).
   - Executes highlight queries over the CST and collects capture ranges matching requested viewport line numbers.
3. **Performance & Memory Optimization**:
   - Reuse query cursors and match buffers across frames.
   - Cache line-level highlight spans until invalidating edits occur on or before those lines.
   - Keep per-keystroke re-parse + viewport query time under 2ms.

## 3. Acceptance Criteria

- `wazeroEngine` passes conformance tests for Go, Markdown, JSON, and YAML source snippets.
- Incremental parsing correctly reflects single-character insertions, deletions, and multi-line pastes.
- Benchmark suite verifies allocation count and latency metrics meet interactive editing requirements.
- Zero CGO dependencies (`CGO_ENABLED=0 go build` succeeds).

## Jules experiment (2026-11-01): wazero + tree-sitter-go

Inconclusive (lead, not host-reproduced): `tree-sitter build --wasm` gives a 217KB grammar module that imports Emscripten `env` (memory etc.) and needs the tree-sitter core engine in the same module, so a bare wazero host fails with `module[env] not instantiated`. Next step would be one wasm build of core + grammar (e.g. along the lines of malivvan/tree-sitter). No startup/size/highlight numbers yet. Report: `reports/jules/120-wazero.md`.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| audit | jules-audit-120 | d15cf1f5-6c2c-48d7-899c-7b8f7d10d41a | - | running |
