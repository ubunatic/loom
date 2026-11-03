# 410 — loom session search allocates a string per scrollback window

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Performance
**Related**: [392](392-loom-session-mouse-modes-off-full-basic-basic-suppresses-drag-to-preserve-text-selection.md)

---

## 1. Problem & Motivation
`findSessionMatches` (`internal/session/session_search.go:38`) builds `string(runes[start:start+len(want)])` for every start position of every scrollback row, then compares it with `==` or `strings.EqualFold`. A long scrollback creates one string per character, which may slow search-as-you-type.

Found by Jules ideas run `session-vt` (GitHub #76, Jules ID 10548991855057164559); the reference was checked against main 14ec5a9.

## 2. Technical Specification / Findings
Compare the rune sub-slice with `want` directly: exact rune equality, or `unicode.SimpleFold`-based equality when case-insensitive. Results must match `strings.EqualFold` for the cases the existing search tests cover.

## 3. Implementation & Verification Plan
/goal Session search no longer allocates per window and returns the same matches; measure first with a benchmark over a large scrollback, and drop the ticket if the gain is negligible. Stop and report when blocked on a user decision.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| implement | jules-implement-410 | 4151500993214720974 | GH #80 | adopted: host benchmark (5000x120 rows, miss query) 15.8 ms to 8.3 ms per search, allocations unchanged (74994, they come from sessionRowRunes); the report's benchmark file was not in the PR; host added an EqualFold parity test; make test-q1 green |
