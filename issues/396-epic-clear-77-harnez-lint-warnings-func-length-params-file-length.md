# 396 — Epic: clear 77 harnez lint warnings (func length, params, file length)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Refactor
**Related**: 225 (closed modularization pass), `docs/Make.md` (lint target), gokay assess 2026-10-31

---

## 1. Problem & Motivation

`make lint` exits 0 but prints 77 warnings across 50 of 1054 tracked Go files:
45 `go-func-length` (limit 100 lines), 17 `go-params` (limit 6), and 15
`go-file-length` (limit 800 lines). Warnings do not fail the target, so the
debt grows unseen. The gokay diagnosis on 2026-10-31 reported lint as
"skipped-missing-tool" because it does not run the repo's `harnez lint` path,
so the warnings were invisible to that check.

Files with the most warnings: `internal/session/session.go` (8), `vt/vt.go` (4),
`graph/treemap.go` (4), `ansibuffer.go` (4), `internal/editor/edit.go` (3).

This is an Epic: the fixes span many packages and each category needs a design
decision (refactor the function, or raise the limit with a reason).

## 2. Technical Specification / Findings

Reproduce: `make lint > <scratch>/make-lint.log 2>&1; grep -c 'go-' <scratch>/make-lint.log`
(expected 77 at 2026-10-31).

Categories and milestones:

- **M1 — decide the policy.** Per category, choose: split the code, or keep the
  warning with a documented reason. Record the decision under `## Decision` in
  this Epic. Do not change lint thresholds without that record.
- **M2 — `go-params` (17).** Replace long parameter lists with an options or
  struct type where the call sites already share data (example:
  `view.go:151 drawViewLine`, 7 parameters).
- **M3 — `go-file-length` (15).** Split files by concern. Pure moves, no
  behaviour change. Start with `internal/session/session.go` (8 warnings
  across its file-length and func-length findings) and `vt/vt.go` (979 lines).
- **M4 — `go-func-length` (45).** Extract helpers from the longest functions
  first (`Table.Draw` 159 lines, `VT.sgr` 128, `TextArea.Draw` 131). Keep
  each step behaviour-neutral and covered by the existing tests.

## 3. Implementation & Verification Plan

- Each milestone is its own commit with the tests of the touched packages.
- Acceptance for the Epic: `make lint` reports 0 warnings, or every remaining
  warning has a recorded reason under `## Decision`.
- Full verification before closing: `make test-q1` on the host, run by the host
  and not by the developer.
- Child tickets are created per milestone when that milestone starts, not
  before.
