<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 003 — Audit and document the public API surface

**Status:** Open

**Priority:** P2 — do before v0.1.0 so the first tag doesn't lock in accidents

## Problem

loom was internal to uzu, where "exported" only meant "visible to uzu". As a
standalone module, every exported identifier becomes a public contract. Some
symbols exported for uzu's convenience may not belong in the public API, and
some genuinely useful helpers (see 002) are still private in uzu.

## Tasks

- [ ] List all exported identifiers (`go doc ./...` / `golang.org/x/tools/cmd/…`).
- [ ] Demote anything that is an internal implementation detail to unexported
  (screenshot/test scaffolding especially — `screenshot.go`, render helpers).
- [ ] Confirm the intended public widgets are all constructible and documented:
  `Choice`, `Table`, `Confirm`, `TextInput`, `TextArea`, `Settings`, `Pane`,
  `Canvas`, plus the YAML layout entry points (`BuildWidget`/`ParseYAML`).
- [ ] Ensure every exported type/func has a doc comment.
- [ ] Consider whether `.loom.yaml` parsing should be a subpackage
  (`loom/yaml`) rather than top-level.

## Notes

- Test-only helpers that must stay exported for external `loom_test` packages
  (e.g. screenshot comparison) are fine to keep, but tag them clearly as
  testing utilities in doc comments.
- This is the last cheap moment to move things — after v0.1.0, renames are
  breaking changes for uman and uzu.
