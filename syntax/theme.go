package syntax

// Capture names a standard syntax token class.
type Capture string

const (
	CaptureKeyword            Capture = "keyword"
	CaptureFunction           Capture = "function"
	CaptureFunctionCall       Capture = "function.call"
	CaptureType               Capture = "type"
	CaptureString             Capture = "string"
	CaptureComment            Capture = "comment"
	CaptureNumber             Capture = "number"
	CaptureOperator           Capture = "operator"
	CapturePunctuation        Capture = "punctuation"
	CapturePunctuationBracket Capture = "punctuation.bracket"
)

// ThemeMap maps capture names to UI-owned theme identifiers. Values are
// intentionally strings so this package does not depend on Loom's UI types.
type ThemeMap map[string]string

// Resolve returns the theme identifier for capture, or the empty string.
func (m ThemeMap) Resolve(capture string) string { return m[capture] }

// DefaultThemeMap provides conventional theme keys for standard captures.
func DefaultThemeMap() ThemeMap {
	return ThemeMap{
		"keyword": "syntax.keyword", "function": "syntax.function",
		"function.call": "syntax.function", "type": "syntax.type",
		"string": "syntax.string", "comment": "syntax.comment",
		"number": "syntax.number", "operator": "syntax.operator",
		"punctuation": "syntax.punctuation", "punctuation.bracket": "syntax.punctuation",
	}
}

// StyleResolver maps standard captures to ANSI SGR styles. Values are SGR
// parameter strings, suitable for wrapping text in ESC[<value>m ... ESC[0m.
// Both 16-color and 256-color terminals understand these standard sequences.
type StyleResolver map[string]string

// DefaultStyleResolver returns conventional colors for standard captures.
func DefaultStyleResolver() StyleResolver {
	return StyleResolver{
		"keyword": "1;35", "function": "1;36", "function.call": "36",
		"type": "1;34", "string": "32", "comment": "90",
		"number": "33", "operator": "37", "punctuation": "37",
		"punctuation.bracket": "37",
	}
}

// Resolve returns an SGR style, or empty when the capture is unknown.
func (r StyleResolver) Resolve(capture string) string { return r[capture] }
