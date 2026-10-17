# 324 — respect spec: move widget chrome glyphs and sort indicators into spec

**Status**: Closed — Table sort glyphs and hint plus Choice placeholder load from validated defaults YAML; render and loader regression tests pass in make test-q1; make build install passed; fixed in the table chrome spec commit
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: `docs/Spec.md`, `spec/defaults.yaml`, `spec/schemas/defaults.schema.json`, `defaults.go`, `table.go`, `choice.go`, `searchbar.go`

---

## 1. Problem & Motivation

In accordance with `docs/Spec.md` and the `/respect` practice, chrome glyphs, indicators, and static placeholder text belong in the embedded YAML specification (`spec/defaults.yaml`) validated by JSON Schema (`spec/schemas/defaults.schema.json`).

While several widgets (like Paginator, Toggle, Tree, and Split) have had their glyphs migrated to `spec/defaults.yaml`, Table sort direction indicators (`▲` and `▼`) and default controls hint templates remain hardcoded in `table.go`. Additionally, `choice.go` hardcodes `DefaultPlaceholder = "Filter..."`, duplicating the concept in `search_bar.placeholder`.

## 2. Technical Specification / Candidates

### Candidate Values

| Location (file:line) | Value | Target Spec File and Key |
|---|---|---|
| `table.go:182-185, 293-296` | Sort indicators: `"▲"` (asc) and `"▼"` (desc) | `spec/defaults.yaml` -> `table.sort_asc_glyph`, `table.sort_desc_glyph` |
| `table.go:186` | Controls hint template: `"[tab:%s%s !:flip]"` | `spec/defaults.yaml` -> `table.controls_hint_fmt` |
| `choice.go:16` | `DefaultPlaceholder = "Filter..."` | `spec/defaults.yaml` -> `choice.default_placeholder` (or unify with `search_bar.placeholder`) |

### What Stays in Code and Why

- Dynamic formatting logic (e.g. `fmt.Sprintf` assembling column names and widths) remains in Go.
- Alignment computations and cell padding arithmetic remain in Go.
- Cell and column measurement logic stays in Go.

## 3. Acceptance Criteria

- `spec/defaults.yaml` declares `table` defaults (glyphs, hint format) and updates `spec/schemas/defaults.schema.json`.
- `table.go` and `choice.go` consume `SpeccedDefaults` instead of hardcoded strings.
- No redundant fallback literals mirror the spec values in Go code.
- Unit tests verify that the spec keys are loaded into `SpeccedDefaults` and applied during rendering.
