// Package syntax defines UI-neutral syntax highlighting interfaces and types.
package syntax

// Point is a zero-based source position. Column is measured in bytes.
type Point struct{ Row, Column int }

// Edit describes a byte-range replacement and its source positions.
type Edit struct {
	StartByte, OldEndByte, NewEndByte    int
	StartPoint, OldEndPoint, NewEndPoint Point
}

// Span identifies a captured token. Line is zero-based; byte and rune offsets
// are zero-based offsets in the complete parsed document (not line-local).
// Start offsets are inclusive and end offsets are exclusive. Byte offsets
// always land on UTF-8 boundaries.
type Span struct {
	Line               int
	StartByte, EndByte int
	StartRune, EndRune int
	Capture            string
}

// Engine parses source and returns capture spans for lines or viewports.
type Engine interface {
	Language() string
	NotifyEdit(edit Edit)
	Parse(source []byte) error
	HighlightLine(line int) []Span
	HighlightViewport(startLine, endLine int) map[int][]Span
}

// Symbol identifies a recognized code entity (function, method, type, heading, key, etc.).
type Symbol struct {
	Name    string
	Kind    string
	Line    int // 0-based start line
	Column  int // 0-based start column
	EndLine int // 0-based end line
}

// OutlineProvider supplies document symbols, breadcrumbs, and fold ranges.
type OutlineProvider interface {
	Symbols() []Symbol
	Breadcrumb(line, col int) []string
	Folds() [][2]int
}

// Navigator provides structural navigation, breadcrumbs, and fold ranges.
type Navigator interface {
	OutlineProvider
	Symbols() []Symbol
	Breadcrumb(line, col int) []string
	Folds() [][2]int
}
