# 078 — Add concurrency-safe MetricStore and metric-bound Gauge and Sparkline widgets

**Status**: Closed — implemented and verified by make test-q1
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `docs/studies/2026-09-20-feature-gap-analysis-monitor.md`, `graph/`, `examples/monitor/monitor/`

---

## 1. Problem & Motivation

Loom provides low-level chart and sparkline rendering functions in `graph/`, but monitoring applications must manually manage bounded history slices, thread-safe snapshotting, and cell-string formatting in application state.

## 2. Desired Behavior & Goal

`/goal`: Provide a standard concurrency-safe `MetricStore` and higher-level `Gauge` and `Sparkline` widgets that bind directly to metric series data.

- Implement `loom.NewMetricStore(opts ...MetricStoreOption)` supporting rolling capacity/retention and immutable snapshots.
- Provide declarative `Gauge` and `Sparkline` widgets that render formatted values or series history within allocated bounds.
- Allow multi-series composition without manual string indexing in application render loops.

## 3. Implementation Plan

1. Create `metric_store.go` with thread-safe publish/snapshot semantics and retention windows.
2. Add `Gauge` and `Sparkline` widget structs in `widget_graph.go` wrapping `graph.RenderBar` and `graph.RenderSparkline`.
3. Add tests verifying race freedom, retention bounds, and rendering output across varying dimensions.
