<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 012 — Copy Harnez rograph with verified provenance and bounded adapters

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 6 — graphs
**Depends on**: [011](011-aligned-dashboard-rows-and-ansi-safe-truncation.md) (transitive prerequisites apply).

## Problem and findings

The user explicitly authorized copying Harnez rograph into Loom while retaining it in Harnez; no further copy permission is needed. Inspected sibling revision: 01e59b331c9d85699e55fbbb946c579e19901083. [bar.go](../../harnez/internal/rograph/bar.go), [sparkline.go](../../harnez/internal/rograph/sparkline.go) and [options.go](../../harnez/internal/rograph/options.go) expose RenderBar, PercentSparkline and RenderSparkline. The checked files have no SPDX headers; the checked sibling root has no tracked LICENSE/REUSE file, and README/go.mod supplied no license attribution. Record that metadata gap accurately alongside source provenance and the existing authorization; preserve applicable notices. Loom's licensing context is in [REUSE.toml](../REUSE.toml) and [005](005-license-clarification.md).

## Acceptance criteria

- [ ] Record the existing user authorization, source revision, copied file list and verified authorship/license metadata; preserve applicable notices and identify missing metadata without inventing attribution or requiring renewed permission.
- [ ] Copy only the required graph implementation and relevant tests into Loom with attribution and a recorded adaptation diff. Do not import another module's internal package, extract a shared package, or modify Harnez/Voxi.
- [ ] Adapt bars and percentage timelines to Loom's cells/styles and declared presentation inputs. Keep metric names, collectors and spec loading outside the renderer; avoid importing Harnez's mutable DefaultBackgroundANSI as a new Loom global.
- [ ] Provide exact allocated graph widths even for empty/short histories. Current PercentSparkline emits one glyph per sample; RenderSparkline emits up to Width from newest 2*Width samples, not automatic full-width padding. Specify padding independently of sample storage and test both semantics.
- [ ] Render fixed Harnez snapshots: paired usage bars, CPU/RAM/GPU timelines and adjacent independent VRAM/GTT timelines with separate values. Budget wrappers explicitly and shrink or omit predictably at tiny sizes.
- [ ] Retain relevant clamping and sub-character precision behavior; pass the geometry milestone in monochrome before configurable colors.

## Verification

Record provenance/permission evidence and verify imported file checksums or source diff. Port focused graph tests for boundaries, empty/short/full/overflow histories, two-sample cells and wrapper widths; run independent wide/slim geometry goldens and isolated vet/tests. Reference [Harnez 198](../../harnez/issues/198-load-box-sparkline-width-and-vram-gtt-split-restoration.md).

## Audit — 2026-09-10

Remains Open after inspecting implementation and tests at Loom `6b80ced`:

- `714bc97` and `e944324` shipped graph primitives and numerical monitor
  integration. [Graph tests](../graph/graph_test.go) assert clamping,
  sub-character precision, short/empty padding, wrappers and Braille pairs.
  [Monitor tests](../examples/monitor/main_test.go) cover rendered numerical
  bars and fixed single/split timeline widths. These are verified slices.
- [applySnapshot](../examples/monitor/main.go) updates only the first usage
  bar (`values[row][1]`); the second bar remains a literal `[    ]` in
  [monitor.yaml](../examples/monitor/spec/monitor.yaml). Paired numerical
  usage bars and their target evidence are incomplete. Generic tiny-terminal
  geometry tests use an injected child; they do not prove full graph-target
  wrapper/paired-row behavior at tiny widths.
- [Provenance](../graph/PROVENANCE.md) records the source revision, file list
  and adaptation summary, but asserts upstream AGPL without recording this
  ticket's authorization and missing-metadata finding. Read-only inspection
  of Harnez `01e59b331c9d85699e55fbbb946c579e19901083` confirms no root
  LICENSE/REUSE file and no license notice in `internal/rograph/bar.go`.
  Its graph history identifies Uwe Jugel as a commit author; that alone does
  not establish a source license. Reconcile the provenance record and retain
  source-diff/checksum evidence before closure; no renewed permission needed.
- Fresh `GOWORK=off make test`, `GOWORK=off go test -race ./...` and
  `GOWORK=off make watch-pty` pass. Passing children and general geometry
  checks do not discharge the remaining parent criteria.

## Scope limits

Copy within the existing user authorization and document provenance accurately. No sibling mutation, public-package extraction, Harnez collectors or palette framework; missing source metadata is an attribution finding, not an automatic approval gate.
