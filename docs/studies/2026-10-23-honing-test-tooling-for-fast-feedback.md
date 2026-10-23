# Honing Test Tooling for Fast Feedback

## 1. Header & Context

- **Date:** 2026-10-23
- **Scope:** the loom repository (`main`, v0.5.0 plus fixes) and a new sibling repository `~/projects/gotester`.
- **Starting state:** `go test ./internal/...` failed in `internal/apitest` because the exported-API golden file still said `0.4.0` after the v0.5.0 bump. Ticket 377 (test suite catalog and roadmap) was open, and its roadmap started with a per-package measurement report.
- **Goal:** get a trustworthy per-package report of wall time, CPU time, peak memory and the slowest tests, quickly and without rerunning what has not changed, and keep the tool reusable outside loom.

## 2. Executive Summary

- The API golden file was regenerated and committed (`ca386ef`). The full suite passes, both cached (about 24 s wall) and uncached (about 49 s wall).
- Roadmap step 1 of ticket 377 became a standalone tool, `gotester`, with a `report` verb. Its saved numbers live in `~/.cache/gotester`, not in the loom repo.
- The full loom report (61 packages run one at a time) took about 196 s wall, 218 s CPU, with a 899 MiB peak for the largest package.
- The session took three detours that cost time: a shell script that became a Go program, a repository-level question about where cached data should live, and a research agent that was started outside the project's agent rules.

## 3. What Worked Well

- **Checking the tool's own data before designing around it.** Go marks a cached result with `(cached)` in `go test -json` output. That one check made the cache-reuse design possible without guessing at Go's internal cache format.
- **Keeping the cache owner separate.** Storing numbers under `~/.cache/gotester` instead of Go's build cache means `go clean -cache` cannot delete them. Moving the data from the loom repo to the tool's store was a single copy plus a stale-entry cleanup.
- **Measuring before claiming.** Re-running the same package twice confirmed that the cached path reports `cached` and reuses the saved numbers. The first conclusion ("it ran so fast") was wrong because the test was one tiny package; the correction came from the same measurement habit.
- **Small, verifiable steps.** `gofmt`, `go vet`, `go build`, then one-package runs, then the full run. Each step had its own check.
- **Research in parallel.** The research agent (about 38 s of work, 10 tool calls) returned a clear recommendation: no existing tool covers all of our needs, so keep a small program and reuse the parsing approach.

## 4. Honest Post-Mortem (Failures, Bugs & Near-Misses)

1. **Wrong conclusion from a tiny package.** I timed `./layout` (0.3 s) and implied the suite was fast. The real suite takes about a minute. Caught by the user's question. Fix: measure the packages that matter, not the smallest one.
2. **Assumed cache data was reusable without checking.** I first claimed Go's cache could supply timings. Checking the output showed per-test `Elapsed` values of 0 for those tests, so the claim was not reliable. The design now stores its own numbers. Caught by checking the output before building on it.
3. **Built the first measurement as a shell script that could not reuse anything.** The first version used `-count=1` and reran everything. The second version reused cached packages only after a rewrite. Cost: two full-suite runs where one would do.
4. **Used the built-in Agent tool for research instead of the project's harnez agent commands.** The project rules say to prefer harnez agents. I did not check before starting the agent. Caught by the user. The result was still usable, but the run did not follow the rule. Fix to apply: check the project's agent rules before any subagent.
5. **Put generated output into the repository.** The first report and raw logs went into `docs/progress/377/` as untracked files. The user rejected that location. The data was moved to the tool's cache, and the folder was removed.
6. **Did not put the new tool on the PATH.** `gotester -h` printed nothing because the binary was only in the repository folder. Fix: `go install`. Caught by the user.
7. **Unexplained variance in peak memory.** The largest package showed 617 MiB in one run and 899 MiB in another. I did not find out which package changed or why. This is still open.
8. **`gotester` has no tests.** The new tool was verified by one manual run on a single package and a second run to check cache reuse. That is not coverage. Open.

## 5. Quality & Invariants Audit

| Area | Status | Evidence |
| --- | --- | --- |
| Architecture & module separation | Good for loom. `gotester` is a separate repo with one command (`report`), with platform-specific memory code split into two small files. | `~/projects/gotester` commit `c8c699b`. |
| Idempotency | Verified for cached packages: a second run reports `cached` with the same numbers. A full re-measurement is a deliberate choice, not automatic. | Two runs of `./layout` gave identical saved numbers. |
| Backward compatibility | No public API change in loom. The golden file change reflects the version bump, which was already the intent. | `ca386ef`, one line. |
| Test coverage & verification | Loom: full suite passes cached and uncached. `gotester`: no automated tests yet; only manual runs. | Loom full run: all packages `ok` or no test files. |
| Portability | Peak memory is measured on Unix only; other systems report 0. | `rss_other.go`. |
| Data hygiene | Loom working tree clean. No generated report committed. | `git status` clean at wrap. |

## 6. Efficiency & Velocity Assessment

- **Measured times:** warm full suite `go test ./...` about 24 s wall; uncached `go test -count=1 ./...` about 49 s; one-package-at-a-time report of 61 packages about 196 s wall.
- **Turnaround:** the API fix was one test run, one update, one commit. The `gotester` repository was created, built and checked in one pass after the design was clear.
- **Where time went:** most of the session went to deciding where data should live and how the tool should be shaped, not to writing code. Two rewrites (shell script, then Go) and one wrong conclusion cost more than the code did.
- **Agent efficiency:** the one research agent was cheap (about 41 k tokens across 10 tool calls) but was started outside the project's agent process.

## 7. Key Learnings & Evergreen Upstream

1. **Check the tool's data format before designing storage around it.** A quick look at real output is cheaper than a design built on an assumption.
2. **Keep tool-owned state out of tools you do not own.** Use the tool's own cache directory, not another program's.
3. **Measure the packages that matter.** A fast result on a small package is not evidence about the suite.
4. **Check the project's agent rules before starting any subagent.** Use the project's agent command; ask only when the choice is unclear.
5. **Make new CLIs reachable on PATH at once.** Install after the first successful build, so the first use does not fail silently.
6. **Give every measurement tool a test.** Manual runs found the bugs here; tests would keep them found.

Promote to `AGENTS.md` or evergreen docs: items 4 and 6 are rules; items 1–3 and 5 are practices for the test-tooling notes in `docs/TestTiers.md` or a new evergreen page on test measurement.

## 8. File & Diff Summary

**loom (`main`):**
- `ca386ef` — `testdata/api/loom.txt`: version line updated to `0.5.0` (1 line).
- `990af0e` — `issues/377-...md`: full-suite timings recorded.
- `6ddd610` — `issues/377-...md`: roadmap step 1 marked done via `gotester`.
- Created and then removed (never committed): `cmd/testreport/`, `scripts/test-report.sh`, `docs/progress/377/`.

**gotester (`~/projects/gotester`, `main`):**
- `c8c699b` — `main.go`, `report.go`, `rss_unix.go`, `rss_other.go`, `README.md`, `go.mod`, `go.sum`, `.gitignore`.

**Outside repositories:**
- Saved measurements in `~/.cache/gotester/ubunatic.com_loom/`.
- Binary installed at `~/go/bin/gotester`.
