# 189 — Rank Choice filter matches with fuzzy scoring

**Status**: Closed
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: choice.go (refilter), issues/177, issues/180

---

## 1. Problem & Motivation
`Choice` filters by case-insensitive substring in original order; typing `fbw` does not find `filebrowser-widget`. Bubbles' list uses fuzzy matching with ranking (issue 177).

## 2. Technical Specification / Findings
Add a fuzzy matcher (subsequence with scores for consecutive and word-start hits, no dependency) as an option on `Choice`; sort by score, keep the original order for ties; highlight the matched runes. Substring stays the default.

## 3. Implementation & Verification Plan
/goal Add opt-in fuzzy ranking to `Choice` with scoring and ordering tests; stop and report if highlighting needs a ChoiceStyle change the user must decide.
