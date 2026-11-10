# 466 — TextEntry and loom edit handle a pasted image as an [Image #n] token that submits as a Markdown link

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 465, 464, 441, 417

---

## 1. Problem & Motivation
Owner request: Ctrl+V of an image in `loom.TextEntry` and `loom edit` should be handled
gracefully. Keep only a link, as plain text or a Markdown image link. TextEntry shows a short
token such as `[Image #1]`; on submit the full Markdown link goes out with the text.

Today (checked at `7ecfba5e`): pastes are text only. TextEntry inserts the pasted text and
`OnSubmit` gets a plain string (`text_entry.go` ~L37). RichTextEdit ignores the terminal paste
(464) and pastes only its internal clip (465). RichTextEdit already has atomic inline tokens,
`RichSpan.PillData` / `RichPill{Kind, ID, Metadata}` (`richtext.go` ~L23-36), selected and deleted
as one unit (`expandRangeForPills`). TextEntry has no such tokens.

## 2. Technical Specification / Findings
- **What a terminal delivers (unverified, check first)**: a bracketed paste carries text only.
  An image copied from a file manager likely arrives as a path or `file://` URI (text/uri-list);
  a pure image clip (screenshot) likely arrives as nothing. Image bytes need a local clipboard
  read (`wl-paste --list-types`, then `--type image/png`), which belongs to the 465 clipboard
  layer, or a terminal protocol (kitty OSC 5522, unverified).
- **Detect**: a paste that is one image path/URI (by extension or MIME type), or an image clip
  read through 465, becomes an image token instead of raw text.
- **Represent**: one token per image, label `[Image #n]` (n per widget, counting up), data holds
  the target (path, URI or a saved file the app chooses) and alt text. RichTextEdit reuses
  `RichPill{Kind: "image"}`; TextEntry needs the same atomic token (edit, select, delete and
  undo as one unit).
- **Submit/serialise**: tokens expand to `![Image #n](<target>)`; a plain-text mode gives the bare
  target. TextEntry needs a submit value with tokens expanded (keep `OnSubmit(string)` getting the
  expanded text, or add a rich variant; API stability is P3). Markdown output ties to 441.
- **Where bytes go**: loom must not decide storage. The app supplies a hook that turns image
  bytes into a target (save to a temp dir, upload); without a hook, a byte-only image clip shows a
  short notice and inserts nothing.

## 3. Implementation & Verification Plan
/goal Pasting an image path or image clip into TextEntry or `loom edit` inserts one atomic
`[Image #n]` token, and submit or save yields the Markdown image link with the surrounding text;
plain text pastes are unchanged. Stop and report when blocked on an API or storage decision or a
denied permission; on a loop or no progress, clear the goal and write a failure report.

- Canary first: record what foot 1.27.0 and kitty 0.47.1 send when pasting (a) an image file
  copied in the file manager, (b) a screenshot clip; an owner check, one step at a time.
- Shared token type for TextEntry and RichTextEdit; table tests for detect, edit-as-unit,
  undo, submit expansion; Pane paste route test.
- Re-check live code first; 464 and 465 may have changed the paste and clipboard paths.
