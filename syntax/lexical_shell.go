package syntax

import (
	"strings"
	"unicode/utf8"
)

// shellKeywords get the keyword capture wherever they appear as a whole word.
var shellKeywords = map[string]bool{
	"if": true, "then": true, "else": true, "elif": true, "fi": true,
	"for": true, "while": true, "until": true, "do": true, "done": true,
	"case": true, "esac": true, "in": true, "function": true, "select": true,
	"return": true, "local": true, "export": true,
}

// isShellLanguage reports whether lang names the shell lexer.
func isShellLanguage(lang string) bool {
	switch lang {
	case "bash", "sh", "shell", "zsh":
		return true
	}
	return false
}

// lexShellLine lexes one line of shell source; the line is src[base:end].
// It is line-based: a quote, heredoc or $( that is not closed on this line
// does not carry over to the next line.
func lexShellLine(src []byte, base, end int) []Span {
	b := src[base:end]
	out := []Span{}
	add := func(a, z int, c string) {
		z = shellBoundary(b, z)
		if z > a {
			out = append(out, Span{Line: lineAt(src, base), StartByte: base + a, EndByte: base + z, StartRune: ByteToRune(src, base+a), EndRune: ByteToRune(src, base+z), Capture: c})
		}
	}
	cmdPos := true // the next word is a command name
	for i := 0; i < len(b); {
		c := b[i]
		switch {
		case c == ' ' || c == '\t':
			i++
		case c == '#' && (i == 0 || b[i-1] == ' ' || b[i-1] == '\t'):
			add(i, len(b), "comment")
			return out
		case c == '\'':
			j := shellSingleEnd(b, i)
			add(i, j, "string")
			cmdPos = false
			i = j
		case c == '"':
			i = shellDoubleSpans(b, i, add)
			cmdPos = false
		case c == '$':
			if j := shellVarEnd(b, i); j > i {
				add(i, j, string(CaptureVariable))
				i = j
			} else {
				i++
			}
			cmdPos = false
		case strings.IndexByte("|&<>=", c) >= 0:
			j := i
			for j < len(b) && strings.IndexByte("|&<>=", b[j]) >= 0 {
				j++
			}
			op := string(b[i:j])
			add(i, j, "operator")
			// A pipe or a list operator starts a new command; a redirect does not.
			cmdPos = strings.ContainsAny(op, "|&") && !strings.ContainsAny(op, "<>")
			i = j
		case c == ';':
			add(i, i+1, "punctuation")
			cmdPos = true
			i++
		case c == '(':
			add(i, i+1, "punctuation.bracket")
			cmdPos = true
			i++
		case strings.IndexByte(")[]{}", c) >= 0:
			add(i, i+1, "punctuation.bracket")
			cmdPos = false
			i++
		default:
			j := shellWordEnd(b, i)
			word := string(b[i:j])
			switch {
			case shellKeywords[word]:
				add(i, j, "keyword")
				cmdPos = word == "then" || word == "do"
			case strings.Contains(word, "="):
				// An assignment such as FOO=1 keeps the command position open.
			case word[0] == '-':
				cmdPos = false
			case isDigit(word[0]) && shellNumber(word):
				add(i, j, "number")
				cmdPos = false
			case cmdPos:
				add(i, j, "function")
				cmdPos = false
			default:
				cmdPos = false
			}
			i = j
		}
	}
	return out
}

// shellWordStop reports whether c ends an unquoted shell word. A '=' inside a
// word, a '#' inside a word and a backslash escape do not end it.
func shellWordStop(c byte) bool {
	return strings.IndexByte(" \t'\"$|&;()<>{}[]", c) >= 0
}

// shellWordEnd returns the end of the word that starts at b[i]. A backslash
// escapes the next byte, so an escaped quote does not start a string.
func shellWordEnd(b []byte, i int) int {
	j := i
	for j < len(b) && !shellWordStop(b[j]) {
		if b[j] == '\\' {
			j = shellEscapeEnd(b, j)
			continue
		}
		j++
	}
	if j == i {
		j = i + 1
	}
	return j
}

// shellEscapeEnd returns the end of the backslash escape at b[j]: the backslash
// and the whole rune after it, so an escape never splits a multi-byte rune.
func shellEscapeEnd(b []byte, j int) int {
	j++
	if j < len(b) {
		_, n := utf8.DecodeRune(b[j:])
		j += n
	}
	return j
}

// shellBoundary moves the span end z back onto a UTF-8 boundary of b. A
// truncated multi-byte rune at the end of the line is left out of the span.
func shellBoundary(b []byte, z int) int {
	if z > len(b) {
		return len(b)
	}
	if z < len(b) {
		for z > 0 && !utf8.RuneStart(b[z]) {
			z--
		}
		return z
	}
	for k := len(b) - 1; k >= 0 && len(b)-k <= utf8.UTFMax; k-- {
		if utf8.RuneStart(b[k]) {
			if !utf8.FullRune(b[k:]) {
				return k
			}
			break
		}
	}
	return len(b)
}

// shellNumber reports whether word, which starts with a digit, is made only of
// digits, dots and underscores. Other words that start with a digit stay plain.
func shellNumber(word string) bool {
	for i := 0; i < len(word); i++ {
		if !isDigit(word[i]) && word[i] != '.' && word[i] != '_' {
			return false
		}
	}
	return true
}

// shellSingleEnd returns the end of the single-quoted string that starts at
// b[i]. Single quotes have no escapes. An unterminated string runs to the end
// of the line.
func shellSingleEnd(b []byte, i int) int {
	j := i + 1
	for j < len(b) && b[j] != '\'' {
		j++
	}
	if j < len(b) {
		j++
	}
	return j
}

// shellDoubleSpans emits the spans of the double-quoted string that starts at
// b[i]: string text with backslash escapes, and each variable inside it as a
// variable span. It returns the end of the string. An unterminated string runs
// to the end of the line.
func shellDoubleSpans(b []byte, i int, add func(a, z int, c string)) int {
	seg := i
	j := i + 1
	for j < len(b) {
		if b[j] == '\\' {
			j = shellEscapeEnd(b, j)
			continue
		}
		if b[j] == '"' {
			j++
			add(seg, j, "string")
			return j
		}
		if b[j] == '$' {
			// Every '$' starts a new segment, so a cut right after it does not
			// merge it into the string before it.
			add(seg, j, "string")
			seg = j
			if k := shellVarEnd(b, j); k > j {
				add(j, k, string(CaptureVariable))
				seg, j = k, k
				continue
			}
		}
		j++
	}
	if j > len(b) {
		j = len(b)
	}
	add(seg, j, "string")
	return j
}

// shellVarEnd returns the end of the parameter expansion that starts with the
// '$' at b[i], or i when the '$' does not start one (for example "$(" or a lone
// '$' at the end of the line). It handles $name, ${...} (nested braces are
// counted; an unterminated one runs to the end of the line), positional $1,
// and the special parameters $@ $* $# $? $$ $! $-.
func shellVarEnd(b []byte, i int) int {
	if i+1 >= len(b) {
		return i
	}
	c := b[i+1]
	switch {
	case c == '{':
		depth := 0
		for j := i + 1; j < len(b); j++ {
			switch b[j] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return j + 1
				}
			}
		}
		return len(b)
	case identStart(c):
		j := i + 1
		for j < len(b) && identContinue(b[j]) {
			j++
		}
		return j
	case isDigit(c):
		return i + 2
	case strings.IndexByte("@*#?$!-", c) >= 0:
		return i + 2
	}
	return i
}
