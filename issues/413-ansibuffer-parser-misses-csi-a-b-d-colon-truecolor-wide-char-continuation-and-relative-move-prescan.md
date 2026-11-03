# 413 — ansibuffer: parser misses CSI A/B/D, colon truecolor, wide-char continuation and relative-move prescan

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: reports/jules/ansibuffer.md (Jules ideas run, GitHub #87)

---

## 1. Problem & Motivation
`ParseAnsiBuffer` (`ansibuffer.go`) reads `.ansi` design files, the sketches that `loom eval`, `loom measure` and `loom check-box` validate. Several common sequences are parsed wrongly, so a sketch can render differently than the terminal shows it.

## 2. Technical Specification / Findings
The host checked each citation against GitHub `main`:
- The CSI switch (`ansibuffer.go:645`) handles `C`, `H`, `f`, `G` and `d`, but not `A`, `B` or `D`.
- SGR parameters are split on `;` only (`ansibuffer.go:759`, `:779`). Colon truecolor such as `38:2::r:g:b` is ignored.
- A wide rune advances `x` by its width (`ansibuffer.go:733`) without writing a continuation cell. Whether `Serialize` then emits an extra space is not verified.
- The size prescan (`ansibuffer.go:579`) looks only at absolute moves. A relative move can go past the computed size.

## 3. Implementation & Verification Plan
/goal `ParseAnsiBuffer` handles CSI A/B/D, colon SGR truecolor and wide-rune continuation cells, and its prescan accounts for relative moves; table tests show each case fails before the fix. Stop and report when blocked on a user decision; on a loop or no progress, write a failure report here.
