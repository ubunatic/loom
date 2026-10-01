# 234 — Release v0.2.16 cannot be fetched: case-insensitive file name collision in docs/progress

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: 213 (added `Dialog.ansi`), docs/GoRelease.md

---

/goal Make the next loom release fetchable with `go get`, remove the colliding file, and add a check
that blocks any future case-insensitive path collision before release, or stop and report when blocked
on a user decision (e.g. whether to publish a new release tag) or denied permission.

## 1. Problem & Motivation
Consumers cannot depend on v0.2.16 (reported from the `loom-games` project):

```
$ go get codeberg.org/ubunatic/loom@v0.2.16
go: downloading codeberg.org/ubunatic/loom v0.2.16
go: codeberg.org/ubunatic/loom@v0.2.16: create zip: docs/progress/dialog.ansi: case-insensitive file name collision: "docs/progress/Dialog.ansi" and "docs/progress/dialog.ansi"
```

The Go module zip rejects paths that differ only in case, so every tagged version containing both
files is unusable as a dependency. Published tags cannot be changed; the fix needs a new release.

## 2. Technical Specification / Findings
- `docs/progress/dialog.ansi` was added in 339f523 (`feat(dialog): add modal dialog widget`);
  `docs/progress/Dialog.ansi` in 78e296d (ticket 213). Check which tags since 339f523..78e296d
  contain both (likely only the ones after 78e296d).
- `git ls-files | sort -f | uniq -di` finds this as the only collision on HEAD (2026-10-01).
- `docs/progress/` mixes naming: most files are PascalCase widget names (`Choice.ansi`), a few are
  lowercase (`chart.ansi`, `datepicker.ansi`, `dialog.ansi`, `form.ansi`, `gallery.ansi`).
  Find what writes these files and make it use one casing, so the collision cannot recur.
- Guard: a check (e.g. in `make test`/`make check`, or the release preflight per docs/GoRelease.md)
  that fails on case-insensitive duplicate paths. `go mod verify` does not catch it; a zip dry run
  (`go mod download` of a local tag via a GOPROXY=off/GOFLAGS trick, or `golang.org/x/mod/zip`
  `CheckDir`) does.

## 3. Implementation & Verification Plan
- Keep one of the two dialog files (merge content if both are referenced), update references.
- Normalize the generator's file naming; rename the remaining lowercase files if safe.
- Add the collision guard and a test proving it fails on a colliding pair.
- Verify: `make test-q1`; a module zip check of HEAD passes. Releasing v0.2.17 and checking
  `go get codeberg.org/ubunatic/loom@v0.2.17` from `loom-games` needs the user's go-ahead (outward-facing).
