# 464 — Text widgets share one paste sanitiser and RichTextEdit accepts paste

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 458, 462

---

## 1. Problem & Motivation
A bracketed paste carrying control characters lands differently in each text widget
(458 batch 4, `docs/studies/tuios/report-input.md`, payload with Ctrl+C, CR, ESC, Tab, CRLF, NEL):

- TextEntry normalises line breaks and strips controls (`text_entry.go` ~L1101-1125).
- TextInput replaces only line breaks and keeps `\x03` and `\x1b` in its value (`textinput.go` ~L54-61).
- TextArea inserts the raw text (`textarea.go` ~L81-96).
- RichTextEdit has no `ConsumePaste`, so a paste does nothing.

The terminal is safe: `Canvas.Write` drops control bytes and escape sequences from cell text
(canary: `a\x1b[31mb\x03c\x07d` flushes as `abcd`). The harm is in the widget value: an
application reads, saves or copies text with hidden ESC or Ctrl+C bytes. TUIOS `SanitizePaste`
drops C0/C1 except Tab, LF and CR.

## 2. Technical Specification / Findings
- Fix the library, not the caller: one shared sanitiser (the TextEntry rules, or a policy the
  widget picks: single-line vs multi-line) used by every widget that consumes paste.
- Single-line widgets fold line breaks as TextEntry does; multi-line widgets keep LF, map CRLF,
  CR and NEL to LF, and keep or expand Tab per the widget's Tab handling.
- RichTextEdit implements `ConsumePaste` with the same sanitiser and its own undo step.

## 3. Implementation & Verification Plan
/goal Every loom text widget stores a pasted payload without C0/C1 controls (except the line
breaks and Tab the widget supports), and RichTextEdit accepts paste; stop and report when blocked
on an API decision or denied permission, or when no progress is made (clear the goal and write a
failure report).

- Re-check live code first; line numbers above are from 458 batch 4.
- One table test per widget with the batch 4 payload, through the Pane paste route.
- Update `docs/Widgets.md` (paste section) and the `features.md` paste row in the TUIOS study.
