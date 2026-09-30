# Loom developer prompt (lean sprints)

Append to the ticket instruction when starting a developer agent, e.g.
`harnez agent start --role developer -p "Work on issue N. $(sed 1,4d .harnez/prompts/developer.md)"`.
Replace <commit trailers> with the host session's attribution lines.

Check live code and recent commits first, then implement the ticket's goal with TDD. No breaking API change; if one is unavoidable, stop and report. Use harnez read -L <from:to> for bounded reads.
If you add or visibly change a widget, add or update its demo in the gallery/ package (shown by `loom widgets --show <Name>`) and save its render as docs/progress/<Widget>.ansi; add an app under examples/ only if the widget needs a realistic context.
Keep spec/widgets.yaml in strict name order and list every exported widget there.
The key decoder emits `pgdown`/`pgup`; also accept `pgdn`, `pagedown`, `pageup` in bindings.
Mouse events reaching widgets are 0-based and child-local; never subtract 1. Mouse handlers must return loom.EventResult (ConsumeMouse or ConsumeMouseEvent); a (quit, consumed bool) mouse handler is never called.
In tests, locate screen text by rune column or display width, never byte offset.
Tests must not depend on the caller's color env (pin LOOMCOLOR=truecolor in TestMain when asserting colours).
Never change a test expectation just to make it pass without stating why the old one was wrong.
Finish all code first, then run make test-q1 exactly once as the last step, writing output to a file, and grep it for '--- FAIL'; never interrupt it; if it fails, report instead of rerunning.
Commit only your code changes with a conventional commit ending in: <commit trailers>. Do not edit the issue file. Run make install. Report the commit hash.

- Avoid hacks: fix the library, not the caller. Change interfaces or the event approach only when needed, and then follow a proven key/mouse model.
- Keyboard and mouse fixes must cover both the key and the mouse path, with a test for each.
