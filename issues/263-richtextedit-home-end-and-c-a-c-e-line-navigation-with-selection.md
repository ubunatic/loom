# 263 — RichTextEdit Home/End and C-a/C-e line navigation with selection

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature
**Related**: `richtextedit.go:342`, `event.go` (DecodeKey), issues 257 (decoder changes 40c153a), 088 (key capture)

---

/goal Make Home/End and C-a/C-e move to line start/end in RichTextEdit, with Shift variants extending the selection, verified by decoder + widget tests and a live gallery check; stop and report when blocked on a user decision.

Re-check live code and recent commits first.

## Requested

- `Home` / `C-a`: cursor to line start. `End` / `C-e`: cursor to line end.
- Shift variants select: `S-Home`, `S-End`, and `C-S-a` / `C-S-e` (CSI-u terminals).

## Findings at filing

- RichTextEdit already handles `home`/`shift-home`/`end`/`shift-end` (`richtextedit.go:342`), but the decoder seems never to emit `shift-home`/`shift-end`: xterm sends `ESC[1;2H` / `ESC[1;2F`, and `DecodeKey` only maps modifiers for A–D cursor keys and Insert/Delete. Likely why selection with Home/End doesn't work live. Fix in the decoder (library), with tests for `ESC[H`, `ESC[1;2H`, `ESC[1;5H`, `ESC[1~`, `ESC[1;2~`, and the F/end equivalents.
- `C-a` / `C-e` are not bound in RichTextEdit. Check that no app/host already uses `ctrl-a` as select-all in a way this would conflict with.

## Open

- Visual line vs. logical line when text wraps: use the logical line unless RichTextEdit wraps soft lines; document the choice.

## Reopened (2026-10-04): S-Home/S-End scroll the terminal instead of selecting

User report: in the live gallery, S-Home/S-End scroll the terminal; in nvim they select. User runs Tilix (VTE 8401).

Finding: VTE terminals (Tilix, GNOME Terminal) use Shift+Home/End (and Shift+PgUp/PgDn) to scroll their own scrollback while the app is on the primary screen, and never send the keys to the app. On the alternate screen (nvim, less) they pass through. `loom widgets --show` runs inline on the primary screen (`pane.go` ScreenMode, auto alt-switch), so the decoder fix from a7a26bb never receives these keys there.

### Next milestone

- Probe first (docs/Canary.md): confirm in Tilix that S-Home reaches the app in `ScreenAlt` and not inline (user runs the check by hand).
- Make text-editing demos able to get these keys: the gallery `--show` (at least for RichTextEdit/editors) runs on the alternate screen, or a widget can request alt screen when focused for editing. Fix it in the library/pane, not per demo.
- Document in docs/Widgets.md: on VTE inline, use C-S-a / C-S-e (CSI-u) or alt screen; S-Home/S-End are terminal-reserved there.
