package syntax

import (
	"reflect"
	"testing"
	"unicode/utf8"
)

// TestShellTokens checks the token list of each sample, lexed as "sh".
// Columns are byte offsets into the source; each token is "capture:text".
func TestShellTokens(t *testing.T) {
	cases := []struct {
		name, src string
		want      []string
	}{
		{
			name: "shebang comment and pipes",
			src:  "#!/bin/sh\nls -l | grep \"x\"",
			want: []string{
				"comment:#!/bin/sh",
				"function:ls", "operator:|", "function:grep", `string:"x"`,
			},
		},
		{
			name: "command substitution in an assignment and a trailing comment",
			src:  "x=$(date) # now",
			want: []string{
				"punctuation.bracket:(", "function:date", "punctuation.bracket:)", "comment:# now",
			},
		},
		{
			name: "hash inside words is not a comment",
			src:  "echo a#b foo#",
			want: []string{"function:echo"},
		},
		{
			name: "indented comment after spaces",
			src:  "  # note",
			want: []string{"comment:# note"},
		},
		{
			name: "parameter length and expansions",
			src:  "echo ${#x} $# ${var:-x}",
			want: []string{
				"function:echo", "variable:${#x}", "variable:$#", "variable:${var:-x}",
			},
		},
		{
			name: "special and positional parameters",
			src:  "echo $1 $@ $? $name",
			want: []string{
				"function:echo", "variable:$1", "variable:$@", "variable:$?", "variable:$name",
			},
		},
		{
			// "if" is not a command-position trigger (only then and do are),
			// so "true" stays plain.
			name: "keywords and a command after then",
			src:  "if true; then echo hi; fi",
			want: []string{
				"keyword:if", "punctuation:;", "keyword:then", "function:echo",
				"punctuation:;", "keyword:fi",
			},
		},
		{
			name: "export is a keyword and its assignment stays plain",
			src:  "export PATH=/x",
			want: []string{"keyword:export"},
		},
		{
			name: "numbers are captured and flags stay plain",
			src:  "head -n 10 file2",
			want: []string{"function:head", "number:10"},
		},
		{
			name: "flags stay plain",
			src:  "ls --all -la",
			want: []string{"function:ls"},
		},
		{
			name: "single quotes do not expand variables",
			src:  "echo 'a $x'",
			want: []string{"function:echo", "string:'a $x'"},
		},
		{
			name: "double quotes keep escapes and split out variables",
			src:  `echo "a\"b $HOME"`,
			want: []string{
				"function:echo", `string:"a\"b `, "variable:$HOME", `string:"`,
			},
		},
		{
			name: "unterminated strings run to the end of the line",
			src:  "echo \"abc\necho 'x",
			want: []string{
				"function:echo", `string:"abc`,
				"function:echo", "string:'x",
			},
		},
		{
			name: "operators start a new command",
			src:  "make && test -f x || echo no",
			want: []string{
				"function:make", "operator:&&", "function:test", "operator:||", "function:echo",
			},
		},
		{
			name: "subshell group",
			src:  "(cd /tmp; ls)",
			want: []string{
				"punctuation.bracket:(", "function:cd", "punctuation:;", "function:ls",
				"punctuation.bracket:)",
			},
		},
		{
			name: "loop with a quoted variable",
			src:  `for f in *.txt; do rm "$f"; done`,
			want: []string{
				"keyword:for", "keyword:in", "punctuation:;", "keyword:do",
				"function:rm", `string:"`, "variable:$f", `string:"`,
				"punctuation:;", "keyword:done",
			},
		},
		{
			name: "heredoc start line",
			src:  "cat <<EOF",
			want: []string{"function:cat", "operator:<<"},
		},
		{
			name: "assignment before a command",
			src:  "FOO=1 make",
			want: []string{"function:make"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tokenText(tc.src, engineSpans(t, "sh", tc.src))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("tokens\n got %q\nwant %q", got, tc.want)
			}
		})
	}
}

// TestShellLanguageNames checks that every shell name gives the same spans, and
// that HighlightString uses the same lexer as the engine.
func TestShellLanguageNames(t *testing.T) {
	src := "#!/bin/bash\nfor f in \"$DIR\"/*.txt; do echo \"${f#x}\" | wc -l; done # done\n"
	want := engineSpans(t, "bash", src)
	if len(want) == 0 {
		t.Fatal("bash has no spans")
	}
	for _, lang := range []string{"bash", "sh", "shell", "zsh", "SH"} {
		t.Run(lang, func(t *testing.T) {
			got := HighlightString(lang, src)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("HighlightString(%q) differs from engine\n got %+v\nwant %+v", lang, got, want)
			}
			if e := NewLexicalEngine(lang); e.Language() == "" {
				t.Errorf("engine for %q has no language name", lang)
			}
		})
	}
}

// TestShellTruncatedNeverPanics cuts a sample at every byte. Each cut must not
// panic, spans must stay inside the cut text, rune offsets must agree with byte
// offsets, and every token that ends before the cut keeps its capture.
func TestShellTruncatedNeverPanics(t *testing.T) {
	sample := "#!/bin/sh\nx=$(date) # é note\nfor f in \"$DIR\"/*.txt; do echo \"${f#x}\" | wc -l; done\ncat <<EOF \\é\nFOO=1 make && echo '$1' $@ ${var:-y} $# \"a\\\"b\" \"é$\"\n"
	full := HighlightString("sh", sample)
	if len(full) == 0 {
		t.Fatal("full sample has no spans")
	}
	for cut := 0; cut <= len(sample); cut++ {
		src := sample[:cut]
		got := HighlightString("sh", src)
		for _, sp := range got {
			if sp.StartByte < 0 || sp.StartByte >= sp.EndByte || sp.EndByte > cut {
				t.Fatalf("cut %d: span out of bounds: %+v", cut, sp)
			}
			if sp.StartRune != utf8.RuneCountInString(src[:sp.StartByte]) || sp.EndRune != utf8.RuneCountInString(src[:sp.EndByte]) {
				t.Fatalf("cut %d: rune offsets wrong: %+v", cut, sp)
			}
		}
		for _, want := range full {
			if want.EndByte >= cut {
				continue
			}
			found := false
			for _, sp := range got {
				if sp.StartByte == want.StartByte && sp.EndByte == want.EndByte && sp.Capture == want.Capture {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("cut %d: complete token %q (%s) lost; got %q", cut, sample[want.StartByte:want.EndByte], want.Capture, tokenText(src, got))
			}
		}
	}
}

// TestShellVariableCaptureHasTheme checks that the variable capture has theme
// and SGR entries, so terminal and UI adapters can colour it.
func TestShellVariableCaptureHasTheme(t *testing.T) {
	if got := DefaultThemeMap().Resolve("variable"); got != "syntax.variable" {
		t.Errorf("theme for variable = %q, want syntax.variable", got)
	}
	if got := DefaultStyleResolver().Resolve("variable"); got == "" {
		t.Errorf("SGR style for variable is empty")
	}
}
