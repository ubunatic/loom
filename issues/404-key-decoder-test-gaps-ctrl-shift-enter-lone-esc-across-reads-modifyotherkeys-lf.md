# 404 — Key decoder test gaps: ctrl-shift-enter, lone ESC across reads, modifyOtherKeys LF

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Test
**Related**: 398 (Alt+Enter), 053 and 354 (split escapes at the Pane level), 391 (CPR/DA replies leak through `scanEscapeKey`)

---

/goal Add decoder-level tests for the three gaps below and decide the LF case, or stop and report when blocked on a user decision.

## 1. Problem & Motivation
A Jules ideas run (2026-11-01, `reports/jules/key-decoding.md`, GitHub PR #49) listed near-misses in `event.go`. Host-checked at 5c3c603; two of its five proposals were dropped (modifyOtherKeys Ctrl+Enter is already tested in `event_alt_enter_test.go:22`; ESC LF keeps its documented `alt-` name, `event.go:165-170`).

## 2. Findings (host-verified)
- No test decodes `CSI 13;6u` as `ctrl-shift-enter` (no match for `13;6u` or `ctrl-shift-enter` in `*_test.go`).
- No `scanKey` test feeds a lone `ESC` then the rest of a CSI in a later call at the decoder level (`event_test.go:41` covers only partial UTF-8). Pane-level coverage exists from 053/354.
- `decodeModifyOtherKeys` maps only code 13 to `enter` (`event.go:479`); code 10 (LF) yields no key. Decide whether `ESC[27;5;10~` should be `ctrl-enter` or stay unmapped, and pin it with a test either way.

## 3. Plan
1. Table cases in `event_test.go` for `ctrl-shift-enter` and the code-10 decision.
2. A `scanKey` test: `[]byte{27}` returns `ok=false, used=0`, then the full sequence decodes.
3. `go test -count=1 .` then `make test-q1`.

## Decision
- 2026-11-02 (decide mode): `ESC[27;5;10~` stays unmapped. No terminal sends code 10 for Enter, and `ctrl-enter` would conflate it with Ctrl+J. Pinned in `event_decoder_gaps_test.go`.
