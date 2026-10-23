# 380 — Extract specdefaults configuration layer from root package

**Status**: Closed — extracted standalone specdefaults configuration layer with root aliases and SpeccedDefaults forwarding (commit 239c62b)
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Refactor
**Related**: 377 (modularization roadmap), `defaults.go`, `spec/defaults.yaml`

---

## 1. Problem & Motivation
Currently `defaults.go` (~900 lines) lives in the root `loom` package, parsing `spec/defaults.yaml` and defining `SpeccedDefaults` along with dozens of configuration structs (`LibDefaults`, `PaneDefaults`, `ChoiceDefaults`, etc.).
Because almost every subpackage (`internal/editor`, `internal/session`, `collector`, `graph`, `cmd/loom`) imports `loom` just to read `SpeccedDefaults`, touching any default or spec file invalidates and rebuilds the root package and all downstream consumers.

## 2. Technical Specification / Findings
- Extract the defaults structs, spec loading, and unmarshaling into a clean subpackage `ubunatic.com/loom/specdefaults` (or promoting `internal/specdefaults`).
- In root package `loom`:
  - Keep full backwards compatibility by re-exporting `SpeccedDefaults = specdefaults.SpeccedDefaults`.
  - Type alias the defaults structs (`type LibDefaults = specdefaults.LibDefaults`, etc.).
- Subpackages (`internal/editor`, `internal/session`, etc.) can then import `specdefaults` directly where appropriate without forcing full root `loom` compilation.
- Ensure `loom` public API stays 100% compatible.

## 3. Implementation & Verification Plan
1. Move/extract defaults definitions and unmarshaling logic to `specdefaults/`.
2. Re-export type aliases and `SpeccedDefaults` in root `loom`.
3. Verify public API with `internal/apitest`.
4. Run `go test $(TEST_PARALLEL) ./...`.
5. Ensure `make install` passes.
