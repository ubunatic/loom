# 467 — TextEntry selection drag owns the pointer: release outside, focus loss and auto-scroll while held outside

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: 458, 449, 452, 448, 206

---

## 1. Problem & Motivation
A desktop-like text field keeps a selection drag going when the pointer leaves it: the selection
follows the pointer, it scrolls while the button is held outside, and releasing anywhere ends the
drag. loom's TextEntry does not. 458 batch 5 (`docs/studies/tuios/report-mouse.md`, headless Pane
PTY canary at `7ecfba5e`) found:

- **Selection does not capture.** `mouseCaptureTarget` claims capture only for
  `barDrag.active || pan.active`, not selection `dragging` (`text_entry.go` ~L1237-1245). A drag
  into the sibling goes to the sibling; outside the window nobody gets it; the selection stops.
- **Release outside leaves `dragging=true`**, so the next motion over the entry may extend a
  selection the user already ended.
- **Focus loss does not end a drag.** `SetFocus(false)` sets `e.focused = false` before testing
  `if e.focused`, so `endMouseCapture(true)` never runs (`text_entry.go` ~L338-347). Canary: pan
  stayed active after focus left and kept scrolling the unfocused entry.
- **No auto-scroll.** Held still outside, zero events arrive over 300 ms. TUIOS starts a 50 ms tick
  on the first outside motion, moves two steps per tick, and stops on release or re-entry
  (report-mouse.md, `mouse_motion.go` / `update.go` permalinks).

## 2. Technical Specification / Findings
- Fix the library, not the caller: selection drag claims the same Pane capture as pan and the
  scrollbar, so drag and release reach the entry with child-local coordinates anywhere.
- Fix the `SetFocus` order; focus loss cancels any capture (pan, bar, selection), as Esc does.
- Auto-scroll: while captured and the pointer is above/below (or left/right of) the content,
  scroll on a tick and extend the selection; stop on release, re-entry, Esc or focus loss. The
  tick must come from loom's ticker path (`widget.go` `tickTree`), not sleeps; check whether a
  shared helper fits TextArea/RichTextEdit/List later.
- Real terminals clamp a drag outside the window to the edge cell (449: kitty, alacritty,
  ghostty, ptyxis via `loom-probe mouse`). Owner check 2026-11-10 in foot, raw SGR log: drags
  clamp to x=1 at the left edge; two of three releases outside arrived at the clamped edge, one
  never arrived (foot logged "stray button release event"); held still at the edge, no report
  comes for 0.4-0.6 s. So auto-scroll needs the tick, an edge cell counts as "outside" for a
  widget at the window edge, and a lost release must not leave the drag on: the next hover
  without a button (cb 35) ends it. Confirmed the same day with `loom-probe mouse` in foot: the
  owner saw the drag lost, and foot logged five "stray button release event" warnings; kitty and
  the other terminals in 449 keep the drag.
- Event-routing ticket: start on terra:med (project rule).

## 3. Implementation & Verification Plan
/goal A TextEntry selection drag keeps following the pointer outside the entry, auto-scrolls while
held outside, and ends on release anywhere, Esc or focus loss, with tests; stop and report when
blocked on an API decision or denied permission; on a loop or no progress, clear the goal and
write a failure report.

- Re-check live code first (capture, `SetFocus`, ticker).
- Port the batch 5 canary rows (nested entry at nonzero origin, sibling, outside the window,
  release outside, focus change mid-drag, Esc) into Pane tests; add a tick-driven auto-scroll test
  with a fake clock (no sleeps).
- Owner check in foot and kitty: drag-select out of an entry, hold, release outside.
