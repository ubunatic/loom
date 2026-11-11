# 485 — render flush: fmt.Sprintf per row and os.Getenv per parse in the hot path

**Status**: Backlog
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Enhancement
**Related**: 481

---

## 1. Problem & Motivation
Lead from Jules ideas run GH #132 (render hot path, not started by the 481 sweep); all five
citations checked by the host on a0ab6203. The speed gain is not measured.

## 2. Technical Specification / Findings
- `parse_ansi.go:16` `useFastANSI` calls `os.Getenv("LOOM_FAST_ANSI")` on every parse
  (callers `parse_ansi.go:25`, `canvas.go:258`). Caching it must keep tests that use
  `t.Setenv` working.
- `fmt.Sprintf` builds cursor moves per row in `FlushWithConfig`: `canvas.go:587`, `:596`,
  `:600`; and in pane clearing: `pane_terminal.go:234`. `strconv.AppendInt` into a reused
  buffer avoids one allocation per row.

## 3. Implementation & Verification Plan
/goal Measure first with the existing render benchmarks (481 PR B), then remove the per-row
`fmt.Sprintf` and the per-parse `os.Getenv` if the benchmarks show fewer allocations with
byte-identical output; stop and report when blocked on a design decision or denied permission;
on a loop or no progress, clear the goal and write a failure report.

- Re-check live code first. Keep output byte-identical (existing golden and flush tests).
- `make test-q1`, `make install`.
