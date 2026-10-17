# 327 — respect spec: move collector retention and resize mode ordering into spec

**Status**: Closed — Collector retention and resize order load from validated embedded YAML; eviction, ordering and negative-schema regressions pass; make test-q1, make check and make build install passed; fixed in e7610b6
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: `docs/Spec.md`, `spec/defaults.yaml`, `spec/resize.yaml`, `collector/collector.go`, `resize.go`

---

## 1. Problem & Motivation

Under `docs/Spec.md` and the `/respect` practice, durable system defaults and presentation ordering belong in the YAML specification files rather than in Go code.

Currently:
- `collector/collector.go` defines `DefaultRetention = 15 * time.Minute` as a hardcoded Go constant instead of deriving it from spec.
- `resize.go` maintains `SpeccedResizeModeIDs = []string{...}` as a hardcoded Go slice to enforce presentation ordering because `spec/resize.yaml` models modes as an unordered map. This creates a dual source of truth where adding or renaming a mode requires modifying Go code and breaks pure declarative schema validation.

## 2. Technical Specification / Candidates

### Candidate Values

| Location (file:line) | Value | Target Spec File and Key |
|---|---|---|
| `collector/collector.go:24` | `DefaultRetention = 15 * time.Minute` | `spec/defaults.yaml` -> `collector.default_retention` |
| `resize.go:55-67` | `SpeccedResizeModeIDs = []string{"coalesce", "atomic_flush", ...}` | `spec/resize.yaml` -> `mode_order: [...]` (or ordered sequence in YAML) |

### What Stays in Code and Why

- Collection concurrency channels, mutex locking, and I/O polling mechanics stay in Go.
- Terminal signal handling (`SIGWINCH`), window measurement, and ANSI buffer clear implementations stay in Go.

## 3. Acceptance Criteria

- `spec/defaults.yaml` (or collector spec) declares `default_retention`, with JSON schema updates.
- `spec/resize.yaml` specifies the presentation order of modes, and `spec/schemas/resize.schema.json` validates it.
- Go code loads `DefaultRetention` and `SpeccedResizeModeIDs` from the embedded specs without hardcoded mirror arrays.
- Unit tests assert that collector defaults and resize mode lists match the spec declarations.
