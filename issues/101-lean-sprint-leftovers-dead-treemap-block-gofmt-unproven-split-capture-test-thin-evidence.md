# 101 — Lean-sprint leftovers: dead treemap block, gofmt, unproven Split capture test, thin evidence

**Status**: Open — leftovers from lean sprints
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Hygiene
**Related**: 081, 092, 088, 060, 097; `graph/treemap.go`, `split.go`, `docs/progress/`

---

## 1. Problem & Motivation

The lean sprints for 034-097 were closed with known gaps that no ticket tracked. They are
small, testable and need no human. Source: `docs/studies/2026-09-21-lean-sprint-session-report.md`.

## 2. Milestones (lean sprint, developer: haiku or luna:low, plan first)

### M1 - Code hygiene (from 081 and gofmt)
- Delete the dead commented block (about 110 lines) after the `return` in `layoutTreemapSquarified`
  in `graph/treemap.go`. No behavior change; `go test ./graph/...` stays green.
- Run `gofmt -w` on exactly the files `gofmt -l .` lists (cmd/loom-bench/main_test.go, cmd_test.go,
  examples/ansiviewer/ansiviewer/viewer.go, layout/layout.go, layout/layout_test.go,
  parse_ansi_evidence_test.go, parse_ansi_test.go, splash*.go, startup.go, theme_test.go).
  Formatting only, one commit, `go vet ./...` and `go test ./...` unchanged.

### M2 - Prove the 092 Split capture test
- The Split test for drag capture (drags must stay with the child that started them) was never
  shown red without the fix. Temporarily revert the fix in a scratch edit, confirm the test
  fails, restore the fix, and record the failing output in the ticket. If it does not fail,
  strengthen the test until it does.

### M3 - Evidence that proves something
- 081: frames s1 and s2 are identical for slice-dice and squarified. Choose inputs where the
  layouts differ (aspect ratios differ) for s1 and s2, keep s3.
- 088: the paged decode frames show only the first 24 rows per group. Emit all rows (paged files
  or one long file) so `cat docs/progress/088/*.ansi` shows the full 217-row audit.
- Evidence rules as before: frames from code, gated on `LOOM_EVIDENCE=1`, repo-root
  `docs/progress/<ticket>/`, run only the ticket's own evidence test.

## 3. Verification

`gofmt -l .` prints nothing, `go vet ./...` and `go test ./...` are green, M2 red-proof is
recorded here, regenerated frames are committed.
