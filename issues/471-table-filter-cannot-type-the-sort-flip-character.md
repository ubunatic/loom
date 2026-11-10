# 471 — Table filter cannot type the sort-flip character

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: 458, 277

---

## 1. Problem & Motivation
Table's type-to-filter cannot contain `!`: the `sort_flip` binding (`spec/defaults.yaml:365`,
`["!"]`) and an extra `e.Text == "!"` test take it before text input (`table.go:567-572`, then
backspace `:573-580`, text `:581-585`). Searching for "Hello!" or a shell command is impossible.
Backspace also removes one rune, not one grapheme (`table.go:576`). Found in 458 batch 6
(`docs/studies/tuios/report-widgets.md`).

## 2. Technical Specification / Findings
- Options: flip the sort with a non-printing key (e.g. Shift+Tab next to Tab's sort cycle,
  subject to 277's Tab question) or by clicking the sorted header; or let `!` flip only while
  the query is empty. Remove the hard-coded `e.Text == "!"` either way, so the spec is the only
  source (docs/Spec.md).
- Backspace by grapheme, as in TextEntry.

## 3. Implementation & Verification Plan
/goal The Table filter accepts every printable character and deletes by grapheme, and sort flip
stays reachable through a spec binding, with tests; stop and report when blocked on the key
choice or a denied permission; on a loop or no progress, clear the goal and write a failure
report.

- Re-check live code and 277 first; table test typing `a!b` and backspacing over `👍🏽`.
