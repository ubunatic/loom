# Syntax providers

The `syntax` package defines source coordinates, capture spans, theme capture
names, and the `Engine` interface without importing Loom's UI packages. A UI
adapter can resolve each span's capture through `ThemeMap` and translate the
resulting theme key into its own style type.

Implementations receive edits through `NotifyEdit`, parse complete source with
`Parse`, and return line or viewport spans from the parsed document state.
Calling `Parse` replaces the document. `Span.Line` is zero-based; byte and rune
coordinates are zero-based document offsets with inclusive starts and exclusive
ends (byte offsets must be UTF-8 boundaries).
`NullEngine` disables highlighting. `NewLexicalEngine` provides a lightweight
standard-library fallback for Go, Markdown, JSON, YAML, and shell (`bash`, `sh`,
`shell`, `zsh`, all using one lexer).

`HighlightString(lang, s)` lexes a string without an editor buffer and returns
its spans; it returns nil for an unknown language. In JSON, object keys get the
`property` capture, `true`, `false` and `null` get `constant`, and numbers
include a sign, fraction and exponent. Truncated input is lexed without
panicking, and complete tokens keep their captures.

In shell, a `#` that starts a word or follows a space is a comment, keywords get
`keyword`, the first word of a command (line start, or after `|`, `&&`, `||`,
`;`, `(`, `then` or `do`) gets `function`, and `$name`, `${...}`, `$1`, `$@`,
`$?` and `$#` get `variable`, also inside double quotes. Single-quoted strings
do not expand, and an unterminated string runs to the end of its line.

`DefaultStyleResolver` maps standard capture names directly to ANSI SGR
parameters for terminal adapters; `DefaultThemeMap` remains available for UI
adapters that use named theme keys.
