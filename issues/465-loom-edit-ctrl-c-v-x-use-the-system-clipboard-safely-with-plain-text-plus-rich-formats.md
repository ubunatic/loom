# 465 — loom edit Ctrl+C/V/X use the system clipboard safely, with plain text plus rich formats

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 464, 194, 257, 441, 415

---

## 1. Problem & Motivation
Owner request: `loom edit` Ctrl+C, Ctrl+V and Ctrl+X should work with the system clipboard,
safely. Idea: copy a compound clip, plain text plus a rich attachment (`.ansi`, HTML or
Markdown), the way browsers offer copied content as text or as rich text.

Today (checked at `7ecfba5e`):
- RichTextEdit copy and cut fill only an in-memory `[]RichLine` (`richtextedit_clipboard.go`
  `copySelectionOrWord`); paste inserts only that (`richtextedit_input.go` ~L170). Nothing
  leaves or enters the app. Pane binds OSC 52 writes only for `SetClipboardWriter` /
  `setClipboardWriter` widgets (`clipboard.go` `bindClipboardTree`); TextEntry has it,
  RichTextEdit does not.
- The terminal's own paste (bracketed paste) is ignored by RichTextEdit; ticket 464 adds
  `ConsumePaste` with a shared sanitiser.

## 2. Technical Specification / Findings
- **Copy/cut**: write plain text through the existing Pane OSC 52 path (works over SSH, needs no
  tool), and keep the styled `[]RichLine` internally as today.
- **Paste**: Ctrl+V pastes the internal rich clip when it still matches the system clipboard,
  else the system text. Ctrl+Shift+V is the terminal's paste (bracketed, 464 path) and stays
  untouched; Ctrl+V reads the desktop clipboard itself (owner design, 466).
- **Fallback (owner, 2026-11-10)**: when no system clipboard is available (no desktop session,
  e.g. a server over SSH), Ctrl+C/X/V fall back to loom's internal copy buffer instead of showing
  a notice, so copy and paste inside the app keep working. Copy/cut still also send OSC 52, which
  can reach the user's local clipboard over SSH.
- **Rich formats (idea, unverified)**: OSC 52 carries one text payload, no MIME types. A
  multi-format clip (text/plain plus text/html, text/markdown or an `.ansi` type) needs a local
  clipboard API: Wayland data-control or `wl-copy --type`, X11 `xclip -t`, macOS pasteboard.
  Check first which of these can offer several types in one clip; `wl-copy` may offer only one
  type per call. HTML output can reuse 415 (`WriteHTML`), Markdown 441.
- **Safety**: sanitise every incoming clip (464 rules) and cap its size; never run a shell, pass
  data to helpers on stdin with fixed argv; no OSC 52 read without an explicit opt-in (terminals
  differ and it can leak); a missing helper falls back to OSC 52 text with no error dialog.

## 3. Implementation & Verification Plan
/goal In `loom edit`, Ctrl+C/X put the selection on the system clipboard as plain text and
Ctrl+V pastes from it safely, with styled round-trips inside loom kept; rich formats land as a
second milestone only if the canary shows a working multi-type offer. Stop and report when
blocked on a format or platform decision or a denied permission; on a loop or no progress, clear
the goal and write a failure report.

- M1: OSC 52 copy/cut for RichTextEdit; paste rule above; tests with a recording clipboard
  writer; owner check in foot and kitty (copy in loom edit, paste in another app).
- M2 (canary first): probe multi-type clipboard offers on the owner's Wayland session; if one
  works, add text/html or text/markdown alongside text/plain, else record the result here.
- Re-check live code first; 464 may have landed and changed the paste path.
