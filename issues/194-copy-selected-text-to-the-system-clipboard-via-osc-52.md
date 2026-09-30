# 194 — Copy selected text to the system clipboard via OSC 52

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: issues/180 (roadmap, decided 2026-09-30), issues/177, issues/179

---

## 1. Problem & Motivation
TextInput and TextArea cannot put text on the system clipboard. OSC 52 is an escape sequence that asks the terminal to set the clipboard; it needs no dependency and fits loom's plain-terminal design. Pasting is already covered by 181 (bracketed paste).

## 2. Scope
- A small helper that writes OSC 52 (base64) through the pane's output; no reading from the clipboard.
- TextInput and TextArea: a copy action (via KeyMap from 158 when present) copies the selection, or the whole value when nothing is selected.
- Terminals that ignore OSC 52 must see no garbage.

## 3. Implementation & Verification Plan
/goal Copy from TextInput and TextArea sets the system clipboard via OSC 52, with a unit test on the exact emitted bytes. Stop and report when blocked on a user decision or denied permission.
