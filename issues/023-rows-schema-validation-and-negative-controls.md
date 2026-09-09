<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 023 — Add negative schema controls and column-value validation for box rows

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Schema / Validation
**Related**: [011](011-aligned-dashboard-rows-and-ansi-safe-truncation.md), [Spec conventions](../docs/Spec.md), [monitor.schema.json](../spec/schemas/monitor.schema.json), [validate-spec.py](../scripts/validate-spec.py)
**Roadmap stage**: 6 spec integrity
**Depends on**: None

## Problem and findings

With the addition of declarative `rows` in commit `c19f226`, schema definitions were added to [`spec/schemas/monitor.schema.json`](../spec/schemas/monitor.schema.json). However, verification controls in [`scripts/validate-spec.py`](../scripts/validate-spec.py) were not updated to include negative tests for rows:

1. `scripts/validate-spec.py` tests negative controls for box layout (`negative_width`, `unknown_box_field`, `missing_id`, `ambiguous_root`), but does not probe `rows` schema bounds.
2. Typos in row specifications (such as `align: center`, negative `gap`, `width: 0`, or unexpected extra properties in `columns`) are not validated via negative controls in CI.
3. As documented in [Spec.md §1 & §4](../docs/Spec.md), YAML specs must be validated strictly against schemas to prevent silent drift.

## Acceptance criteria

- [ ] Extend [`scripts/validate-spec.py`](../scripts/validate-spec.py) with negative controls for `rows`:
  - invalid alignment enum (e.g. `align: center`),
  - non-positive column width (`width: 0` or negative),
  - negative gap (`gap: -1`),
  - unknown fields under `rows` and `columns` items (confirming `additionalProperties: false`).
- [ ] Confirm that `monitor.schema.json` rejects these configurations independently of Go code.
- [ ] Ensure `make validate-spec` passes all positive and negative assertions.

## Verification

Run `make validate-spec` (and `python3 scripts/validate-spec.py`) and verify that all negative tests pass and report validation status.

## Scope limits

Limited to JSON Schema definitions in `spec/schemas/` and Python validation in `scripts/validate-spec.py`. Does not change Go runtime behavior.
