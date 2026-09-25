package syntax

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// LexicalEngine provides lightweight standard-library tokenization.
type LexicalEngine struct {
	lang   string
	source []byte
}

// NewLexicalEngine creates a lexer for go, markdown, json, or yaml.
func NewLexicalEngine(language string) *LexicalEngine {
	return &LexicalEngine{lang: strings.ToLower(language)}
}
func (e *LexicalEngine) Language() string { return e.lang }
func (e *LexicalEngine) NotifyEdit(Edit)  {}
func (e *LexicalEngine) Parse(source []byte) error {
	e.source = append(e.source[:0], source...)
	return nil
}
func (e *LexicalEngine) HighlightLine(line int) []Span {
	source := e.source
	lines := lineRanges(source)
	if line < 0 || line >= len(lines) {
		return nil
	}
	start, end := lines[line][0], lines[line][1]
	return e.lexLine(source, start, end)
}
func (e *LexicalEngine) HighlightViewport(startLine, endLine int) map[int][]Span {
	out := map[int][]Span{}
	source := e.source
	lines := lineRanges(source)
	if startLine < 0 {
		startLine = 0
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if endLine < startLine {
		return out
	}
	for i := startLine; i < endLine; i++ {
		out[i] = e.HighlightLine(i)
	}
	return out
}
func lineRanges(b []byte) [][2]int {
	out := make([][2]int, 0)
	start := 0
	for i, c := range b {
		if c == '\n' {
			end := i
			if end > start && b[end-1] == '\r' {
				end--
			}
			out = append(out, [2]int{start, end})
			start = i + 1
		}
	}
	if start < len(b) || len(b) == 0 || b[len(b)-1] != '\n' {
		out = append(out, [2]int{start, len(b)})
	}
	return out
}
func lineAt(source []byte, offset int) int {
	line := 0
	for _, b := range source[:offset] {
		if b == '\n' {
			line++
		}
	}
	return line
}
func previousWord(b []byte, at int) string {
	for at > 0 && (b[at-1] == ' ' || b[at-1] == '\t') {
		at--
	}
	end := at
	for at > 0 && identContinue(b[at-1]) {
		at--
	}
	return string(b[at:end])
}
func (e *LexicalEngine) lexLine(src []byte, base, end int) []Span {
	b := src[base:end]
	out := []Span{}
	add := func(a, z int, c string) {
		if z > a {
			out = append(out, Span{Line: lineAt(src, base), StartByte: base + a, EndByte: base + z, StartRune: ByteToRune(src, base+a), EndRune: ByteToRune(src, base+z), Capture: c})
		}
	}
	lang := e.lang
	if lang == "markdown" || lang == "md" {
		if strings.HasPrefix(string(b), "#") {
			i := 0
			for i < len(b) && b[i] == '#' {
				i++
			}
			if i < len(b) && b[i] == ' ' {
				add(0, len(b), "keyword")
				return out
			}
		}
		for i := 0; i < len(b); {
			if b[i] == '`' {
				j := i + 1
				for j < len(b) && b[j] != '`' {
					j++
				}
				if j < len(b) {
					add(i, j+1, "string")
					i = j + 1
					continue
				}
			}
			i++
		}
		return out
	}
	for i := 0; i < len(b); {
		if (lang == "go" && i+1 < len(b) && b[i] == '/' && b[i+1] == '/') || ((lang == "yaml" || lang == "yml") && b[i] == '#') {
			add(i, len(b), "comment")
			break
		}
		if lang == "json" && b[i] == '/' && i+1 < len(b) && b[i+1] == '/' {
			add(i, len(b), "comment")
			break
		}
		if b[i] == '"' || (lang == "go" && b[i] == '`') || (lang == "yaml" && (b[i] == '\'' || b[i] == '"')) {
			q := b[i]
			j := i + 1
			for j < len(b) {
				if b[j] == '\\' && q == '"' {
					j += 2
					continue
				}
				if b[j] == q {
					j++
					break
				}
				j++
			}
			if j > len(b) {
				j = len(b)
			}
			add(i, j, "string")
			i = j
			continue
		}
		if isDigit(b[i]) {
			j := i + 1
			for j < len(b) && (isDigit(b[j]) || b[j] == '.' || b[j] == '_' || b[j] == 'x' || b[j] >= 'a' && b[j] <= 'f') {
				j++
			}
			add(i, j, "number")
			i = j
			continue
		}
		if isIdentifierStart(b[i:]) {
			j := i
			for j < len(b) {
				_, size := utf8.DecodeRune(b[j:])
				if !isIdentifierContinue(b[j : j+size]) {
					break
				}
				j += size
			}
			word := string(b[i:j])
			cap := ""
			switch lang {
			case "go":
				if goKeywords[word] {
					cap = "keyword"
				} else if j < len(b) && b[j] == '(' && previousWord(b, i) == "func" {
					cap = "function"
				} else if unicode.IsUpper(firstRune(word)) {
					cap = "type"
				} else if j < len(b) && b[j] == '(' {
					cap = "function.call"
				}
			case "json":
				if word == "true" || word == "false" || word == "null" {
					cap = "keyword"
				}
			case "yaml", "yml":
				if j < len(b) && b[j] == ':' {
					cap = "type"
				}
			}
			if cap != "" {
				add(i, j, cap)
			}
			i = j
			continue
		}
		if strings.ContainsRune("{}[]()", rune(b[i])) {
			add(i, i+1, "punctuation.bracket")
		} else if strings.ContainsRune(".,:;", rune(b[i])) {
			add(i, i+1, "punctuation")
		} else if strings.ContainsRune("+-*/%=!<>|&^~", rune(b[i])) {
			add(i, i+1, "operator")
		}
		_, n := utf8.DecodeRune(b[i:])
		i += n
	}
	return out
}
func firstRune(s string) rune { r, _ := utf8.DecodeRuneInString(s); return r }
func isIdentifierStart(b []byte) bool {
	r, _ := utf8.DecodeRune(b)
	return r == '_' || unicode.IsLetter(r)
}
func isIdentifierContinue(b []byte) bool {
	r, _ := utf8.DecodeRune(b)
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
func isDigit(b byte) bool       { return b >= '0' && b <= '9' }
func identStart(b byte) bool    { return b == '_' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' }
func identContinue(b byte) bool { return identStart(b) || isDigit(b) }

var goKeywords = map[string]bool{"break": true, "case": true, "chan": true, "const": true, "continue": true, "default": true, "defer": true, "else": true, "fallthrough": true, "for": true, "func": true, "go": true, "goto": true, "if": true, "import": true, "interface": true, "map": true, "package": true, "range": true, "return": true, "select": true, "struct": true, "switch": true, "type": true, "var": true}
