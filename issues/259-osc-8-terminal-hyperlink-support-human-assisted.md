# 259 — OSC 8 terminal hyperlink support (human-assisted)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature
**Related**: issue 258 (RichTextEdit view mode), `richtextedit.go`, `gallery/richtextedit.go`

---

/goal Make hyperlinks clickable via OSC 8 in RichTextEdit view mode, with a human verifying clicks in real terminals; stop and ask the human before each terminal probe and when blocked.

Split out of 258. A human assists: clickability can only be checked by hand in real terminals.

## Scope

- Emit OSC 8 (`ESC ]8;;URL ESC \ text ESC ]8;; ESC \`) for link spans.
- Likely belongs in the renderer/cell layer ("fix the library, not the caller"): cells need a link attribute, and the screen diff must open/close links correctly across cells and redraws.
- Gallery RichTextEdit demo shows `https://ubunatic.com/loom` as a clickable link.
- Probe target terminals (kitty, foot, gnome-terminal, tmux passthrough) with the human first, per `docs/Canary.md`.

## Jules audit (2026-11-03, GitHub main ab547ca)

Verdict: still valid (4/4 refs checked). `ansi.Style` (`ansi/style.go:227`) and `Cell` (`canvas.go:30`) carry no link attribute, and `AppendANSIFor` (`ansi/style.go:259`) emits SGR only; no OSC 8 code exists. Starting point: a `Link` field on `ansi.Style`, emitted by `AppendANSIFor` when it changes, with `RichSpan.Link` mapped onto it. Correction: the gallery demo links `https://loom.dev` (`gallery/richtextedit.go:22`), not `https://ubunatic.com/loom`. Report: `reports/jules/259.md`.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| audit | jules-audit-259 | 5441328c-20f4-451e-b921-f67587c9d08d | - | failed: ended AWAITING_USER_FEEDBACK, no PR, not answered |
| audit | jules-audit-259b | 11861938100957675957 | GH #74 | kept: still valid, 4/4 refs hold; replies: 1 |
