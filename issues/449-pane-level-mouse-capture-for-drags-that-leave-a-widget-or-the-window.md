# 449 — Pane-level mouse capture for drags that leave a widget or the window

**Status**: Closed — pane-level mouse capture delivered in 5547167a; View and Viewport drags use it; make test-q1 passed (developer run on the same tree, host re-ran go vet, go test . and apitest)
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 448

---

## 1. Problem & Motivation
A scrollbar drag should survive the pointer leaving the widget or the window and coming back, with the button still held. Today each widget tracks its own drag (`view.go`), so every new draggable (the TextEntry bar of 448, future bars) would repeat the rule or add a workaround.

## 2. Technical Specification / Findings
- No terminal-specific path is needed. The kitty protocol is keyboard-only; mouse input stays on `?1002`/`?1003` + `?1006`. Detecting kitty does not separate good terminals from bad ones (foot supports the kitty keyboard protocol too).
- `loom-probe mouse` results: kitty, alacritty, ghostty and ptyxis (VTE) keep sending drag reports clamped to the edge and send the release at the last inside position. foot 1.28 sends hover instead of drag once the pointer leaves, and no release at all (log: "stray button release event (compositor bug?)"). Whether foot or the compositor causes it is unverified.
- On foot a held button and a button released outside look the same (hover, no button info), so the SDK cannot tell them apart.
- Decision (decide mode): hover during a capture ends it. This is safe (no scrollbar stuck to the pointer) and means foot loses the drag on leaving, as old terminals do.

## 3. Implementation & Verification Plan
- Add a pane-level capture: a widget claims the mouse on press; drag and release go to it even when the pointer is outside its bounds (coordinates clamped or passed through, child-local per `docs/Widgets.md`). It ends on release, a new press, hover, focus loss or `Esc`.
- Move `View` drag handling onto it; `Viewport` and the TextEntry bar (448) reuse it. This touches event routing, so the ticket starts on `terra:med`.
- Tests: a scripted drag that leaves the widget bounds and returns keeps dragging; a release outside completes it; hover ends it without a stuck state; no change for widgets that never claim.
- Document in `docs/Widgets.md` and `docs/TuiInput.md`; confirm with `loom-probe mouse` in a real terminal, noting foot's expected drop.
- Optional follow-up: a foot bug report with the probe output.
