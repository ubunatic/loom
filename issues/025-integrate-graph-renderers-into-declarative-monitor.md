<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 025 — Integrate graph renderers into declarative monitor

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [012](012-copy-harnez-rograph-with-verified-provenance-and-bounded-adapters.md), [024](024-port-harnez-rograph-primitives-with-provenance.md), [Harnez target](../docs/HarnezUsageTarget.md), [Roadmap](../docs/Roadmap.md)
**Roadmap stage**: 6 — graphs (sub-ticket of 012)
**Depends on**: [024](024-port-harnez-rograph-primitives-with-provenance.md)

## Problem and findings

With the graph primitives ported in [024](024-port-harnez-rograph-primitives-with-provenance.md), the declarative monitor currently renders static dummy string values (`[⣿⣿  ]`, `[⣠⣀⣀⣀⣀⣀⣀⣄⣀⣀]`) declared directly in YAML.
To complete the Stage 6 graph milestone:
1. Connect the adapted `RenderBar` and `PercentSparkline` renderers to the monitor example so graph columns are generated dynamically from numerical inputs rather than hardcoded string glyphs.
2. Support single rolling timelines (CPU, RAM, GPU) and composite dual timelines (split VRAM/GTT).
3. Validate that changing numerical inputs renders accurate bar fills and timeline trends while maintaining strict column alignment and box geometry.

## Acceptance criteria

- [ ] Connect the adapted graph renderer to monitor row data structures, replacing static dummy string bars with rendered outputs from numerical sample inputs.
- [ ] Support composite split timelines for VRAM/GTT rendering two independent 6-character timelines side by side.
- [ ] Preserve exact 27-column inner box layout without width drift across wide (64x10) and slim (40x20) layouts.
- [ ] Verify that empty or short histories render clean padded empty brackets (`[    ]`).
- [ ] Pass full test suite and geometry verification with `make test`.

## Verification

Run `go test ./...`, `make test`, and execute `go run ./examples/monitor` to verify live rendered graphs.

## Scope limits

Limited to static snapshot graph rendering and composite split timeline wiring. Dynamic timer-based rolling histories and live sampling loops are part of Stage 7 ([013](013-deterministic-live-snapshots-and-independent-rolling-histories.md)).
