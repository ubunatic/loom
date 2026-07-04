<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 003 — Audit and document the public API surface

**Status:** Done (2026-07-04)

**Priority:** P2 — do before v0.1.0 so the first tag doesn't lock in accidents

## Problem

loom was internal to uzu, where "exported" only meant "visible to uzu". As a
standalone module, every exported identifier becomes a public contract. Some
symbols exported for uzu's convenience may not belong in the public API, and
some genuinely useful helpers (see 002) are still private in uzu.

## Tasks

- [x] List all exported identifiers (`go doc .`). Surface reviewed; the widget
  constructors, layout primitives, event/style types, and YAML entry points are
  the intended public API.
- [x] Demote internal implementation detail. `Render` and `ScreenshotScript` are
  consumed by the external `loom_test` package, so they stay exported (per the
  note below); `ScreenshotScript` is now tagged as a testing/debugging utility in
  its doc comment. No other exported symbol was an accidental leak — the
  low-level bits already unexported (`cmdBar`, `helpWidget`, `colorMode`, the
  pane internals) are correctly hidden.
- [x] Confirm the intended public widgets are all constructible and documented:
  `Choice`, `Table`, `Confirm`, `TextInput`, `TextArea`, `Settings`, `Pane`,
  `Canvas` all have `New*` constructors with doc comments, plus the YAML entry
  points (`BuildWidget`/`ParseYAML`).
- [x] Ensure every exported type/func has a doc comment. Audited: all exported
  types/funcs are documented; grouped `const` blocks are covered by their type
  doc + per-const inline comments (idiomatic iota heads).
- [x] Consider whether `.loom.yaml` parsing should be a subpackage (`loom/yaml`).
  **Decided: keep top-level for v0.1.0.** The entry points are already
  name-namespaced (`ParseYAML`, `BuildWidget`, `ValidateYAML`) and moving them is
  a breaking change with little payoff. Revisit only if the YAML surface grows.

## Notes

- Test-only helpers that must stay exported for external `loom_test` packages
  (e.g. screenshot comparison) are fine to keep, but tag them clearly as
  testing utilities in doc comments.
- This is the last cheap moment to move things — after v0.1.0, renames are
  breaking changes for uman and uzu.
