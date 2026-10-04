# 263 — RichTextEdit Home/End and C-a/C-e line navigation with selection

**Status**: Closed — decoder now keeps modifiers on Home/End (ESC[1;<m>H/F and tilde forms); RichTextEdit binds C-a/C-e and shift variants; Upgrading row (a7a26bb). Host reran make test-q1 green; live tmux: C-e/C-a moved to line end/start, End+S-Home selected the line, C-b bolded it
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
