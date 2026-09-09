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

Copy the bounded dependency-free Harnez rograph implementation into Loom initially. Inspected sibling revision: 01e59b331c9d85699e55fbbb946c579e19901083. [bar.go](../../harnez/internal/rograph/bar.go), [sparkline.go](../../harnez/internal/rograph/sparkline.go) and [options.go](../../harnez/internal/rograph/options.go) expose RenderBar, PercentSparkline and RenderSparkline. The checked files have no SPDX headers; the checked sibling root has no tracked LICENSE/REUSE file, and README/go.mod supplied no license attribution. Loom's [REUSE.toml](../REUSE.toml) and [005](005-license-clarification.md) establish Loom's AGPL intent but do not establish permission for sibling code.

## Acceptance criteria

- [ ] Before copying, verify and record source revision, file list, authorship, applicable source license/permission and required notices. If authoritative permission is absent, record the precise blocker and resolve it before copying; do not infer it from common ownership or Loom's license.
- [ ] Copy only the required graph implementation and relevant tests into Loom with attribution and a recorded adaptation diff. Do not import another module's internal package, extract a shared package, or modify Harnez/Voxi.
- [ ] Adapt bars and percentage timelines to Loom's cells/styles and declared presentation inputs. Keep metric names, collectors and spec loading outside the renderer; avoid importing Harnez's mutable DefaultBackgroundANSI as a new Loom global.
- [ ] Provide exact allocated graph widths even for empty/short histories. Current PercentSparkline emits one glyph per sample; RenderSparkline emits up to Width from newest 2*Width samples, not automatic full-width padding. Specify padding independently of sample storage and test both semantics.
- [ ] Render fixed Harnez snapshots: paired usage bars, CPU/RAM/GPU timelines and adjacent independent VRAM/GTT timelines with separate values. Budget wrappers explicitly and shrink or omit predictably at tiny sizes.
- [ ] Retain relevant clamping and sub-character precision behavior; pass the geometry milestone in monochrome before configurable colors.

## Verification

Record provenance/permission evidence and verify imported file checksums or source diff. Port focused graph tests for boundaries, empty/short/full/overflow histories, two-sample cells and wrapper widths; run independent wide/slim geometry goldens and isolated vet/tests. Reference [Harnez 198](../../harnez/issues/198-load-box-sparkline-width-and-vram-gtt-split-restoration.md).

## Scope limits

No source copy until licensing is established; no sibling mutation, public-package extraction, Harnez collectors or palette framework. Any missing license is an implementation gate, not a reason to omit this roadmap ticket.
