# 210 — loom info: terminal and capability report with --watch mouse probe

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 204 mouse hit-testing, 206 mouse capture

---

## 1. Problem & Motivation
Debugging mouse and rendering issues (204, 206) needs a quick view of what loom sees in the current terminal. User request: add `loom info`, plus a live `loom info -w|--watch` mode that makes mouse-pointer vs target-cell discrepancies visible.

## 2. Technical Specification / Findings
- `loom info`: one-shot report: loom version, terminal size (cells, and pixels if known), TERM/COLORTERM/TERM_PROGRAM, colour depth, detected capabilities (mouse modes, true colour, Unicode/emoji width, images/graphics protocol, etc.; reuse existing detection code).
- `loom info -w|--watch`: full-screen live view; values refresh when they change (resize, theme/capability changes). It captures the mouse and shows the pointer x/y (0-based cell and raw report). It draws a `+` crosshair at the target cell and follows the pointer, only inside the panel area where mouse input is accepted, so the user can compare the real pointer with the cell loom resolves. q/Esc quits.
- Use the library event model (ConsumeKey/ConsumeMouse returning EventResult, docs/Widgets.md); no special mouse path in the command.

## 3. Implementation & Verification Plan
- Unit test for the report fields; PTY test against the real binary: `loom info` prints the version and size; in `--watch`, a mouse report at (x,y) shows those coordinates and a `+` at that cell; a resize updates the size.
- Document in the loom CLI docs/man page.

/goal `loom info` and `loom info --watch` work and are proven by PTY tests; or stop and report when blocked on a user decision or denied permission.
