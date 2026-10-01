# Loom module migration reports (issue 235 M3)

Loom moved from `codeberg.org/ubunatic/loom` to `ubunatic.com/loom`; v0.2.18 is the first release
under the new path. Every dependent therefore also upgrades to v0.2.18. One section per repo, written
by the developer agent (or the host, for the first three) for later assessment of the upgrade cost.

## Cross-repo findings

- **The bool → EventResult mapping is the main trap.** Before ~v0.2.13 `HandleKey`/`HandleMouse`
  returned `bool`, where `true` meant **quit**, not "consumed". Now `ConsumeKey`/`ConsumeMouse` return
  `loom.EventResult`: `Quit()` for old `true`, and `Handled()` or `Ignored()` for old `false`, depending
  on whether the widget used the key. harnez got this wrong first (q no longer quit).
- A consumed key that changes internal state (e.g. a wizard step) must return `Handled()`, not fall
  through to `Ignored()` (settings).
- Never text-rewrite `go.sum`/`go.work.sum`; let `go mod tidy` update them (loom-games, cati).
- Verify with `GOWORK=off`: a local `go.work` with `../loom` hides problems with the released module.
- Wish for loom: a migration note per release listing interface changes and the bool semantics above.

## loom-games — v0.2.17 → v0.2.18 (host)

- Commits e78fc2d, 9b66920. 12 files, imports only; no API breakage.
- Problem: the path rewrite also changed a `go.work.sum` checksum line and a dated study; both reverted.
- `make test-q1` green, `GOWORK=off` build/vet green. Effort: small.

## settings — v0.2.13 → v0.2.18 (host)

- Commits 4720239, c31329c. 8 files.
- API: `Editor` and `Wizard` moved to `ConsumeKey`/`ConsumeMouse`; wizard uses `EventResult.Done` to tell
  a confirmed selection from navigation, and `Quit` from the form to step back.
- Problems: first test run had 2 wizard failures (Done vs navigation), fixed; host review then fixed
  step transitions returning `Ignored()` instead of `Handled()`.
- Host `make test-q1` green. Effort: medium.

## cati — v0.2.11 → v0.2.18 (host)

- Commits a346b6a, c73e875. Only `examples/imgbrowser` and `examples/mediabrowse` use loom.
- API: browser, preview and media panes moved to `ConsumeKey`/`ConsumeMouse`.
- Problems: first commit left `go.sum` stale (instruction misread), fixed by `go mod tidy`. Pre-existing:
  `make test` pins `GOTOOLCHAIN=go1.25.0` while `go.mod` needs 1.26, so tests ran via `GOWORK=off go test ./...`.
- Full suite green. Effort: medium. Cosmetic leftover: `if result.Quit { return result }; return result`.

## harnez — v0.2.5 → v0.2.18 (developer dev-235-harnez)

- Commits 23929d4, edcee35. 5 files (`go.mod`, `go.sum`, `internal/usage/loom.go` + test,
  `docs/ubunatic/GoPerformance.yaml`, the source of the bundled doc).
- API: `UsageLoomWidget.HandleKey` → `ConsumeKey` (q/Esc/Ctrl-C return `loom.Quit()`, others
  `loom.Ignored()`); `HandleMouse` → `ConsumeMouse` returning `loom.Ignored()`.
- Problems: first mapped old `true` (quit) to `Handled()`; caught in host review, fixed with a test.
- Host `make test-q1` green. Effort: ~15 min plus one review round.
- Loom improvement (agent): migration note with minimum Go version and interface changes per release.

## voxi — v0.2.1 → v0.2.18 (developer dev-235-voxi)

- Commit 9052d83. 5 files: `examples/miclevel/watch.go` (only importer), its README,
  `docs/LiveMicMeter.md`, `go.mod`, tool-generated `go.sum`.
- API: no breakage; the example uses `Pane.RunWatch`, `BuildWidget`, `New`, `Frame`, `Cadence`,
  `graph.RenderBar` and defines no key/mouse handlers. No other dependency was bumped.
- `make test-q1` green; `make install` skipped on purpose (reinstalls voxi's user services, and the voxi
  binary does not import loom); `go build ./examples/miclevel` green. Effort: small, one pass.

## psync — v0.2.0 → v0.2.18 (developer dev-235-psync)

- Commit fb1993e. 4 files: `internal/pick/pick.go` (only importer), `ROADMAP.md`, `go.mod`,
  tool-generated `go.sum`. `docs/LoomAdoption.md` keeps its repository URL.
- API: no breakage; `loom.Item`, `NewChoice`, `RunPane` compile unchanged; no key/mouse handlers.
- `make check` (no test-q1 target) green, `make install` green. Effort: ~10 min.

## uzu — not migrated

- Archived instead (user decision): moved to `~/projects/archive/uzu`; it stays on
  `codeberg.org/ubunatic/loom v0.1.0`.
