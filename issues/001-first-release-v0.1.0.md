<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 001 — Cut the first release (v0.1.0)

**Status:** Done — v0.1.0 tagged and pushed; one doc follow-up left (see below)

**Priority:** P1 — the goal this repo exists for

## Motivation

loom was extracted from uzu so that uman (and future tools) can import a shared
widget library instead of reaching into `uzu/loom`. Consumers need a tagged,
`go get`-able version.

## Current state (already done in this repo)

- [x] Source copied from `uzu/loom` (35 `.go` files).
- [x] Module path set to `codeberg.org/ubunatic/loom` (`go.mod`).
- [x] Test-package import paths rewritten `uzu/loom` → `loom`.
- [x] `go build ./...` and `go test ./...` green.
- [x] `REUSE.toml`, `.gitignore`, `Makefile`, `README.md` scaffolded.
- [x] Remote wired: `ssh://git@codeberg.org/ubunatic/loom.git`.

## Remaining before tagging

- [x] Create the repository on Codeberg (`ubunatic/loom`) and push `main`.
  Public and reachable.
- [x] Resolve blocking issues: 002 (driver helper, done), 003 (API audit, done),
  005 (license — keep AGPL, decided). 004 (uzu migration) follows the tag.
- [x] `go vet ./...` clean (also `go build`/`go test` green).
- [x] Decide versioning: start at `v0.1.0` (pre-1.0, API may still move) —
  consistent with uman/uzu which are both `v0.1.x`.
- [x] Tag `v0.1.0` and push the tag. Confirmed 2026-08-04:
  `git ls-remote --tags origin` → `refs/tags/v0.1.0` (`0600f48`).
- [x] Smoke-test consumption from a scratch module. Superseded by a stronger
  proof: `uzu/go.mod` requires `codeberg.org/ubunatic/loom v0.1.0` and
  `uzu/go.sum` carries real module-proxy hashes for it, so the tag resolves and
  builds through the public proxy.

## Follow-up (2026-08-04)

- [ ] **`README.md` "Status" is stale.** It still reads "Freshly extracted from
  `uzu`. See `issues/` for the work remaining before the first tagged release."
  — untrue since the tag landed. Same for this issue index's header, which
  frames every issue as pre-release work. Replace with the actual state:
  released `v0.1.0`, pre-1.0, API may still move.

## Open questions

- ~~Does Codeberg need the repo created via the web UI / API first?~~ Answered:
  Codeberg does **not** auto-create on push — the repo must exist first. It
  now does.
- Any CI on Codeberg (Woodpecker) desired for future releases, or manual tag
  for now? Still open.
