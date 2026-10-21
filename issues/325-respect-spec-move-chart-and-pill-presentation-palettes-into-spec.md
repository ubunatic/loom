# 325 — respect spec: move chart and pill presentation palettes into spec

**Status**: Closed — Done in 514ad29: chart palette/label/axis and pill status colors are theme roles in spec/themes.yaml with schema; tests compare rendering to the spec; additive API (ThemeColors fields); q1 green after the final change
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: `docs/Spec.md`, `spec/themes.yaml`, `spec/defaults.yaml`, `spec/schemas/themes.schema.json`, `chart.go`, `pill.go`

---

## 1. Problem & Motivation

Under `docs/Spec.md` and the `/respect` practice, semantic colors and presentation palettes must reside in `spec/themes.yaml` or `spec/defaults.yaml`, rather than as hardcoded numeric literals in Go package variables.

Currently:
- `chart.go` defines `chartColors` with hardcoded ANSI 256/16 color indices `[6, 3, 5, 2, 4, 1]` and hardcoded label/axis styles (`ColorIndex(7)`, `ColorIndex(8)`).
- `pill.go` defines package-level status colors: `colorYellow = ColorIndex(3)`, `colorGreen = ColorIndex(2)`, `colorRed = ColorIndex(1)`.

## 2. Technical Specification / Candidates

### Candidate Values

| Location (file:line) | Value | Target Spec File and Key |
|---|---|---|
| `chart.go:43` | `[...]Color{ColorIndex(6), ColorIndex(3), ColorIndex(5), ColorIndex(2), ColorIndex(4), ColorIndex(1)}` | `spec/themes.yaml` -> `roles.chart_palette` or `spec/defaults.yaml` -> `chart.series_colors` |
| `chart.go:91-92` | Label `ColorIndex(7)`, Axis `ColorIndex(8)` | `spec/themes.yaml` -> theme roles (`chart_label`, `chart_axis`) |
| `pill.go:34-38` | `ColorIndex(3)`, `ColorIndex(2)`, `ColorIndex(1)` | `spec/themes.yaml` -> status roles (`status_pending`, `status_success`, `status_error`) |

### What Stays in Code and Why

- Standard ANSI/xterm 256 color palette definitions and color space conversion arithmetic remain in Go.
- Drawing algorithms, line plotting math, and interpolation mechanics in `chart.go` remain in Go.

## 3. Acceptance Criteria

- Presentation palette colors and status badge color roles are declared in YAML (`spec/themes.yaml` or `spec/defaults.yaml`) with schema coverage.
- Code in `chart.go` and `pill.go` accesses these colors via `Theme` roles or `SpeccedDefaults`.
- No hardcoded `ColorIndex(...)` literals shadow spec values in package variables.
- Unit tests assert that the palette and status colors match the values declared in the spec.
