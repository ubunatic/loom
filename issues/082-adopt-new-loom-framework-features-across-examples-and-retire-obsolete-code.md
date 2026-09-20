# 082 — Adopt new loom framework features across examples and retire obsolete code

**Status**: In Progress
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Refactor
**Related**: `docs/Widgets.md`, `docs/studies/`, `examples/`

---

## 1. Problem & Motivation

Several new high-level framework primitives have landed in Loom (`loom.Split`, dynamic `loom.Tabs`, `loom.MetricStore`, metric-bound `Gauge` / `Sparkline`, `loom.Directory` / `loom.OpenFile`, and `Pane.RunStartup`), but the reference applications in `examples/` still contain ad-hoc custom implementations, duplicated logic, or outdated patterns that predate these features.

## 2. Desired Behavior & Goal

`/goal`: Refactor all example applications in `examples/` to adopt newly introduced Loom framework features, eliminate obsolete custom boilerplate, capture lingering demo app issues, and publish adoption reports in `docs/studies/`.

- **Adoption Scope**:
  - `examples/split`: Adopt first-class `loom.Split` widget and nested focus container.
  - `examples/tabs`: Adopt dynamic `loom.Tabs` methods and declarative `TabsKeys`.
  - `examples/monitor`: Adopt `loom.MetricStore` and bound `Gauge`/`Sparkline` widgets.
  - `examples/filebrowser`: Fully leverage `loom.Directory`, `loom.OpenFile`, and path formatters.
  - `examples/splash`: Leverage `Pane.RunStartup` and `loom.RenderTo`.
  - Remaining examples (`screens`, `treemap`, `winch`, `background`): Update to use unified widget and layout patterns where applicable.
- **Hygiene & Verification**:
  - Remove all retired application-side boilerplate.
  - Maintain 100% green tests (`make test-q1`).
  - Document gaps/issues observed in each demo app.
  - Publish adoption summary studies in `docs/studies/`.

## 3. Implementation Plan

### M1 — Multi-Agent Example Audit & Gap Identification
- Launch `luna:low` agents to audit each example application against the latest Loom API surface (`docs/Widgets.md`).
- Inventory obsolete code sections and identify bugs/defects within the demo apps.

### M2 — Refactor & Code Retirement
- Refactor target example applications to use core Loom primitives.
- Delete duplicated helper functions, manual sleep loops, and custom layout arithmetic from `examples/`.

### M3 — Verification & Study Documentation
- Run full test suite (`make test-q1` / `go test ./...`).
- Author adoption and findings studies in `docs/studies/`.
