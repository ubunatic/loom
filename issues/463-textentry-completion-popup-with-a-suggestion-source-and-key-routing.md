# 463 — TextEntry: completion popup with a suggestion source and key routing

**Status**: Draft
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 417, 418, 423, 441, 453, 456, 274

---

## 1. Problem & Motivation
`TextEntry` has submit and newline keys, selection, placeholder, scrolling and a highlight hook, but no completion support. Taia's chat input needs completion for slash commands, `@` files, `/model` and skills. Taia builds all of it in its own code: `completionPopup` (`internal/tui/completion.go`), `updateCompletion`, `acceptCompletion`, `enterCompletion`, `clearCompletion`, and a completion key table in `Chat.ConsumeKey` that runs before the input. Every chat app built on `TextEntry` would repeat this: where the popup anchors, how Up/Down/Tab/Enter/Esc are split between popup and entry, and how an accepted item replaces a range as one undo step.

Taia also needed Unicode-safe range replacement with undo (`TestCompletionUnicodeReplacementUndo`), which the library should own.

## 2. Technical Specification / Findings
Proposed shape, not final:
- A suggestion hook on `TextEntryOptions`, for example `Complete func(text string, caret int) Completion`, returning items, the replaced unit range and a trigger rune.
- The popup is anchored to the caret row, inside the entry bounds or as an overlay. Layout follows the existing popup rules (`popup.go`, issue 274).
- While the popup is open: Up/Down move the selection, Tab/Enter accept, Esc closes. Every other key returns `Ignored()`, so the entry and parent keep their behaviour. While closed, the entry handles keys as now.
- Accepting replaces the range through one undoable edit, on the same path as `ReplaceRange` (issue 418).
- Public accessors for popup state (open, items, selected) so parents can render hints or add modal rules.

Open questions:
- Whether the popup lives inside `TextEntry` or is a separate widget that `TextEntry` hosts.
- How a parent overrides Enter for submit versus accept. Taia needs `/command` Enter to submit, and `/model` picks to apply on Enter.
- Whether Up/Down belong to the popup in a multi-line entry.

## 3. Implementation & Verification Plan
- Start on `terra:med`: this ticket changes key routing (AGENTS.md).
- Unit tests: open and close; Up/Down/Tab/Enter/Esc with and without the popup; Unicode range replacement and undo; `Ignored()` for keys the popup does not use.
- Gallery demo with a completion source (see 423).
- Docs: the `docs/Widgets.md` TextEntry key table, and `docs/Upgrading.md` only if an existing key changes meaning.
- Taia validates by moving its completion code onto the new API.

/goal Add a completion popup to loom.TextEntry: a suggestion hook, popup key routing that returns Ignored() when closed, one-step undoable replacement, tests and docs; or stop and report when blocked on an open design question or a denied permission. On a loop, an architectural blocker, or no progress, clear the goal and write a failure report.
