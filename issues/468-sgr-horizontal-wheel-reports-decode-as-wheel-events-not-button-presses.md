# 468 — SGR horizontal wheel reports decode as wheel events, not button presses

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 458, 451

---

## 1. Problem & Motivation
Touchpads and tilt wheels send horizontal scroll as SGR buttons 66 (left) and 67 (right). loom's
decoder handles only 64 and 65 as wheel events; 66 and 67 fall through to `MousePress` with
buttons 2 and 3 (`event.go` ~L538-556, `case scroll && btn == 0/1`). 458 batch 5 confirmed it in
a Pane PTY canary: `\e[<66;45;15M\e[<67;45;15M` reached TextEntry as two presses
(`docs/studies/tuios/report-mouse.md`). So a sideways swipe acts like a right-click or a
middle-click: it can start a pan or open a context menu. ultraviolet and tcell decode both axes.

## 2. Technical Specification / Findings
- Add `MouseScrollLeft` / `MouseScrollRight` (or one wheel action with a direction) and decode
  64-67 by `cb & 3` with the scroll bit, keeping modifiers.
- Widgets that do not scroll sideways ignore them; widgets that do (wide Table, TextArea without
  wrap, Viewport) scroll horizontally. Shift+vertical wheel as horizontal is a separate choice.
- The wheel burst curve (`pane_wheel.go`, 451) should apply per axis. TUIOS also locks to the
  majority axis for a 250 ms gesture (`wheel_axis.go`); consider it only after both axes decode.
- New public constants change the API (P3 while alpha; add a row to `docs/Upgrading.md` if a
  caller-visible meaning changes).

## 3. Implementation & Verification Plan
/goal SGR buttons 66 and 67 arrive as horizontal wheel events in every widget, never as presses,
and scrolling widgets with horizontal overflow move sideways; stop and report when blocked on an
API decision or denied permission; on a loop or no progress, clear the goal and write a failure
report.

- Re-check live code first; decoder table test for 64-67 with Shift/Alt/Ctrl; Pane route test
  that a horizontal report never produces `MousePress`.
- Owner check: a sideways touchpad swipe in foot and kitty over a wide Table.
