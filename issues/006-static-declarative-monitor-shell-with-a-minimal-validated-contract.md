<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 006 — Static declarative monitor shell with a minimal validated contract

**Status**: Closed — static declarative shell implemented and verified
**Priority**: P1 (High)
**Severity**: Moderate
**Category**: Feature
**Related**: [Roadmap](../docs/Roadmap.md), [context](../docs/RoadmapContext.md), [targets](../docs/HarnezUsageTarget.md), [spec conventions](../docs/Spec.md)
**Roadmap stage**: 1
**Depends on**: None; first roadmap implementation ticket.

## Problem and findings

Loom needs its first runnable declarative monitor, with a minimal contract proven by the shell itself. [yaml.go](../yaml.go) already has BuildWidget, ValidateYAML and RootView, but BuildWidget compiles pane/views without cfg.View, and no-grid views iterate a map. Validation can therefore accept a root that does not render and declaration order is unstable. [Stack](../stack.go) divides space equally; it is not application chrome. There is no implemented spec/ tree or generic box.

## Acceptance criteria

- [x] Provide a documented runnable example with a minimal Go entrypoint: show once, print one frame and exit without waiting for input or requiring a controlling TTY. The frame contains a title bar, exactly two empty titled bordered boxes, and a bottom status bar.
- [x] Implement only the generic box/chrome composition needed here. Declare ordered children, titles, spacing, padding and initial dimensions in embedded YAML; Go does not duplicate those values or place application widgets at coordinates.
- [x] Add small baseline checks for intact borders, declared padding, child-content clipping and tiny bounds. These protect the first shell; the broader independent final-display oracle and visual gate follow in 010.
- [x] Include the smallest consumed spec/declaration contract and companion JSON Schema under spec/ in this ticket, with a schema reference, embedding, and automated schema validation. Reject unknown fields, invalid dimensions, duplicate identifiers and unresolved references with useful locations. Do not create an unused actions/source registry.
- [x] Define deterministic child order explicitly, not through Go map iteration. Make validation and construction agree on the accepted root forms (pane, view, views), root precedence and missing/ambiguous roots; cover existing forms without silently dropping cfg.View.
- [x] Retain existing widget/YAML behavior except documented, tested corrections. Add one small sample and its exact invocation; no mandatory new nested module or invented public runtime API.

## Verification

Run GOWORK=off go vet ./... then GOWORK=off go test ./.... Add positive/negative schema and root-form cases, repeated-build order checks, and a deterministic headless shell golden using Render. Execute the documented one-shot command with redirected stdout and no TTY. Confirm that changing a declared title/order changes output without a Go edit.

## Scope limits

No watch loop, clock, visibility actions, responsive breakpoints, graphs, data bindings, general schema compiler or wholesale YAML redesign. This is the prerequisite contract within a visible deliverable, not a separate framework project.

## Implementation evidence

- [Example and exact invocation](../examples/monitor/README.md): `GOWORK=off go run ./examples/monitor`, with embedded `spec/monitor.yaml` in the root module.
- [Frame/Box](../frame.go) isolate child canvases; [baseline tests](../frame_test.go) check exact rows, borders, padding, tiny bounds, invalid declarations and YAML-only title/order changes.
- [Root/order tests](../yaml_contract_test.go) exercise all three forms, deterministic legacy ordering, duplicate IDs, missing references and matching validation/build errors.
- [Schemas](../spec/schemas/monitor.schema.json) validate the initial shell vocabulary through `make validate-spec`; legacy widgets retain Go validation. Border presentation is consumed from embedded `spec/box.yaml`.
- Verified `GOWORK=off make test` (standard JSON Schema checks with negative controls, Go vet, Go tests), standalone `sh scripts/check-no-tty.sh`, and `setsid --wait env GOWORK=off go run ./examples/monitor </dev/null`.
- Intentional limits: fixed declared dimensions, printable ASCII shell labels, spec-defined border glyphs; unknown YAML fields and ambiguous roots now fail explicitly. Full Unicode/ANSI geometry and human visual acceptance remain ticket 010.
