---
title: Agentic Sprint from Mid-Flight Rows to Stable Graph Milestone
---

# Agentic Sprint from Mid-Flight Rows to Stable Graph Milestone

## Header & Context

Date: 2026-09-10. Scope: evaluate mid-flight progress on ticket 011, review
alignment against the Harnez target, execute an agentic sprint across immediate
blockers and feature requirements, reach a stable dashboard build, and decompose
the upcoming graph milestone into sub-tickets.

Loom had shipped stages 1–5 (static shell, watch mode, responsive layout, controls,
and the geometry evidence gate). Work was mid-flight on stage 6 (ticket 011).
However, test runs were intermittently stalling on an interactive TTY hang, golden
regression tests were coupled to the demo specification via in-memory mutations,
and row declarations lacked graph placeholders and box height capacity for the
canonical four-row target.

## Executive Summary

Executing this phase via the structured `/sprint` loop was demonstrably
effective. It prevented an entangled, high-risk refactor by separating concurrent
advisory discovery from single-threaded, test-driven sequential execution.

In Phase 1, three concurrent read-only advisors audited three distinct subsystems
in parallel: test TTY decoupling, test fixture isolation, and dashboard row
geometry. In Phase 2, four intermediate milestones were developed, tested, and
committed sequentially:
1. `fb0730a`: Decoupled `:help` from interactive `/dev/tty` in test environments
   (Ticket 021).
2. `1c9fedc`: Decoupled the baseline empty-shell golden test from the evolving
   example monitor via a dedicated fixture (Ticket 022).
3. `a3edd93`: Added negative schema controls for box rows in `validate-spec.py`
   (Ticket 023).
4. `a1efd2c`: Completed the 4-row dashboard target (Tickets 011 & 020) with
   `TruncateTextLeft` for right-aligned metrics, 8-row boxes, 10-row frame,
   Braille graph placeholders (`[⣿⣿  ]` and `[⣿⣿⣿⣿][⣀⣀⣀⣀]`), and a
   programmatic `SetRowsValues`/`Box(id)` data-binding interface for Ticket 013.

Once the application reached this stable build, stage 6 graph work (ticket 012)
was decomposed into sub-tickets 024 (primitives porting) and 025 (monitor
integration) in commit `540b053`. All tests, schema validations, and geometry
checks pass cleanly in ~1.1s.

## Was the "/sprint" Workflow Helpful?

**Yes, decisively.** The value of the `/sprint` workflow in this session came
from three specific structural properties:

1. **Parallel Discovery without Codebase Contention**:
   Rather than jumping straight into code edits while unaware of lurking debt,
   dispatching three concurrent read-only advisors surfaced critical hidden
   blockers simultaneously:
   - Advisor 1 reproduced the exact 2-second timeout panic in `TestCmdHelpDoesNotQuitInTest`
     caused by `/dev/tty` raw mode capture.
   - Advisor 2 flagged that `frame_test.go` was using an in-memory `b.Child = nil`
     mutation to keep empty-shell goldens passing, warning that expanding box
     height in `monitor.yaml` would break multiple regression tests.
   - Advisor 3 calculated exact 27-cell inner column budgets, identifying the need
     for `TruncateTextLeft` for numeric suffixes and an 8-row box height for the
     4 canonical rows.

2. **Strict Ordering of Hygiene before Feature Work**:
   Had ticket 011 been completed first, the test suite would have continued to
   hang on developer terminals, and modifying `monitor.yaml` would have broken
   shell regression tests. The sprint sequence enforced fixing the test hang
   (021), isolating the baseline fixture (022), and adding schema controls (023)
   *before* altering the example monitor specification and widget geometry.

3. **Incremental Commits and Clean Traceability**:
   Instead of a monolithic, diff-heavy commit touching 15 files at once, the
   sprint produced discrete, self-verifying commits. Each intermediate step had
   its own reproduction baseline, implementation diff, test verification, and
   issue tracker closure.

## What Worked Well

- **Repro-Before-Fix Baseline**: Running `go test -v -timeout 2s -run TestCmdHelpDoesNotQuitInTest`
  before touching `cmd.go` produced a concrete goroutine panic in `pane.Run`
  blocked on `/dev/tty`. The resulting fix with `testing.Testing()` and `export_test.go`
  eliminated the hang permanently (reducing test time from 16s timeout to 0.002s).
- **Test Fixture Decoupling**: Extractating `testdata/fixtures/empty-shell.yaml`
  allowed `examples/monitor/spec/monitor.yaml` to evolve from a 9-row empty shell
  into a 10-row, 4-row-per-box rich dashboard with Braille placeholders without
  breaking the baseline static shell golden test.
- **Oracle Corpus Expansion**: When `make test` caught `glyph outside oracle corpus: U+28E0`
  (`⣠`), inspecting `geometry_test.go` revealed that the test oracle whitelist
  had only hardcoded two specific Braille glyphs. Expanding the whitelist to the
  full Unicode Braille block (`\u2800`–`\u28ff`) brought the checker in line
  with the supported text policy documented in `docs/Geometry.md`.
- **Target Fidelity**: The rendered output of `go run ./examples/monitor` now
  directly mirrors the monochrome reference from `docs/HarnezUsageTarget.md`:
  four rows in All Usage (dual determinate bars, percentages), four rows in Load
  (single Braille timelines, and the dual split VRAM/GTT timeline).

## Honest Post-Mortem (Failures, Bugs & Near-Misses)

- **Test Expectation Typo in Left-Truncation**: In `truncate_test.go`, the test
  fixture for `é中ABC` with budget 4 and marker `.` initially expected `.中C`.
  Because `C` (1), `B` (1), `A` (1) fit 3 cells from the right, the correct
  cluster sequence from the right was `.ABC`. The test caught this immediately
  and the expectation was corrected.
- **Canvas Row Access vs Height**: In `cmd_test.go`, `cv.Rows()` was mistakenly
  used as a row slice in `strings.Join`, but `Canvas.Rows()` returns an `int`
  (height). The test was corrected to use `loom.Render(w, width, height)`.
- **In-Memory Fixture Mutation Debt**: The `b.Child = nil` hack in commit `c19f226`
  was a classic near-miss where a quick shortcut to pass tests introduced silent
  coupling. The sprint successfully excised this debt under Ticket 022 before
  it caused broader regressions.
- **Subagent State Polling**: Early in Phase 1, the orchestrator checked subagent
  status with a manual query before switching to idle wait. The loop adhered to
  reactive wakeup, and all subagents were cleanly drained and terminated in Phase 4.

## Quality & Invariants Audit

| Area | Evidence and result | Limit |
|---|---|---|
| Test Stability & TTY Independence | `go test -race ./...` runs in ~1.1s with zero interactive TTY prompts or stalls. | Full interactive `:help` on a physical terminal remains manually verifiable. |
| Test Decoupling & Isolation | Baseline empty-shell golden test loads `empty-shell.yaml`; `b.Child = nil` hack eliminated. | Example monitor has its own integration golden test. |
| Schema Conformance & Negative Controls | `validate-spec.py` validates 4 YAML documents and asserts 7 negative controls on row schemas. | Column-to-value count matching is checked at runtime in Go. |
| Layout & Geometry Invariants | 64x10 wide layout preserves 31x8 box rectangles, 1-cell padding, and 2-cell gap. Zero border clipping. | Tested headlessly against oracle cells; physical raster capture remains unattended. |
| Data-Binding Decoupling | `Rows.SetValues` and `Box.SetRowsValues` allow programmatic snapshot updates without YAML modification. | Simulated snapshot generation loop is scheduled for Stage 7 (Ticket 013). |
| Backlog Synchronization | Tickets 011, 020, 021, 022, 023 marked Closed with delivery evidence; tickets 024 and 025 filed. | `harnez status` depends on external CLI installation; tracker files verified directly. |

## Efficiency & Velocity Assessment

The sprint executed across 5 commits (`fb0730a`, `1c9fedc`, `a3edd93`, `a1efd2c`, `540b053`):
- 4 resolved tickets closed (021, 022, 023, 011), 1 review ticket closed (020), and 2 sub-tickets filed (024, 025).
- Net diff: ~420 insertions, ~85 deletions across Go core, tests, fixtures, specs, and issue docs.
- Wall-clock time: ~8 minutes from initial discovery to final sub-ticketing commit.
- Zero broken builds or regressions introduced across the sprint loop.

## Key Learnings & Evergreen Upstream

1. **Decouple TTY Hardware from Widget Tests**: In-memory widgets (`Choice`, `Table`,
   `cmdBar`) must never invoke hardware-dependent terminal allocators (`loom.New`)
   directly without a headless fallback or mock runner hook.
2. **Never Mutate Widgets to Satisfy Golden Fixtures**: When adding content to an
   example document, decouple the golden regression fixture into a dedicated
   file rather than wiping parsed children in test code.
3. **Truncate from the Left for Right-Aligned Data**: Right-aligned columns
   holding metrics, percentages, and timestamps must truncate leading characters
   (`TruncateTextLeft`) so unit suffixes and low-order values remain readable.
4. **Decompose Epics When Reaching Stable Platforms**: Breaking ticket 012 into
   pure rograph porting (024) and monitor integration (025) provides bounded,
   independently testable milestones for the next sprint.

## File & Diff Summary

Commits produced during the sprint:

| Commit | Scope | Description |
|---|---|---|
| `fb0730a` | `fix(cmd)` | Decouple `:help` from interactive `/dev/tty` in tests for ticket 021. |
| `1c9fedc` | `test(geometry)` | Decouple static shell golden from monitor spec for ticket 022. |
| `a3edd93` | `test(spec)` | Add negative schema controls for box rows for ticket 023. |
| `a1efd2c` | `feat(monitor)` | Complete tickets 011 and 020 with 4-row layout, graph placeholders, and data binding. |
| `540b053` | `docs(issues)` | Sub-ticket 012 into 024 (primitives) and 025 (integration). |
