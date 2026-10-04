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
