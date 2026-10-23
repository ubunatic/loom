# 378 — move loom session and loom edit into sub-packages

**Status**: Closed — session and editor moved to internal/session, internal/editor, internal/clitest; cmd/loom thin commands (990a30a, ea9ce6b). Build, vet, package tests of the moved packages and cmd/loom green, install ok; full-suite run and cache probe parked in 377
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Refactor
**Related**: 377 (test roadmap, step 0 modularization), 369–376 (session), loom edit tickets

---

/goal `loom session` and `loom edit` live in their own packages with their tests, `cmd/loom` keeps only
thin Cobra commands that call them, behaviour is unchanged, and touching one feature re-runs only its
package and `cmd/loom`; or stop and report when blocked on a user decision.

## Decision (user, 2026-10-23)

Modularize before the big test assessment (377). A pure move, no behaviour change.

## Scope

- `cmd/loom/session.go` (1,266 lines) and `session_pty_test.go` → `internal/session`.
- `cmd/loom/edit.go` (998 lines) and the `edit_*_test.go` files → `internal/editor`.
- `internal/` so no public API is added (no API snapshot or docs/Upgrading.md change). If the developer
  finds a reason for a public package, stop and ask.
- Shared helpers used by both stay in one place (a small internal package or `cmd/loom`), no copies.
- `cmd/loom` keeps flag parsing and help; command names, flags, exit codes and output stay identical.
- Update callers: Makefile, docs, scripts that name the moved files or test names.

## Verification

- `go build ./...`, `go vet ./...`; `go test ./internal/session ./internal/editor ./cmd/loom`.
- Cache check: after the move, touch only `internal/session` and show that `go test ./...` re-runs
  only `internal/session` and `cmd/loom` (list from `go test` output).
- Host rule (377): the host asks the user before any full test run.

## Approved plan (host, 2026-10-23)

d378 plan approved (`internal/session` with `Run`, `internal/editor` with `Run(path, cfg)`, shared
`internal/clitest`, thin Cobra commands in `cmd/loom`). Changes:

- Two commits: session first, then editor, each building and passing its package tests.
- No "failing first" step for a pure move; existing tests are the guard.
- Machine load (user): do not run `make test-q1`. Run `go build ./...`, `go vet ./...`, and
  `go test -p 2 ./cmd/loom ./internal/session ./internal/editor ./internal/clitest`.
- Cache probe: after both commits, add a comment line to one file in `internal/session` (do not commit),
  run `go test -p 2 ./... 2>&1 | grep -v '(cached)'` once, report the re-run list, then revert.
- `make install` at the end. The host decides with the user on a full run.
