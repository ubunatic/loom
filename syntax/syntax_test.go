package syntax

import "testing"

func TestThemeMapResolve(t *testing.T) {
	m := DefaultThemeMap()
	for capture, want := range map[string]string{"keyword": "syntax.keyword", "function.call": "syntax.function", "type": "syntax.type", "string": "syntax.string", "comment": "syntax.comment", "number": "syntax.number", "operator": "syntax.operator", "punctuation.bracket": "syntax.punctuation"} {
		if got := m.Resolve(capture); got != want {
			t.Errorf("Resolve(%q)=%q, want %q", capture, got, want)
		}
	}
	if got := (ThemeMap{}).Resolve("missing"); got != "" {
		t.Fatalf("unknown capture=%q", got)
	}
}

func TestDefaultStyleResolverCoversStandardCaptures(t *testing.T) {
	r := DefaultStyleResolver()
	want := map[string]string{
		"keyword": "1;35", "function": "1;36", "function.call": "36",
		"type": "1;34", "string": "32", "comment": "90", "number": "33",
		"operator": "37", "punctuation": "37", "punctuation.bracket": "37",
	}
	for capture, style := range want {
		if got := r.Resolve(capture); got != style {
			t.Errorf("Resolve(%q)=%q want %q", capture, got, style)
		}
	}
	if got := r.Resolve("unknown"); got != "" {
		t.Errorf("unknown capture style=%q", got)
	}
}

func TestNullEngine(t *testing.T) {
	var e Engine = NullEngine{}
	if e.Language() != "" {
		t.Fatal("null language should be empty")
	}
	if err := e.Parse([]byte("x")); err != nil {
		t.Fatal(err)
	}
	e.NotifyEdit(Edit{})
	if got := e.HighlightLine(0); len(got) != 0 {
		t.Fatalf("spans=%v", got)
	}
	if got := e.HighlightViewport(0, 1); len(got) != 0 {
		t.Fatalf("viewport=%v", got)
	}
}

func TestSpanCoordinates(t *testing.T) {
	source := []byte("aé")
	s := Span{Line: 0, StartByte: 1, EndByte: 3, StartRune: 1, EndRune: 2, Capture: "string"}
	if ByteToRune(source, s.StartByte) != s.StartRune || ByteToRune(source, s.EndByte) != s.EndRune {
		t.Fatalf("span rune positions inconsistent: %+v", s)
	}
}
