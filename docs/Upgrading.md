# Upgrading Loom in Dependent Projects

How to move a project that imports loom to a newer release, and the breaking changes that need code
changes. Field data: the 2026-10-01 migration of 8 workspace repos
([study](studies/2026-10-01-loom-module-migration.md), issue 235).

## Breaking changes by release

| Release | Change | What callers do |
|---|---|---|
| Unreleased | `loom.Style` gained `Italic`, `Strike`, and `Invert` fields (issue 254 M1) | Use keyed `Style` literals or add values to positional literals |
| v0.2.18 | Module path `codeberg.org/ubunatic/loom` → `ubunatic.com/loom` (issue 235) | Rewrite imports, require `ubunatic.com/loom` |
| v0.2.15 | `Widget` lost the bool `HandleKey`/`HandleMouse`; only `ConsumeKey`/`ConsumeMouse` returning `EventResult` remain (9dd183b) | Map results as below |
| v0.2.6 | `EventResult` introduced (issue 126) | — |

Add a row here with every release that breaks callers.

## Module path

- `ubunatic.com/loom` is a vanity path: `https://ubunatic.com/loom?go-get=1` (301 to `/loom/`) serves
  `go-import "ubunatic.com/loom git https://codeberg.org/ubunatic/loom"`. Source: `~/projects/ubunatic.com/loom/index.html`.
- The code stays on Codeberg: repository URLs (`https://codeberg.org/ubunatic/loom`, issues, releases)
  do not change. Tags before v0.2.18 declare the old path, so `go get ubunatic.com/loom` needs ≥ v0.2.18.
- Rewrite only module/import paths, never URLs:
  `perl -pi -e 's{(?<![/@])codeberg\.org/ubunatic/loom}{ubunatic.com/loom}g' <files>`.

## Switching a dependent

1. Rewrite `.go` files and live docs; leave closed `issues/` and dated `docs/studies/` as history.
2. `go mod edit -droprequire codeberg.org/ubunatic/loom`, `go get ubunatic.com/loom@<tag>`, `go mod tidy`.
   Never text-edit `go.sum` or `go.work.sum`: a rewritten line is a checksum for a module that does not exist.
3. Repos that build against the checkout keep it: `replace ubunatic.com/loom => ../loom`.
4. Verify with `GOWORK=off go build ./... && GOWORK=off go vet ./...`: a `go.work` that uses `../loom`
   hides problems with the released module.

## bool → EventResult

In the old API, `HandleKey`/`HandleMouse` returning **`true` meant quit**, not "consumed".

| Old | New |
|---|---|
| `return true` | `return loom.Quit()` |
| `return false`, key was used (navigation, state change) | `return loom.Handled()` |
| `return false`, key not used | `return loom.Ignored()` |
| `return child.HandleKey(e)` | `return child.ConsumeKey(e)` (forward the result unchanged) |
| child confirmed a selection | check `result.Done` (see [Widgets](Widgets.md) §8) |

Pitfalls seen in the migration: mapping old `true` to `Handled()` (q no longer quits), and falling
through to `Ignored()` after a key that changed state, e.g. a wizard step. Add a test that the quit keys
return `Quit` and an unrelated key returns `Ignored`.

## Releasing so dependents can fetch

The Go module zip rejects paths that differ only in case; such a tag can never be fetched (v0.2.16,
issue 234). A test guards tracked paths. After a release, check from a scratch module:
`go mod init x && go get ubunatic.com/loom@<tag> && go build ./...`.
