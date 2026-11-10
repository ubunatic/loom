# 470 — TextInput and TextArea get user selection, undo and cut like TextEntry

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 458, 469, 464, 465, 467, 433

---

## 1. Problem & Motivation
loom's text widgets are not equally desktop-like. TextEntry and RichTextEdit have Shift
selection, mouse selection, undo/redo, cut and copy. TextInput and TextArea have only
`SetSelection` from code: no keyboard or mouse selection, no undo, no cut, no `ConsumeMouse`
(`textinput.go:31-155`, `textarea.go:38-343`; 458 batch 6, `docs/studies/tuios/report-widgets.md`).
Their Ctrl+C also differs: TextInput copies the whole value when nothing is selected, while
TextEntry passes Ctrl+C through (`textinput.go:82-90` vs `text_entry.go:533-534`; canary). A
masked TextInput refuses copy, which is right and must stay. In TextArea Ctrl+A moves to the
line start (`textarea.go:303-305`), not select-all. `spec/defaults.yaml` (lines 36-56) scopes
"no undo" to TextInput and TextArea, so the spec matches the code today.

## 2. Technical Specification / Findings
- Decide first (owner): give TextInput and TextArea the TextEntry editing core (shared
  selection, undo and clipboard code), or make TextEntry the single-/multi-line editor and keep
  TextInput/TextArea as thin wrappers. Do not copy the code a third and fourth time.
- Same key rules everywhere: Ctrl+C copies only with a selection and passes through otherwise;
  Ctrl+A selects all; undo/redo keys from `spec/defaults.yaml`; update the spec rows.
- Mouse selection reuses the drag ownership from 467.

## 3. Implementation & Verification Plan
/goal TextInput and TextArea offer keyboard and mouse selection, undo/redo and cut with the same
keys and Ctrl+C rule as TextEntry, through shared code, with tests; stop and report when blocked
on the owner's design choice or a denied permission; on a loop or no progress, clear the goal and
write a failure report.

- Re-check live code first; 464, 465 and 467 may change paste, clipboard and drag.
- Table tests per widget for select, undo, cut and copy with and without selection, masked copy.
