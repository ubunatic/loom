# 351 — Scripted live check: run installed loom in tmux, wait for render, send keys, print screen

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Tooling
**Related**: 342, 346, 347 (live checks done by hand), docs/LeanSprints.md (2026-10-20/21 section)

---

/goal Provide one repo script (plus a self-documented make target) that runs the installed `loom`
command in a detached tmux server, waits until a given text is rendered, sends a key list step by
step and prints the screen (plain or with SGR) after each step, then kills its server; document it
in docs/Testing.md and docs/LeanSprints.md; or stop and report when blocked on a user decision or
denied permission.

## Problem

Hosts verify user-visible behaviour live (lean-sprint rule) and rebuilt the same tmux one-liners
for every check in 342/346/347. Pitfalls hit on the way:
- Keys sent before the editor rendered were echoed raw (`^@^[[C`) and lost.
- `sleep` is blocked in agent shells; waits must poll the pane for a text with a timeout.
- The popover bar needs an arrow towards it before Left/Right move inside it (342/316 semantics),
  so key lists like "Ctrl+Space, 4×Right" silently did something else.
- Filtering colored capture output with greedy greps hid the rows of interest.

## Sketch (to confirm in a plan turn)

- `scripts/live-tmux.sh` (Bash rules in docs/Bash.md):
  `live-tmux.sh [-x 90 -y 20] [-e] [--rows 4:7] --wait <text> -- <command...>` reading steps from
  stdin, one per line: `keys <tmux keys...>`, `wait <text>`, `snap [label]`.
- Unique tmux socket per run (`-L loom-live-$$`), always killed via trap; poll interval and timeout
  as flags; nonzero exit and a clear message on wait timeout.
- `make live-check` (⚙️ phony, `##` help) as a thin wrapper; no use in the test tiers (not a test).
- Example in docs: open the 🎨 picker in `loom edit` and snapshot FG and BG mode.

## Verification

Run the docs example against the installed binary; a deliberate wrong wait text exits nonzero
within the timeout and leaves no tmux server (`tmux -L <socket> ls` fails).

## Findings from 359–361 live checks (host, 2026-10-22)

- `tmux send-keys S-Right` did not select text in `loom edit` (bar never showed a selection), so
  coloured-selection cases could not be checked live; find a working key encoding or a mouse drag.
- The useful assertion form was a per-cell dump of FG/BG codes (`capture-pane -pe` with escapes made
  visible) compared with the design sketch; plain-text captures hid every look bug in 359/360.
- (2026-10-22, 354) `timeout 0.1 true` does not wait; waits must really pause (e.g.
  `timeout 0.8 tail -f /dev/null`) or poll the pane. Compare screens with `loom diff` (362).
