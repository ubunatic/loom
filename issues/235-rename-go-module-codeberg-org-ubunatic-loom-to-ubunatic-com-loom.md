# 235 — Rename Go module codeberg.org/ubunatic/loom to ubunatic.com/loom

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Chore
**Related**: 234 (fetchable release first); ubunatic.com repo `loom/index.html`

---

/goal Make `ubunatic.com/loom` the module path of loom, released and fetchable with
`go get ubunatic.com/loom@<tag>`, with in-workspace dependents switched over, or stop and report when
blocked on a user decision (release/push/deploy approval) or denied permission.

## 1. Problem & Motivation
The user wants loom published under the vanity path `ubunatic.com/loom` instead of
`codeberg.org/ubunatic/loom`, like `ubunatic.com/harnez`. The code stays hosted on Codeberg.

## 2. Technical Specification / Findings (2026-10-01)
- Redirect page exists: `~/projects/ubunatic.com/loom/index.html` has
  `go-import "ubunatic.com/loom git https://codeberg.org/ubunatic/loom"` and `go-source`.
  But `curl "https://ubunatic.com/loom?go-get=1"` returned no `go-import` tag — check whether the
  page is deployed (ubunatic.com deploys on push to `main`); fix that first (canary-first, docs/Canary.md).
- In loom: `go.mod`, ~177 `.go` files, `.goreleaser.yaml` (ldflags/paths), README.md and several
  docs reference the old path. `dist/` is build output, not source. Leave closed `issues/` and dated
  `docs/studies/` as history unless a link breaks.
- Workspace projects whose `go.mod` mention the old path: cati, harnez, loom-games, psync, settings,
  termaid, uman, uzu, voxi. Each is a separate repo (see ~/projects/CLAUDE.md multi-repo rule).
- Existing tags stay under the old path; the rename takes effect from the next release.
  Decide whether the old path gets a deprecation note (`// Deprecated:` in a final old-path release
  is impossible without a separate module; a README note suffices).
- Do after 234, so the first `ubunatic.com/loom` release is fetchable.

## 3. Implementation & Verification Plan
- M1: verify/deploy the vanity page (`go list -m ubunatic.com/loom@latest` must resolve).
- M2: rename in loom (`go mod edit -module`, rewrite imports, goreleaser, docs); `make test-q1`,
  `make install`; release after user approval; `go get ubunatic.com/loom@<tag>` from a temp module.
- M3: switch dependents one repo at a time (`go.mod` + imports, `go mod tidy`, build/test),
  one commit per repo; no pushes without approval.

## M1 delivered (host, 2026-10-01)
The vanity page is live: `https://ubunatic.com/loom?go-get=1` redirects (301) to `/loom/`, which serves
`go-import "ubunatic.com/loom git https://codeberg.org/ubunatic/loom"`. The earlier check missed the
redirect. `GOPROXY=direct go list -m -versions ubunatic.com/loom` lists v0.1.0..v0.2.17. Those tags still
declare the old module path, so `go get ubunatic.com/loom` works only from the first release after M2.

## M2 — Pre-Work / Required Refinements (host)
- Rename with `go mod edit -module ubunatic.com/loom` and a scripted rewrite of import paths in all tracked
  `.go` files (no hand edits); then `gofmt -l` must be empty and `go build ./... && go vet ./...` pass.
- Also update `.goreleaser.yaml`, Makefile, README.md, AGENTS.md, live docs in `docs/` and `website/` (if
  present) and any test that builds `codeberg.org/ubunatic/loom/cmd/...`. Keep Codeberg *repository* URLs
  (clone, issues, releases, go-source) — only the Go *module/import path* changes.
- Leave closed `issues/`, `docs/studies/` and `dist/` untouched.
- Final check: `git grep -n 'codeberg.org/ubunatic/loom'` — every remaining hit is a repository URL or history;
  list the categories in "M2 delivered".
- One `make test-q1`, `make install`. No release, tag or push.
