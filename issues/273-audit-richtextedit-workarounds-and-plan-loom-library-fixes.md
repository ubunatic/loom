# 273 — Audit RichTextEdit workarounds and plan Loom library fixes

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Research
**Related**: [269](269-refine-richtextedit-hints-picker-focus-and-help.md), [272](272-richtextedit-save-as-picker-reopen-enter-on-file-focus-follows-mouse.md), [225](225-modularize-high-loc-core-components-and-example-packages.md)

---

## 1. Problem & Motivation
RichTextEdit (about 5,600 lines across `richtextedit*.go`) has grown local fixes for things the library should handle: popups, focus, coordinates, key routing. Each new bug (see 269, 272) adds another one. The user asks to find all of them and plan how Loom itself should improve, so widgets stop needing them.

## 2. Technical Specification / Findings
Candidates seen while filing 272 (unverified, starting list only):
- RTE keeps its own modal stack: `helpPopup` / `savePopup` / `savePicker` fields, manual nil-ing after `Open` turns false, and an Escape special case that bypasses the popup and goes to the picker.
- `Popup.Draw` sets the inner widget's focus during drawing (`SetFocus(p.Open)`).
- `FilePicker.ConsumeMouse` changes focus on hover; `activate()` changes `nameFocus` without `updateFocus()`.
- Gallery RTE demo (`gallery/richtextedit.go`) shifts mouse Y by hand and routes out-of-rect events to the editor.
- File bar menu focus handled inside RTE `ConsumeKey`/`ConsumeMouse`.

Deliverable: a table of every workaround (file:line, what it works around, which library part should own it) and a plan grouped into library tickets, e.g. a shared overlay/modal layer, focus that changes only on click/keys, and container-owned coordinate translation. Follow the rule "fix the library, not the caller" and a proven key/mouse model (global or local routing). Event-routing work starts on `codex:sol:med`.

## 3. Implementation & Verification Plan
/goal Produce the workaround inventory and a ranked plan, and file one library ticket per proposed change, linked here; no production code changes in this ticket; stop and report when blocked on a user decision or denied permission.

Acceptance: inventory and plan recorded in this ticket (or a linked doc under `docs/`); follow-up tickets filed and listed; user reviews the plan before any implementation starts.
