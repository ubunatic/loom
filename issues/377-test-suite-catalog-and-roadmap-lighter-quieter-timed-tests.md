# 377 — test suite catalog and roadmap: lighter, quieter, timed tests

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Testing
**Related**: docs/TestTiers.md (no sleeps, PTY tests in the full tier), 369–376 (session PTY tests)

---

/goal The loom test suite is tight and concise: it tests the key behaviour, does not load the machine
(CPU, swap) or produce noise, and has a test report with run time and importance per package and per
slow test. Work follows the roadmap below one step at a time, each step measured before and after; or
stop and report when blocked on a user decision.

## Finding (user, 2026-10-23)

A host `make test-q1` run (right after a developer's own run) took the machine's CPU and swap
(swap 7.8 of 8 GiB used afterwards). The user wants step-by-step cleanup and a report, not one big run.

## Catalog (host, 2026-10-23, from files and existing logs, no test run)

- `make test-q1` = `harnez exec --quota-1 -- make test` = validate-spec, geometry-replay, `go vet ./...`,
  `go test ./...`. No `-p` limit: one package process per CPU core, each PTY test spawns children.
- 245 test files; 131 in the root package, 30 in `cmd/loom`, the rest in examples and helpers.
  39 packages report `ok`.
- Slowest packages (host log 374b): cmd/loom 53.4s, root 35.9s, gallery 28.1s, internal/examplesreg
  23.6s, examples/screens 21.4s, examples/winch 16.1s, internal/specdefaults 11.7s, examples/loomoji
  11.6s, examples/media 9.3s.
- `cmd/loom` grew from ~30s (issue 350) to 53.4s (374) with `session_pty_test.go` (42 tests).
- `time.Sleep` in tests (against TestTiers "no sleeps"): pane_pty_test.go 25, gallery_test.go 23,
  ticker_test.go 12, edit_pty_test.go 9, winch_pty_test.go 8, session_pty_test.go 6, others ≤5.
- `cmd/loom/live_tmux_test.go` runs real tmux in the full tier (skipped with `-short` or no tmux).
- Test logs are small (~4 KB per q1 run); output noise is not on disk, but verbose PTY logging was
  not measured yet.

## Roadmap (one step per milestone, measure before/after)

1. **Report:** a script or make target that runs one package at a time with `-p 1`, records wall time,
   CPU time and peak RSS (`/usr/bin/time -v`) per package, and the 20 slowest tests (`go test -json`),
   into a short table under `docs/progress/377/`. No full parallel run.
2. **Bound the load:** `make test` uses a spec'd `-p` limit (e.g. half the cores) and PTY tests use
   `t.Parallel` limits; measure peak RSS and swap before/after.
3. **Session PTY tests:** cut the 42 tests to the key behaviours (bar, tabs, restore, mouse, help,
   search), merge overlapping ones, replace sleeps with waits on screen content.
4. **Sleeps elsewhere:** pane_pty, gallery, ticker, edit_pty, winch: replace with event waits or move
   to the full tier only.
5. **Importance:** tag each slow test key / regression / redundant in the report; drop or merge
   redundant ones after host review.
6. **Host review flow:** the host reruns only the touched packages plus a bounded full run, not a second
   unbounded `make test-q1` right after the developer's.

## Rule until this lands (host)

Before any full test run the host tells the user and waits for a go.

## Incremental testing and modularization (user + host, 2026-10-23)

Go caches test results per package; a package re-runs only when its code or a dependency changes. Last
host run: 11 of 39 packages cached, 28 re-ran, because 373/374 touched the root package (`keycap.go`,
`defaults.go`), which almost every package imports.

- Root package `loom`: 118 files / 29,170 lines plus 131 test files / 28,456 lines. The real monolith.
- `cmd/loom`: 13 files / 4,453 lines plus 30 test files / 6,331 lines; `session.go` 1,266 and `edit.go`
  998 lines. Agents keep adding files to `cmd/loom` and the root package instead of sub-packages.

Roadmap step 0 (before step 2):

- Measure cache hits per typical change (one widget file, spec, `cmd/loom/session.go`) with
  `go test ./...` twice, and list which packages re-run.
- Propose a package split by an advisor (read-only), host + user review before any move:
  `session` (PTY tabs, bar, search) and `editor` out of `cmd/loom` with thin commands left there;
  in the root package, candidates are widget groups, spec defaults and terminal I/O, kept inside the
  API compatibility rules (docs/Upgrading.md breaking-changes table if exports move).
- Add a rule to AGENTS.md/docs: new features go in a sub-package, not `cmd/loom` or the root package,
  unless they are root widgets.
- Each split is its own milestone with before/after cache-hit and run-time numbers.
