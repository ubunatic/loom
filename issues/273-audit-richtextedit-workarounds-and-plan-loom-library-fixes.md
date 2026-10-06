# 273 — Audit RichTextEdit workarounds and plan Loom library fixes

**Status**: Closed — user runbook passed
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

---

## 4. Workaround Inventory

| File:Line | Description / Workaround | Proposed Library Owner | Linked Ticket |
|---|---|---|---|
| `richtextedit.go:58-59`, `260-269`, `720-732`, `910-922` | **Local Modal Stack Management**: RTE maintains local `helpPopup`, `savePopup`, `savePicker` fields, measures popup dims during `Draw`, and manually routes input to popups ahead of itself. | `Canvas` / `Pane` Overlay Modal Manager | [#274](274-overlay-and-modal-layer-for-canvas-to-eliminate-local-popup-stacks.md) |
| `popup.go:58-61` | **Focus Side-Effects in Draw**: `Popup.Draw` calls `SetFocus(p.Open)` during rendering. | Focus Manager & Modal Controller | [#275](275-decouple-focus-state-updates-from-popup-and-widget-draw-rendering.md) |
| `gallery/richtextedit.go:52-70` | **Manual Mouse Coordinate Math**: Demo manually calculates header offset (`w.viewH`), subtracts `mouse.Y -= w.viewH`, and handles out-of-bounds mouse dispatch to child. | Container Layout (`VStack`, `Split`, `Viewport`) | [#276](276-container-owned-coordinate-translation-and-mouse-event-clipping.md) |
| `filepicker.go:295-303`, `choice.go:627` | **Mouse Hover Focus / Selection Shift**: `FilePicker` shifted input focus on `MouseHover`; `Choice` shifted selection index on hover. | Loom Core Input Invariants | [#277](277-standardize-click-only-focus-and-cursor-invariants-across-compound-widgets.md) |
| `richtextedit_filebar.go:39-65` | **File Bar Menu Bar Handoff**: Manual management of `MenuBar` focus, action callbacks, and keyboard shortcuts inside RTE. | Layout Window / Frame Container | [#274](274-overlay-and-modal-layer-for-canvas-to-eliminate-local-popup-stacks.md) |
| `filepicker.go:181-186` | **Compound Widget Sub-Field Focus Desync**: `FilePicker.activate()` set `p.nameFocus = true` without calling `updateFocus()`. | Compound Widget Focus Encapsulation | [#277](277-standardize-click-only-focus-and-cursor-invariants-across-compound-widgets.md) |

---

## 5. Ranked Plan for Library Improvements

1. **Rank 1 (P1 / Stability) — Click-Only Focus & Cursor Invariants ([#277](277-standardize-click-only-focus-and-cursor-invariants-across-compound-widgets.md))**:
   Ensure no compound input or selection widget changes active keyboard focus or hardware terminal cursor on `MouseHover`. Require `MousePress` or explicit keys.

2. **Rank 2 (P2 / Architecture) — Decouple Focus Mutations from Draw ([#275](275-decouple-focus-state-updates-from-popup-and-widget-draw-rendering.md))**:
   Eliminate `SetFocus` calls inside `Popup.Draw` and widget render methods. Move focus state updates into explicit lifecycle event handlers.

3. **Rank 3 (P2 / Ergonomics) — Canvas Overlay Modal Layer ([#274](274-overlay-and-modal-layer-for-canvas-to-eliminate-local-popup-stacks.md))**:
   Introduce a shared overlay/modal layer on `Canvas`/`Pane` so widgets do not need local `savePopup`/`helpPopup` state fields or manual modal key routing.

4. **Rank 4 (P3 / Refactor) — Container-Owned Mouse Coordinate Translation ([#276](276-container-owned-coordinate-translation-and-mouse-event-clipping.md))**:
   Standardize coordinate translation and event clipping across `VStack`/`Split`/`Viewport` containers so host gallery demos do not perform manual `mouse.Y` subtraction.

---

## Delivered Notes

- Completed audit across `RichTextEdit`, `FilePicker`, `Popup`, and `gallery/richtextedit.go`.
- Filed four follow-up library tickets: [#274](274-overlay-and-modal-layer-for-canvas-to-eliminate-local-popup-stacks.md), [#275](275-decouple-focus-state-updates-from-popup-and-widget-draw-rendering.md), [#276](276-container-owned-coordinate-translation-and-mouse-event-clipping.md), and [#277](277-standardize-click-only-focus-and-cursor-invariants-across-compound-widgets.md).
- Resynced index via `harnez index`.
