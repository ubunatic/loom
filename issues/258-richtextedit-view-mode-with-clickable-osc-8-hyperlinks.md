# 258 — RichTextEdit view mode

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature
**Related**: `richtextedit.go`, `gallery/richtextedit.go`, issues 254, 257

---

/goal Add a read-only view mode to RichTextEdit with link spans styled as links (not yet clickable), show it in the gallery, verify via `make test-q1`, or stop and report when blocked on a user decision or denied permission.

Re-check live code and recent commits first.

## Requested behavior

- **View mode**: a RichTextEdit option/mode where editing is disabled (no text input, no format popover). Selection/copy may stay if cheap.
- **Links**: a link span type carrying a URL, rendered with a link style. Clickable OSC 8 output is moved to issue 259 (human-assisted).
- **Gallery**: the RichTextEdit gallery demo includes `https://ubunatic.com/loom` as an example link.

## Notes / uncertainties

- OSC 8 emission is out of scope here (issue 259); keep the URL on the span so 259 can use it.
- RichTextEdit may not have a link span type yet; adding one (and whether edit mode can create links) is in scope only as far as view mode needs.
