# 478 — Measure and bound key latency and memory under Post load

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Performance
**Related**: 458, 320, 479

---

## 1. Problem & Motivation
`Pane.Post` never blocks and has no bound (`pane_run.go:50-62`), and `runPosted` runs the whole queued
snapshot before the loop reads input (`pane_run.go:66-81`). A fast producer (log tail, file watcher,
network feed) can therefore grow memory without limit and make keys wait behind its batch.

458 batch 9 (`docs/studies/tuios/report-testing.md`, canary A, `TestB9ABurst` in scratchpad
`b9-src/loom-head`): 8 producers x 12,500 posts, max sampled queue 44,821 to 70,823, the key ran after
151 to 85,055 posts (5 runs). The canary did not measure key send-to-handler time in ms, so the user
impact is not yet proved.

## 2. Technical Specification / Findings
- Canary first: measure key send-to-handler latency (p50/p99/max) and peak queue memory under a
  sustained producer, in a PTY test with deadlines, no sleeps.
- If latency is visible (decide the threshold from the numbers), fix the library: cap the work per loop
  turn so input interleaves, and choose a documented policy for the bound (block, coalesce, or document
  unbounded). Bubble Tea's `Send` blocks by design; compare before choosing.
- Event routing: start on terra:med (project rule).

## 3. Implementation & Verification Plan
/goal Key latency under Post load is measured, and either shown acceptable with the numbers recorded,
or bounded by a library change with a test that fails without it; stop and report when blocked on a
policy decision or denied permission; on a loop or no progress, clear the goal and write a failure report.

- Re-check live code first; start from `TestB9ABurst`.
- Full test tier for PTY tests (`docs/TestTiers.md`), raceFactor budgets; `make test-q1`, `make install`.
