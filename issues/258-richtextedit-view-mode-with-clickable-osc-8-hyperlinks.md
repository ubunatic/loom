# 258 — RichTextEdit view mode with clickable OSC 8 hyperlinks

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature
**Related**: `richtextedit.go`, `gallery/richtextedit.go`, issues 254, 257

---

/goal Add a read-only view mode to RichTextEdit that renders hyperlinks as clickable terminal links, show it in the gallery, verify via `make test-q1`, or stop and report when blocked on a user decision or denied permission.

Re-check live code and recent commits first.

## Requested behavior

- **View mode**: a RichTextEdit option/mode where editing is disabled (no text input, no format popover). Selection/copy may stay if cheap.
- **Hyperlinks**: link spans emit the OSC 8 terminal hyperlink sequence (`ESC ]8;;URL ESC \ text ESC ]8;; ESC \`) so terminals make them clickable.
- **Gallery**: the RichTextEdit gallery demo includes `https://ubunatic.com/loom` as an example link.

## Notes / uncertainties

- No OSC 8 support exists in loom yet (no hits in Go code). It likely belongs in the renderer/cell layer, not just the widget ("fix the library, not the caller"): cells need a link attribute, and the screen diff must open/close links correctly across cells and redraws.
- RichTextEdit may not have a link span type yet; adding one (and whether edit mode can create links) is in scope only as far as view mode needs.
- Probe target terminals (kitty, foot, gnome-terminal, tmux passthrough) before building on it, per `docs/Canary.md`.
