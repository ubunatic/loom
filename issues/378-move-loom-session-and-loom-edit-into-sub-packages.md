# 378 — move loom session and loom edit into sub-packages

**Status**: Open
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
