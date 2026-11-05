# 421 — syntax: shell lexer, colour the command after if, else, while and !

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 420

---

## 1. Problem & Motivation
The shell lexer from 420 (v0.8.0) colours the first word of a command, but not the word after the keywords
`if`, `elif`, `else`, `while`, `until` or after `!`. In `if grep -q x f; then`, `grep` stays plain.
Nice to have, not key.

## 2. Specification
- After `if`, `elif`, `else`, `while`, `until`, `!` (and the existing `then`, `do`), the next word is a command name
  (`CaptureFunction`), unless it is itself a keyword.

## 3. Verification
Table tests in the shell lexer tests for `if grep`, `while read`, `! test`, `else echo`. Release as a patch.
