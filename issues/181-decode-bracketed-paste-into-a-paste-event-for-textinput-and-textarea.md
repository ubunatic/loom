# 181 — Decode bracketed paste into a paste event for TextInput and TextArea

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: event.go, textinput.go, textarea.go, issues/177, issues/180

---

## 1. Problem & Motivation
Pasting into a loom pane arrives as a stream of key events: newlines trigger Enter, and large pastes are slow. Bubble Tea and tview deliver paste as one event (issue 177 Findings).

## 2. Technical Specification / Findings
Enable bracketed paste (`ESC[?2004h`) in `Pane` and restore it on exit; decode `ESC[200~ … ESC[201~` into a `PasteEvent{Text string}`. `TextInput` inserts the text with newlines stripped or replaced by spaces; `TextArea` inserts it as-is as one undoable edit. Widgets that do not handle paste receive nothing new.

## 3. Implementation & Verification Plan
/goal Decode bracketed paste into `PasteEvent`, handle it in `TextInput` and `TextArea`, with decoder and widget tests and a docs/Widgets.md note; stop and report if a supported terminal breaks with bracketed paste enabled.
