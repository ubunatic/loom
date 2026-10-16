# 321 — Opt-in kitty keyboard protocol so Shift+Enter and Ctrl+J differ from Enter

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Feature
**Category**: Input / Key Decoding
**Related**: [317](317-distinguish-ctrl-i-from-tab-in-terminal-input-and-key-decoding.md) (Ctrl+I vs Tab: CSI-u would let Ctrl+I differ from Tab), harnez #738

---

## 1. Problem & Motivation
Enter and Ctrl+J reach Loom as the same key, and Shift+Enter as plain Enter.
Chat inputs need a newline key distinct from send; harnez tui fell back to
Alt+Enter and a trailing `\` (harnez #738).

## 2. Technical Specification / Findings
- `decodeCSIu` (`event.go`) handles letters, space and tab, not Enter (13).
- Loom never asks the terminal for the kitty protocol, so CSI-u is not sent.
- Add an opt-in Pane setting that pushes progressive enhancement flag 1
  (`CSI > 1 u`) on start and pops it (`CSI < u`) on close and on Suspend (319).
- Decode `CSI 13;<mod> u` as `shift-enter` / `ctrl-enter`; extend CSI-u to
  other keys the flag changes (Escape, Backspace), keeping legacy names for
  unmodified keys.
- Terminals without support ignore the push; behavior stays as today.

## 3. Implementation & Verification Plan
- Decoder tests for CSI-u Enter/Escape/Backspace with each modifier.
- PTY test: push on start, pop on close and on panic exit.
- Manual check in kitty, foot, WezTerm, Ghostty and a legacy terminal; record in the ticket.
