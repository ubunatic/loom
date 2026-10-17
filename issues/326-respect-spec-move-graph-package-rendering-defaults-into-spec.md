# 326 — respect spec: move graph package rendering defaults into spec

**Status**: Closed — Graph widths, ANSI defaults, glyph tables, and spinner frames now load from schema-validated spec/defaults.yaml; make check and make container-test pass (a8f7d57).
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: `docs/Spec.md`, `spec/defaults.yaml`, `spec/schemas/defaults.schema.json`, `graph/options.go`, `graph/bar.go`, `graph/sparkline.go`, `graph/stackedbar.go`, `graph/treemap.go`, `graph/spinner.go`

---

## 1. Problem & Motivation

Under `docs/Spec.md` and the `/respect` practice, presentation defaults such as fallback dimensions, default glyph sequences, spinner frames, and background ANSI sequences belong in a declarative spec rather than as hardcoded package-level constants and variables.

The `graph/` package defines several rendering defaults directly in Go code:
- `graph/options.go`: `DefaultBackgroundANSI = "100"`
- `graph/bar.go`: `MaxWidth = 10`
- `graph/sparkline.go`: `percentSparkChars = []rune("▁▂▃▄▅▆▇█")`
- `graph/stackedbar.go`: `defaultStackedBarGlyphs = []rune("█▓▒░")`
- `graph/treemap.go`: `defaultTreemapGlyphs = []rune("█▓▒░")`
- `graph/spinner.go`: `DefaultSpinnerFrames = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}`

## 2. Technical Specification / Candidates

### Candidate Values

| Location (file:line) | Value | Target Spec File and Key |
|---|---|---|
| `graph/options.go:12` | `DefaultBackgroundANSI = "100"` | `spec/defaults.yaml` -> `graph.default_background_ansi` |
| `graph/bar.go:13` | `MaxWidth = 10` | `spec/defaults.yaml` -> `graph.bar.max_width` |
| `graph/sparkline.go:13` | `[]rune("▁▂▃▄▅▆▇█")` | `spec/defaults.yaml` -> `graph.sparkline.glyphs` |
| `graph/stackedbar.go:14` | `[]rune("█▓▒░")` | `spec/defaults.yaml` -> `graph.stacked_bar.glyphs` |
| `graph/treemap.go:14` | `[]rune("█▓▒░")` | `spec/defaults.yaml` -> `graph.treemap.glyphs` |
| `graph/spinner.go:12` | `DefaultSpinnerFrames = []rune{'⠋', ...}` | `spec/defaults.yaml` -> `graph.spinner.frames` |

### What Stays in Code and Why

- Graph clipping and quantization algorithms (e.g. eighth-block sub-character math) remain in Go.
- Layout division and squarified treemap aspect-ratio calculations stay in Go.
- ANSI escape string generation mechanics stay in Go.

## 3. Acceptance Criteria

- `spec/defaults.yaml` (or a dedicated `spec/graph.yaml`) defines graph rendering defaults and glyph tables, covered by JSON Schema.
- The `graph` package consumes these values via specced loaders instead of hardcoded package constants.
- No duplicate fallback glyph tables or constants are retained in Go code.
- Unit tests verify that graph functions produce expected outputs using the specced defaults.
