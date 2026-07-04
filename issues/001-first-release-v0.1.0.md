<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 001 — Cut the first release (v0.1.0)

**Status:** Open

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

- [ ] Create the repository on Codeberg (`ubunatic/loom`) and push `main`.
  **(User action — Codeberg does not auto-create on push.)**
- [x] Resolve blocking issues: 002 (driver helper, done), 003 (API audit, done),
  005 (license — keep AGPL, decided). 004 (uzu migration) follows the tag.
- [x] `go vet ./...` clean (also `go build`/`go test` green).
- [x] Decide versioning: start at `v0.1.0` (pre-1.0, API may still move) —
  consistent with uman/uzu which are both `v0.1.x`.
- [ ] Tag `v0.1.0` and push the tag. **(User action — needs the Codeberg repo.)**
- [ ] Smoke-test consumption from a scratch module: `go get
  codeberg.org/ubunatic/loom@v0.1.0` then import + build.

## Open questions

- Does Codeberg need the repo created via the web UI / API first, or will a push
  to a non-existent repo be rejected? (Codeberg does **not** auto-create on
  push — the repo must exist first.)
- Any CI on Codeberg (Woodpecker) desired for the release, or manual tag for
  now?
