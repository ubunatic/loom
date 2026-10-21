# Two-Agent Lean Sprint: Issue 353 (Ctrl+Space menu labels and hotkeys)

Date: 2026-10-21

Searcher skipped: the user's design file (docs/data/cspace-menu-001.ansi) fixed the target, and the
developer (d353, codex:sol:med, key routing per AGENTS.md) started with a read-only plan turn.

## Plan (1m24s, 63k new)

Found the bracket and width code, that the bar already auto-shows on selection, and that Link was
never wired (disabled, no handler). Host decision: Link stays disabled; a Link prompt is separate.

## M1 labels and hotkeys (3906319, 13m33s, 314k new)

Specced `popover_hotkeys` and `popover_hotkey_indices` (additive API). Dev q1 failed in two PTY tests
that still expected `Box` and `[ B`; the dev fixed them and their click helper, which had counted 🎨
as one column. Not rerun.

## M2 focus model (bcc267a, 8m35s, 414k new)

The user changed the model mid-run: Ctrl+Space focuses the bar, and letters act only on a focused
bar. The gate became `popoverFocusSet` alone. Dev q1 and host q1 both green. Live tmux: Ctrl+Space
then `b` bolds and saves; a selection-shown bar lets typing through.

## Observations

- The host's first focus decision ("letters work after Ctrl+Space, focused or not") let an unfocused
  bar take letters. The user caught it. Typing-safety choices are worth one line to the user before
  dispatch, even under "decide and report".
- The live check found 354: the first key after a lone Escape is swallowed. It predates 353.
