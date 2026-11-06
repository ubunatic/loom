# Issue 441: TextEntry visual bold and Markdown submission

## Scope

M1 added opt-in Ctrl+B bold state to `TextEntry`, tracks formatting with grapheme units through edits and history, draws it with `Style.Bold`, and exposes `MarkdownValue()`. M2 adds actual gallery submission and terminal proof while keeping the widget's text callbacks plain.

## Implementation

`EnableBold()` opts an entry into Ctrl+B, preserving host handling for existing entries. A selected range toggles all selected non-newline graphemes together; with no selection, Ctrl+B changes the typing style. Bold flags and typing style are snapshotted with undo state. Text callbacks and clipboard contents remain plain. `MarkdownValue()` escapes ASCII punctuation and backslashes, writes strong runs per hard line, preserves newlines and leaves edge whitespace outside delimiters.

## Evidence

M1 regression coverage exercises opt-in routing, mixed selection, typing style, format-only undo/redo and callbacks, grapheme-aligned replacement and paste, multiline typing, Ctrl+U history, placeholder acceptance, plain clipboard text, and style composition with highlight and selection. M2 adds compatibility tests for legacy selection colors, exact mutation and serialization expectations, and Ctrl+D immutability across enabled and disabled states. The gallery now submits `MarkdownValue()` in its demonstration callback and displays the serialized result while preserving plain `Value()` and callback data. A PTY test checks selected and typing bold cell attributes, caret position, format undo/redo, retained controls, placeholder behavior, and visible Markdown submission.

M1 focused checks, vet, API snapshot, single `make test-q1` and install passed; implementation commit: `22ec2fc`. M2 focused core, gallery and PTY checks passed, as did `go vet ./...`, the public API snapshot check, and `make install`. The single M2 `make test-q1` run failed only in unrelated `internal/session/TestSessionDetachPrefixDAndReattach`: its PTY output contained `BASE_SCREEN` and `DETACHED_OUTPUT` instead of the retained screen and tabs. No q1 rerun was made. M2 implementation commit: `a04eb55`; M2 scope/evidence commit: `5f78764`.

## Limits

Whitespace-only emphasis has no Markdown bold semantics. This API serializes current text and formatting; it does not parse Markdown or provide a round trip.
