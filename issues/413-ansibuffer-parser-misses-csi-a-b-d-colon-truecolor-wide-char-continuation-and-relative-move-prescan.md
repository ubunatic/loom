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

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| implement | jules-implement-413 | 18068907042553961998 | GH #112 | rejected (wave 1, 429): no tests; claims a prescan fix it did not make; CSI B moves y without bound while the prescan ignores it (likely out-of-range write); rewrites every `:` to `;` in SGR and CSI positions |

Lead from the rejected PR: CSI A/D clamp at 0, CSI B adds to y, a wide rune writes a continuation cell at x+1. Do it locally with table tests per case, a prescan that tracks relative moves, and colon parsing limited to the SGR 38/48 sub-parameters.
