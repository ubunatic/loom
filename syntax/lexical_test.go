package syntax

import "testing"

func TestLexicalEngineLanguages(t *testing.T) {
	cases := []struct {
		lang, src string
		captures  []string
	}{
		{"go", "package main\nfunc run() { return 42 }; call() // note", []string{"keyword", "function", "function.call", "punctuation.bracket", "number", "comment"}},
		{"markdown", "# Heading\n`code`", []string{"keyword", "string"}},
		{"json", `{"ok": true, "n": 12}`, []string{"punctuation.bracket", "string", "keyword", "number"}},
		{"yaml", "name: 'Ada' # note", []string{"type", "string", "comment"}},
	}
	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			e := NewLexicalEngine(tc.lang)
			if err := e.Parse([]byte(tc.src)); err != nil {
				t.Fatal(err)
			}
			got := e.HighlightViewport(0, 20)
			all := map[string]bool{}
			for _, spans := range got {
				for _, s := range spans {
					if s.StartByte < 0 || s.EndByte <= s.StartByte || s.StartRune < 0 || s.EndRune <= s.StartRune {
						t.Errorf("invalid span: %+v", s)
					}
					all[s.Capture] = true
				}
			}
			for _, capture := range tc.captures {
				if !all[capture] {
					t.Errorf("missing %s in %#v", capture, got)
				}
			}
		})
	}
}

func TestLexicalLineBoundsAndEdit(t *testing.T) {
	e := NewLexicalEngine("go")
	src := []byte("var x = 1\nreturn")
	if err := e.Parse(src); err != nil {
		t.Fatal(err)
	}
	e.NotifyEdit(Edit{})
	if got := e.HighlightLine(-1); len(got) != 0 {
		t.Fatalf("negative line: %v", got)
	}
	if got := e.HighlightViewport(1, 2); len(got) != 1 || got[1][0].Capture != "keyword" {
		t.Fatalf("viewport: %#v", got)
	}
}

func TestParseOwnsDocumentAndSpansUseDocumentCoordinates(t *testing.T) {
	e := NewLexicalEngine("go")
	if err := e.Parse([]byte("func Café() {}\r\ncall()")); err != nil {
		t.Fatal(err)
	}
	spans := e.HighlightLine(0)
	if len(spans) < 3 {
		t.Fatalf("spans: %#v", spans)
	}
	if got, want := spans[1], (Span{Line: 0, StartByte: 5, EndByte: 10, StartRune: 5, EndRune: 9, Capture: "function"}); got != want {
		t.Errorf("declaration span=%+v want %+v", got, want)
	}
	line1 := e.HighlightLine(1)
	if len(line1) != 3 {
		t.Fatalf("call spans: %#v", line1)
	}
	if got, want := line1[0], (Span{Line: 1, StartByte: 17, EndByte: 21, StartRune: 16, EndRune: 20, Capture: "function.call"}); got != want {
		t.Errorf("call span=%+v want %+v", got, want)
	}
	if err := e.Parse([]byte("return 1")); err != nil {
		t.Fatal(err)
	}
	if got := e.HighlightLine(1); len(got) != 0 {
		t.Fatalf("old document retained: %#v", got)
	}
	if got := e.HighlightLine(0); len(got) != 2 {
		t.Fatalf("new parsed document not used: %#v", got)
	}
}
