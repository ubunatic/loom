<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 004 — Migrate uzu to depend on the shared loom module

**Status:** Closed — invalid

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

This migration is obsolete: uzu is deprecated and was never used as an active
downstream consumer of Loom. There is no live migration to verify or complete.
The earlier open-status assessment below is superseded by this finding.

The former `/home/uwe/projects/uzu` checkout is absent, but that absence is no
longer a blocker because the downstream application itself is deprecated.

The remaining migration and drift notes below are historical context only; they
do not describe active work while uzu is deprecated.

- Historical sequencing was: tag loom → point uzu at it → verify → delete the
  copy.
- At extraction time, uzu and loom were byte-identical except import paths.
- This issue is the mirror of **uman issue 005** (uman adopting the same shared
  module).
