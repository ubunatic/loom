# 414 — measure: share one grapheme cluster boundary scanner between StringWidthNew and ClustersNew

**Status**: Closed — done: 8397a50 (Jules PR 109, host-checked parity test); lone VS16 width follow-up filed
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Refactor
**Related**: reports/jules/measure.md (Jules ideas run, GitHub #88)

---

## 1. Problem & Motivation
`StringWidthNew` (`measure/measure.go:45`, its cluster loop near `:117`) and `ClustersNew` (`:299`) each find grapheme cluster boundaries (base rune plus modifiers, ZWJ, VS16) separately. If the two disagree, width and cell layout drift apart, which breaks the display-width rule.

## 2. Technical Specification / Findings
Jules proposal 1. Proposals 2 to 5 were micro-optimizations with no benchmark behind them, so they were dropped.

## 3. Implementation & Verification Plan
/goal One private `nextClusterBoundary` serves both functions; a test asserts, over an emoji/ZWJ/VS16/CJK corpus, that the sum of `ClustersNew` widths equals `StringWidthNew`. Stop and report when blocked on a user decision; on a loop or no progress, write a failure report here.
