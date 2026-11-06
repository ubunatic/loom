# Issue 441 M1: TextEntry bold state and Markdown serialization

## Scope

M1 adds opt-in Ctrl+B bold state to `TextEntry`, tracks formatting with grapheme units through edits and history, draws it with `Style.Bold`, and exposes `MarkdownValue()`. The gallery and actual submission callback remain in M2.

## Implementation

`EnableBold()` opts an entry into Ctrl+B, preserving host handling for existing entries. A selected range toggles all selected non-newline graphemes together; with no selection, Ctrl+B changes the typing style. Bold flags and typing style are snapshotted with undo state. Text callbacks and clipboard contents remain plain. `MarkdownValue()` escapes ASCII punctuation and backslashes, writes strong runs per hard line, preserves newlines and leaves edge whitespace outside delimiters.

## Evidence

Regression coverage exercises opt-in routing, mixed selection, typing style, format-only undo/redo and callbacks, grapheme-aligned replacement and paste, multiline typing, Ctrl+U history, placeholder acceptance, plain clipboard text, and style composition with highlight and selection.

Focused checks passed: `go test . -run 'TestTextEntry(Bold|Markdown)' -count=1` and `go test . -run '^TestTextEntry' -count=1`. `go vet ./...` passed and the public API snapshot was refreshed. The single final `make test-q1` passed with no `--- FAIL` lines; `make install` passed. Implementation commit: `22ec2fc`.

## Limits

Whitespace-only emphasis has no Markdown bold semantics. This API serializes current text and formatting; it does not parse Markdown or provide a round trip.
