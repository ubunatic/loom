# Public API stability

Loom is a public library. Stable packages are `ubunatic.com/loom`,
`graph`, `layout` and `measure`. Preserve their exported signatures and documented
behavior between compatible releases. Additions require review too. An intentional
breaking release must record migration instructions in `docs/Upgrading.md`.

`collector`, `gallery`, `media`, `spec`, `syntax`, and all importable packages under
`examples/` are experimental. Their package documentation says so; they may evolve
between releases. `internal/` packages and `cmd/` executables are implementation
details rather than supported Go imports. Experimental does not authorize accidental
breakage: snapshots cover every importable package outside `internal/`.

## Reviewing API changes

`go test ./...` runs `internal/apitest.TestPublicAPI`, which discovers packages with
`go list` and compares `go doc -all` output with `testdata/api/<package>.txt`.
Subpackage separators become `__` in filenames. Missing and stale snapshots fail,
so adding or removing a package also requires review. Documentation is part of this
baseline. Use the Go toolchain declared in `go.mod` and CI when updating; a toolchain
format change may require a separately reviewed baseline refresh.

After reviewing the source change, accept it with:

```sh
go test ./internal/apitest -run '^TestPublicAPI$' -update
git diff -- testdata/api
```

Review and commit the snapshots with the implementation. Never update a snapshot
solely to silence a failure. This check detects declaration/documentation drift;
behavioral compatibility still depends on tests and code review. Platform-specific
APIs need validation on their supported platforms; the checked-in baseline describes
Linux, the repository's current CI platform.
