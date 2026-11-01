# 398 — DecodeKey: ESC CR decodes as alt-\r instead of alt-enter; modifyOtherKeys alt-enter decodes empty

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 321 (kitty keyboard: shift-enter, ctrl-enter), harnez/taia#1 and taia#2 (chat input: Enter sends, Alt+Enter newline)

---

/goal Alt+Enter decodes to one stable key name (`alt-enter`) for the legacy `ESC CR` encoding and the
xterm modifyOtherKeys form, proven by `DecodeKey` table tests; or stop and report when blocked on a user decision.

## Findings (taia#1 review, 2026-11-01, probe of `DecodeKey`)

- Legacy terminals send Alt+Enter as `ESC CR`; `DecodeKey` returns `alt-\r`, not `alt-enter`. Bindings
  written as `alt-enter` never match.
- xterm modifyOtherKeys `\x1b[27;3;13~` (Alt+Enter) decodes to `""`.

## Motivation

`taia` (harnez org) uses loom's `TextArea` for chat input with Enter = send, Alt+Enter = newline. taia
works around this in its own key routing for now; the fix belongs in loom.

## Notes

Check live code first; with kitty keyboard (321) on, Alt+Enter may already decode correctly.

## M1 — decode both forms as `alt-enter` (lean sprint 2026-11-01)

Searcher findings (haiku:low): legacy `ESC CR` falls to the generic alt fallback `event.go:217` (`"alt-"+string(b[1:])`); `decodeModifyOtherKeys` (`event.go:458`) accepts only `a..z`, so `ESC[27;3;13~` reaches `event.go:300` (unhandled → ""). Kitty `ESC[13;3u` already yields `alt-enter`. Only test pinning old name: `event_kitty_test.go:100`.

Write set: `event.go`, `event_kitty_test.go` (+ new table cases), `docs/TuiInput.md:95` (alt-enter in all modes), `docs/Upgrading.md` breaking-changes row (`alt-\r` → `alt-enter`), optionally `cmd/loom-probe/keys.go:24` hint.
Tests: `ESC CR`, `ESC[27;3;13~`, kitty `ESC[13;3u` all → `alt-enter`; near-miss: `ESC LF` unchanged, `ESC[27;2;13~` (shift-enter) and `ESC[27;5;13~` (ctrl-enter) decode to their names or stay as before (state which), letters via modifyOtherKeys unchanged.

## Decision
- Models per owner: searcher haiku:low, dev haiku:med, reviewer opus:low (overrides CLAUDE.md codex:sol:med default for routing tickets; decoder-only change).
- Rename `alt-\r` → `alt-enter` without alias; recorded as breaking in docs/Upgrading.md.
