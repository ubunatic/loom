# 054 — Add installable `loom-bench` and `loom-demo` apps covering all examples

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: `examples/filebrowser`, `examples/monitor`, `examples/splash`,
`examples/split`, `examples/treemap`, `Makefile`

---

## 1. Problem & Motivation

Loom ships five example programs under `examples/` (`filebrowser`, `monitor`,
`splash`, `split`, `treemap`), each runnable only via `go run
./examples/<name>`. There is no single installable entry point that lets a
user or reviewer browse and exercise all of them, and no lightweight
harness that exercises every example as a smoke/benchmark check. `make
install` currently reports "nothing to install" (see `Makefile` `install:`
target), so there is no installed binary at all today.

Add two small, installable apps:

- `loom-demo` — an interactive launcher/menu that lists all `examples/*`
  programs and lets the user pick one to run, so the examples are
  discoverable and runnable without knowing each package path.
- `loom-bench` — a non-interactive harness that runs each example through a
  bounded smoke/benchmark pass (e.g. scripted input or a fixed-duration PTY
  run) and reports pass/fail and basic timing, so example regressions and
  performance drift are caught without manual review of every example.

## 2. Design Questions

- Where do the new apps live: `cmd/loom-demo` and `cmd/loom-bench` (matching
  the existing `cmd/validate-spec` convention), or under `examples/` itself?
- How does either app discover the set of examples: a static registry, or
  reflection over `examples/*` directories at build time?
- `loom-demo`: does it exec each example as a subprocess, or does it need
  Cobra-style menu integration where examples are wired in as subcommands?
  Consider reuse of existing key-handling/pane infrastructure documented in
  `docs/AgenticLoop.md`-adjacent issues (see #053, multi-key TTY reads) rather
  than duplicating input handling.
- `loom-bench`: what does "run" mean for examples requiring a real file tree
  (`filebrowser`) or live data source (`monitor`)? Needs bounded, scripted, or
  synthetic fixtures, not open-ended interactive runs. Follow the canary-first
  principle in `docs/Canary.md` before wiring PTY automation.
- Should `loom-bench` reuse the PTY-based verification pattern already used
  by `scripts/check-watch-pty.py` (see `Makefile` `watch-pty` target), or
  introduce a new harness?
- What does "installable" require here given `make install` currently states
  there is nothing to install — should `install:` gain real `go install`
  targets for these two new binaries, and does that change `docs/GoRelease.md`
  release wiring (GoReleaser build matrix)?

## 3. Verification & Acceptance

- `loom-demo` and `loom-bench` build via `go build ./...` and are wired into
  `make install` (or documented reason why not).
- `loom-demo` lists and launches every current example under `examples/`
  without hard-coded duplication of example logic.
- `loom-bench` runs a bounded pass over every example and reports pass/fail
  per example, exiting non-zero on any failure, suitable for CI use.
- Adding a new example under `examples/` requires no manual registration step
  beyond what the design in Section 2 settles on (or an explicit, documented
  registration step if full auto-discovery isn't feasible).
- `go test ./...` and `go vet ./...` pass.
- `make test` continues to pass with the new targets included.
