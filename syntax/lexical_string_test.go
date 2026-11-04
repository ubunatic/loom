package syntax

import (
	"reflect"
	"testing"
	"unicode/utf8"
)

func TestHighlightStringOffsetsAcrossLines(t *testing.T) {
	// Offsets are into s: "é" is two bytes and one rune, so the byte and rune
	// columns of the spans on line 1 differ by one.
	s := "{\"é\": 1}\n[null]"
	want := []Span{
		{Line: 0, StartByte: 0, EndByte: 1, StartRune: 0, EndRune: 1, Capture: "punctuation.bracket"},
		{Line: 0, StartByte: 1, EndByte: 5, StartRune: 1, EndRune: 4, Capture: "property"},
		{Line: 0, StartByte: 5, EndByte: 6, StartRune: 4, EndRune: 5, Capture: "punctuation"},
		{Line: 0, StartByte: 7, EndByte: 8, StartRune: 6, EndRune: 7, Capture: "number"},
		{Line: 0, StartByte: 8, EndByte: 9, StartRune: 7, EndRune: 8, Capture: "punctuation.bracket"},
		{Line: 1, StartByte: 10, EndByte: 11, StartRune: 9, EndRune: 10, Capture: "punctuation.bracket"},
		{Line: 1, StartByte: 11, EndByte: 15, StartRune: 10, EndRune: 14, Capture: "constant"},
		{Line: 1, StartByte: 15, EndByte: 16, StartRune: 14, EndRune: 15, Capture: "punctuation.bracket"},
	}
	got := HighlightString("json", s)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spans\n got %+v\nwant %+v", got, want)
	}
	for _, sp := range got {
		if ByteToRune([]byte(s), sp.StartByte) != sp.StartRune || ByteToRune([]byte(s), sp.EndByte) != sp.EndRune {
			t.Errorf("rune offsets disagree with bytes: %+v", sp)
		}
	}
}

func TestHighlightStringMatchesEngine(t *testing.T) {
	samples := map[string]string{
		"go":       "package main\nfunc run() { return 42 }; call() // note\n",
		"markdown": "# Heading\n`code` text\n",
		"json":     "{\n  \"a\": [1, -2.5e3, true],\n  \"b\": null\n}",
		"yaml":     "name: 'Ada' # note\nlist:\n  - 1\n",
	}
	for lang, src := range samples {
		t.Run(lang, func(t *testing.T) {
			want := engineSpans(t, lang, src)
			got := HighlightString(lang, src)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("HighlightString differs from engine\n got %+v\nwant %+v", got, want)
			}
		})
	}
}

func TestHighlightStringUnknownLanguageIsNil(t *testing.T) {
	for _, lang := range []string{"", "cobol", "text"} {
		if got := HighlightString(lang, `{"a": 1}`); got != nil {
			t.Errorf("HighlightString(%q)=%+v, want nil", lang, got)
		}
	}
	if got := HighlightString("JSON", `{"a": 1}`); len(got) == 0 {
		t.Errorf("language names are case-insensitive: got no spans")
	}
}

// TestHighlightStringTruncatedJSON cuts a sample at every byte. Each cut must
// not panic, every span must lie inside the cut text and be non-empty, and every
// token that ends before the cut must keep the capture it has in the full text.
func TestHighlightStringTruncatedJSON(t *testing.T) {
	sample := `{"id":-5,"e":1.5e+3,"ok":true,"s":"a\"b","n":null,"x":[-0.25E-2]}`
	full := HighlightString("json", sample)
	if len(full) == 0 {
		t.Fatal("full sample has no spans")
	}
	for cut := 0; cut <= len(sample); cut++ {
		src := sample[:cut]
		got := HighlightString("json", src)
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
