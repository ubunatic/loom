<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 004 — Migrate uzu to depend on the shared loom module

**Status:** Open

**Priority:** P2 — can land right after the v0.1.0 tag (001)

## Motivation

Once loom is tagged (001), uzu should consume the shared module instead of
carrying its own `uzu/loom` copy. This keeps a single source of truth and proves
the extraction round-trips.

## Tasks (in the uzu repo)

- [ ] Delete `uzu/loom/` (the copy now living here).
- [ ] `require codeberg.org/ubunatic/loom vX.Y.Z` in `uzu/go.mod`.
- [ ] Import paths already say `codeberg.org/ubunatic/uzu/loom` — rewrite to
  `codeberg.org/ubunatic/loom` across uzu (`main.go`, `internal/wizard/*`,
  `cmd/loomdemo`, `cmd/loomtest`).
- [ ] `go mod tidy`, `go build ./...`, `go test ./...` green in uzu.
- [ ] Move uzu's private `runPane` helper to `loom.RunPane` if 002 lands.

## Risks / notes

### Audit — 2026-09-10

Remains Open. Loom's README reports downstream consumption, but the expected
`/home/uwe/projects/uzu` checkout is absent. Imports, removal of the private
copy, dependency version, `RunPane` migration and downstream build/test results
cannot be verified here. No task checkbox is completed from the README claim.

- **Do not** delete `uzu/loom` until the tag exists and uzu builds against it —
  otherwise uzu is broken in the interim. Sequence: tag loom → point uzu at it →
  verify → delete the copy.
- uzu and loom drift: any local uzu edits to loom made after this extraction
  must be ported here first, or they are lost on deletion. As of extraction the
  two are byte-identical except import paths.
- This issue is the mirror of **uman issue 005** (uman adopting the same shared
  module).
