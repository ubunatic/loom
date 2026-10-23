# Test Optimization and Root Modularization

## 1. Header & Context

- **Date:** 2026-10-23
- **Scope:** The `loom` repository (`main`, post-v0.5.1 release).
- **Starting state:**
  - Release bump from v0.5.0 to v0.5.1 broke `internal/apitest` because the exported API snapshot (`testdata/api/loom.txt`) rigidly matched the literal version constant (Issue #379).
  - Test execution lacked concurrency bounds, causing machine load and swap exhaustion during full runs. CLI integration tests in `internal/session` and `internal/editor` rebuilt `./cmd/loom` repeatedly per test invocation (Issue #377).
  - PTY integration tests in `internal/session` relied on arbitrary `time.Sleep(200 * time.Millisecond)` calls instead of observable state synchronization (Issue #377).
  - The root package `loom` remained a 29k LOC monolith containing core ANSI styling, SGR parameter parsing, and all YAML specification unmarshaling and validation structures (`defaults.go`, ~900 lines), creating an invalidation choke point for the entire repository.
- **Goal:** Execute a lean sprint loop to:
  1. Make the API snapshot golden test resilient against release version bumps (#379).
  2. Bound test load, eliminate CLI test compilation overhead, and replace session PTY sleeps with deterministic state waits (#377).
  3. Execute foundational modularization of the root monolith by extracting `ansi` (#116) and `specdefaults` (#380) with 100% public API backwards compatibility.

## 2. Executive Summary

- **Issue #379 Resolved (`0f81580`)**: `internal/apitest` was updated with `normalizeAPI` to filter out the `var Version = ...` line from both observed API output and golden snapshots for `ubunatic.com/loom`. Golden files no longer break on clean version releases, while real public API additions, signature changes, and removals remain strictly guarded.
- **Issue #377 Load & Reliability Addressed (`072e4e8`, `88ce120`)**:
  - `Makefile` test targets now default to `TEST_PARALLEL ?= -p 2`, bounding package build/test concurrency.
  - `internal/clitest/clitest.go` was refactored with `RunTests` and `sync.Once` process caching: the CLI binary is compiled once per test process and cleaned up in `TestMain`, avoiding dozens of redundant `go build` executions during integration runs.
  - All 6 fixed `time.Sleep` calls in `internal/session/session_pty_test.go` were replaced with condition predicates (`waitForSessionScreenshot`, sentinel shell echo acknowledgments, and output pattern matching). Session test suite duration dropped to ~3.2s.
- **Issue #116 Modularization Completed (`0c111b9`)**:
  - Created standalone `ubunatic.com/loom/ansi` subpackage encapsulating `ColorProfile`, `Color`, `Style`, zero-alloc buffer appenders (`AppendANSI`), and fast SGR parsing.
  - Re-exported type aliases and constructors from `loom` so callers and existing code remain 100% backward compatible without breaking changes.
- **Issue #380 Modularization Completed (`239c62b`)**:
  - Extracted all YAML unmarshaling, validation routines, and configuration structures from `defaults.go` into `ubunatic.com/loom/specdefaults`.
  - Re-exported `SpeccedDefaults` and all 40 configuration struct type aliases in `loom`. Downstream packages and tools can now import `specdefaults` directly without triggering root package compilation.

## 3. What Worked Well

- **Strict Zero-Coding Orchestration**: The host orchestrator operated entirely via clear specification handoffs to `codex:sol:med` developer agents and an architecture advisor, maintaining a clean diff-first review discipline without editing code directly.
- **Isolated Architectural Planning**: Launching `advisor-377` as a read-only architecture consultant mapped out package dependency graphs and identified circular dependencies (e.g., `measure` already importing `spec`, which required placing defaults in `specdefaults` rather than directly in `spec`).
- **Backwards-Compatible Go Type Aliasing**: Using Go 1.9+ type aliases (`type Style = ansi.Style`, `type LibDefaults = specdefaults.LibDefaults`) allowed moving hundreds of lines of implementation logic into focused subpackages without invalidating external downstream code or breaking public API contracts.
- **Process-Level Test Binary Caching**: Introducing `clitest.RunTests(m)` with a `sync.Once` build cache removed the massive I/O and CPU overhead of compiling `./cmd/loom` inside every single test function in the editor and session suites.

## 4. Honest Post-Mortem (Failures, Bugs & Near-Misses)

1. **Missing Directory in Canary Fixture Copy Lists**:
   - *Issue*: During both the `ansi` (#116) and `specdefaults` (#380) extractions, full test runs failed in `internal/specdefaults/defaults_test.go` because the test fixture copies specific repository directories into an isolated temporary module and omitted the new directories (`ansi`, `specdefaults`).
   - *How caught*: Caught immediately by developer agents running `make test-q1`.
   - *Fix applied*: Developer agents updated the glob/copy lists in `internal/specdefaults/defaults_test.go` to include the new packages before committing.
2. **Missing Dependency in `specdefaults` Extraction**:
   - *Issue*: When extracting `defaults.go`, `TabsKeys` remained in the root package while `specdefaults` validation needed it, causing a compilation failure on the first build.
   - *How caught*: Caught by `go vet ./...` during developer verification.
   - *Fix applied*: Moved `TabsKeys` into `specdefaults` and re-exported it in `loom`. Added regression test `internal/apitest/specdefaults_test.go` to assert that `specdefaults` never imports root `loom`.
3. **Repository-Wide Pre-Existing Lint Blockers**:
   - *Issue*: `make lint` failed across developer runs due to pre-existing formatting issues in untouched files (`key_defaults.go`, `terminal/terminal.go`).
   - *How caught*: Reported by developers during their pre-commit lint checks.
   - *Fix applied*: Developers verified that all touched/changed files were cleanly formatted with `gofmt` and passed Harnez scoped linting, noting the repository-wide blocker in their commit reports.

## 5. Quality & Invariants Audit

| Area | Status | Evidence |
| --- | --- | --- |
| **Architecture & Separation** | Excellent | `ansi` and `specdefaults` are now completely independent packages. Downstream packages can consume defaults and styling without pulling in the root package monolith. |
| **Idempotency** | Verified | Running tests repeatedly with warm caches returns cached passes without state mutation or lingering test directories. |
| **Backward Compatibility** | 100% | Re-exports and type aliases verified against `internal/apitest` snapshots (`testdata/api/loom.txt`, `testdata/api/ansi.txt`, `testdata/api/specdefaults.txt`). |
| **Test Coverage & Verification** | Complete | All unit, regression, PTY, and CLI tests passing under `make test-q1` and `TEST_PARALLEL=-p 2`. |
| **Lifecycle & Hygiene** | Clean | Zero zombie background tasks or subagents remaining; `make install` passed cleanly; git tree clean. |

## 6. Efficiency & Velocity Assessment

- **Execution Velocity**:
  - Issue #379 resolved and verified in under 3 minutes.
  - Issue #377 (load bounding and session sleeps) delivered across two milestones in ~5 minutes.
  - Issue #116 (`ansi` extraction) completed in ~7 minutes.
  - Issue #380 (`specdefaults` extraction) completed in ~4.5 minutes.
- **Turnaround & Parallelism**: Entire 4-ticket refactoring loop completed autonomously with zero human intervention required during code edits.
- **Resource Footprint**: Limiting test concurrency to `-p 2` and eliminating redundant CLI compiles drastically reduced CPU spikes and prevented system swap exhaustion.

## 7. Key Learnings & Evergreen Upstream

1. **Avoid Version Constants in API Golden Comparisons**:
   - Release version strings in source code (`var Version = "..."`) change purely as release metadata and do not constitute an API break. Exclude them from AST/doc golden comparisons to avoid breaking post-release builds.
2. **Always Cache Test Binary Compilations at the Process Boundary**:
   - Integration tests that execute a CLI binary should compile it once per `testing.M` execution via `sync.Once` and `TestMain`, rather than recompiling per test case.
3. **Check Canary Fixture Copy Lists When Adding Subpackages**:
   - When a repository uses in-tree fixtures that copy source packages into a temporary sandbox for isolation testing, adding a new subpackage requires updating those fixture copy lists.
4. **Use Type Aliases for Zero-Downtime Package Migrations**:
   - When decomposing monolith packages, introduce the new subpackage with full functionality and retain Go type aliases (`type T = pkg.T`) in the old package to keep 100% API stability across releases.

## 8. File & Diff Summary

**Commits (`main`):**
- `0f81580` — `fix(apitest): exclude release version from API snapshots (issue 379)`
- `84dd82f` — `docs(issues): close 379`
- `072e4e8` — `test: bound test concurrency and reuse CLI binary (issue 377)`
- `88ce120` — `test(session): replace PTY sleeps with observable waits (issue 377)`
- `ebb7b7c` — `docs(issues): close 377`
- `0c111b9` — `refactor(ansi): extract zero-allocation styling and SGR parsing (issue 116)`
- `e5f189b` — `docs(issues): close 116`
- `75e6ebf` — `docs(issues): file 380 extract specdefaults configuration layer`
- `f0c7a1a` — `docs(issues): start 380`
- `239c62b` — `refactor: extract standalone specdefaults configuration (issue 380)`
- `c8f89e8` — `docs(issues): close 380`

**Key Files Changed:**
- `ansi/`: Created `style.go`, `parse.go`, `ansi_test.go`.
- `specdefaults/`: Created `defaults.go`.
- `defaults.go`, `style.go`, `parse_ansi.go`, `canvas.go`: Replaced monolith implementations with subpackage forwarding and type aliases.
- `internal/clitest/clitest.go`: Added shared binary process cache and `RunTests`.
- `internal/session/session_pty_test.go`: Eliminated sleeps via condition polling.
- `Makefile`: Added `TEST_PARALLEL ?= -p 2`.
- `testdata/api/`: Added snapshots for `ansi.txt` and `specdefaults.txt`; updated `loom.txt`.
