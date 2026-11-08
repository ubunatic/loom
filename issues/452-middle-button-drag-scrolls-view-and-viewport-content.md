# 452 — Middle-button drag scrolls View and Viewport content

**Status**: Closed — middle-button drag pan delivered in dcb3de70; make test-q1 passed (developer run on the same tree; host re-ran go vet, go test . and apitest)
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 449, 450, 451

---

## 1. Problem & Motivation
Owner request: scrolling by dragging with the middle button. Today only the scrollbar thumb is draggable (449/450); the content itself cannot be grabbed.

## 2. Technical Specification / Findings
- Decision (decide mode), revised after the owner's test: a scrub, not grab-and-drag. Pressing the middle button inside a scrollable `View` or `Viewport` starts it; the pointer row maps onto the whole scroll range (top row = first line, last row = end, the press row keeps its offset so nothing jumps; dragging down scrolls down); release ends it. The first version moved the content with the pointer one line per row, which the owner found inverted and too slow. It reuses the pane-level capture from 449, so the pan continues outside the widget and the window (not in foot, which turns the drag into hover and ends it).
- Esc cancels the pan and restores the scroll position, as for the scrollbar drag.
- The pointer reports whole rows, so the pan moves in whole lines; the half-row positions of 450 stay for wheel, keys and the thumb.
- A middle press on the scrollbar column keeps its current meaning (no pan).
- Child content that handles the middle button itself (for example a `Viewport` child) is not affected: a press the child consumes is not a pan. Mouse events reaching widgets are 0-based and child-local (docs/Widgets.md).
- Keyboard path: scrolling by keys already exists; Esc cancel needs its own test.

## 3. Implementation & Verification Plan
- Claim the capture on a middle press that lands on scrollable content; apply the row delta to `Scroll`/`ScrollY` on drag; end on release, press, hover, focus loss or Esc.
- Tests (mouse and key path): middle-press-drag-release scrolls by the row delta; leaves and re-enters the bounds; Esc restores; a left press does not pan; a middle press on the scrollbar does not pan; a child that consumes the middle press wins.
- Docs: `docs/Widgets.md` and `docs/TuiInput.md`; check in `loom widgets --show Viewport`.
