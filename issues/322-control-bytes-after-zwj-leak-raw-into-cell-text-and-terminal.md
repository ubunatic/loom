# 322 — Control bytes after ZWJ leak raw into cell text and terminal

**Status**: Closed — Fixed: shared measure.JoinsAfterZWJ ends clusters at controls; Canvas.Row drops control runes from cell text.
**Priority**: P1 (High)
**Severity**: Major
**Category**: Security
**Related**:

---

## 1. Problem & Motivation
Grapheme grouping joined the rune after a ZWJ (U+200D) unconditionally. A
control after a ZWJ became part of the cluster and was written raw as cell
text: `"a‍\x1b[2J"` cleared the screen, `"‍\x1b]52;c;…\a"` wrote the
clipboard (OSC 52); `\n`, `\r`, BEL, C1 (U+009B) and DEL leaked too. Any app
rendering untrusted text through loom was exposed. Found while reviewing
harnez, which mirrors loom's grapheme logic.

## 2. Technical Specification / Findings
Duplicated ZWJ joins: `ParseANSI`/`ParseANSIOld` (parse_ansi.go), `ClipRow`
(rawscreen.go, used by RawScreen), `measure.StringWidth`,
`measure.ClustersOld`/`ClustersNew`, and the test VT (internal/ptytest/vt.go).
`Canvas.Set` re-clusters via `measure.Clusters`, but `Canvas.Row` wrote cell
text verbatim.

## 3. Implementation & Verification Plan
- Shared helper `measure.JoinsAfterZWJ(r)`: only letters, numbers,
  punctuation and symbols join after a ZWJ; controls, spaces, format runes and
  marks end the cluster and go through the normal parser/sanitiser. All sites
  above use it.
- Defense in depth: `Canvas.Row` drops C0/ESC/DEL/C1 runes from cell text
  (`safeCellText`); an emptied cell renders as a blank, keeping geometry.
- Tests (`zwj_control_test.go`): ZWJ + CSI, OSC 52, `\n`, `\r`, BEL, U+009B,
  DEL across ParseANSI, ParseANSIOld, ClipRow, Clusters*, Canvas.Write/
  WriteANSI + Row, raw cell text in Row; ZWJ emoji (👨‍👩‍👧, 🏳️‍🌈, ❤️‍🔥)
  stay one cluster of width 2.
