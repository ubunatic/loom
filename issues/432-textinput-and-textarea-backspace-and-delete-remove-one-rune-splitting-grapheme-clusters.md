# 432 — TextInput and TextArea: backspace and delete remove one rune, splitting grapheme clusters

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 429 (Jules ideas run text-input, GH #113, proposal 3), 414 (shared cluster scanner), taia chat input

---

/goal Backspace and delete in `TextInput` and `TextArea` remove one whole grapheme cluster (base plus combining marks, ZWJ sequences, VS16), with table tests that fail before the fix; stop and report when blocked on a user decision.

## 1. Problem
`TextInput` backspace removes `t.runes[t.caret-1]` (`textinput.go:110`), one rune. Host probe on 8c8adec: `NewTextInput("é")` then backspace leaves `"e"`; editors remove the whole `é`. `TextArea` backspace and delete (`textarea.go:331`, `:334`) are rune-based too (not probed). Caret moves (left/right) likely split clusters the same way (not verified).

## 2. Plan
Use the measure cluster scanner from 414 (`measure` package) to find the previous and next cluster boundary; apply it to backspace, delete and left/right in both widgets. Table tests: `é`, a ZWJ family emoji, a flag, `❤️`, CJK.

## 3. Taia
taia's chat input is a loom text widget; users type accents and emoji there.
