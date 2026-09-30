# 225 — Modularize high-LOC core components and example packages

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Refactor
**Related**: `frame.go`, `ansibuffer.go`, `yaml.go`, `graph/treemap.go`, `graph/options.go`, `examples/loomoji/loomoji/loomoji.go`

## Goal

Decompose monolithic multi-responsibility files flagged during repo assessment (>500 LOC) into focused, highly cohesive files without breaking public API contracts or widget semantics.

## Problem & Motivation

`harnez assess` identified several core components combining unrelated subsystems into single files:
1. `frame.go` (1,124 LOC): mixes leaf `Box` container/border styling (~270 LOC) with `Frame` dynamic layout orchestration and focus cycling (~850 LOC).
2. `ansibuffer.go` (956 LOC): embeds box border character detection and validation (`ValidateAnsiBox`) alongside `AnsiBuffer` text streaming and serialization.
3. `yaml.go` (786 LOC): bundles `Router` navigation widget (~160 LOC), `ParseASCIIGrid` template parser (~135 LOC), and YAML declarative compiler (~490 LOC).
4. `graph/treemap.go` (1,305 LOC) & `graph/options.go` (543 LOC): combine color scales, squarified treemap algorithms, legend formatting, and bar/sparkline rendering.
5. `examples/loomoji/loomoji/loomoji.go` (1,604 LOC): contains 1,160 lines of static emoji table data in the main application file.

## Acceptance

- `Box`, `BoxBorder`, and border styling parsed/rendered extracted from `frame.go` into `box.go`; `Frame` layout retained in `frame.go`.
- `ValidateAnsiBox` and box character helpers moved from `ansibuffer.go` into `ansibox.go` (preparing for issue #147).
- `Router` extracted to `router.go` and ASCII grid parser extracted to `grid_ascii.go` from `yaml.go`.
- Subpackage `graph/` modularized into `treemap_layout.go`, `treemap_legend.go`, `colorscale.go`, `bar.go`, and `sparkline.go`.
- Static emoji data in `examples/loomoji` moved to `entries.go`.
- 100% backwards compatibility maintained for public APIs, tests, and examples.
- `make test` passes without regressions.
