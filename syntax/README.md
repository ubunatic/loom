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
standard-library fallback for Go, Markdown, JSON, and YAML.

`DefaultStyleResolver` maps standard capture names directly to ANSI SGR
parameters for terminal adapters; `DefaultThemeMap` remains available for UI
adapters that use named theme keys.
