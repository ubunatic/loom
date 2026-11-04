package syntax

import (
	"reflect"
	"testing"
)

// engineSpans lexes src with NewLexicalEngine and returns every line's spans in order.
func engineSpans(t *testing.T, lang, src string) []Span {
	t.Helper()
	e := NewLexicalEngine(lang)
	if err := e.Parse([]byte(src)); err != nil {
		t.Fatal(err)
	}
	var out []Span
	for line := range lineRanges([]byte(src)) {
		out = append(out, e.HighlightLine(line)...)
	}
	return out
}

// tokenText lists each span as "capture:text", reading text from src by byte offsets.
func tokenText(src string, spans []Span) []string {
	out := make([]string, 0, len(spans))
	for _, s := range spans {
		out = append(out, s.Capture+":"+src[s.StartByte:s.EndByte])
	}
	return out
}

func TestJSONKeysConstantsAndNumbers(t *testing.T) {
	cases := []struct {
		name, src string
		want      []string
	}{
		{
			name: "taia tool args",
			src:  `{"id":5,"offset":0,"limit":8000}`,
			want: []string{
				"punctuation.bracket:{", `property:"id"`, "punctuation::", "number:5",
				"punctuation:,", `property:"offset"`, "punctuation::", "number:0",
				"punctuation:,", `property:"limit"`, "punctuation::", "number:8000",
				"punctuation.bracket:}",
			},
		},
		{
			name: "signs exponents and constants",
			src:  `[-5, 1e3, 2.5E-4, -0.5e+10, true, false, null]`,
			want: []string{
				"punctuation.bracket:[", "number:-5", "punctuation:,", "number:1e3",
				"punctuation:,", "number:2.5E-4", "punctuation:,", "number:-0.5e+10",
				"punctuation:,", "constant:true", "punctuation:,", "constant:false",
				"punctuation:,", "constant:null", "punctuation.bracket:]",
			},
		},
		{
			name: "nested objects and arrays",
			src:  `{"a":{"b":[true]}}`,
			want: []string{
				"punctuation.bracket:{", `property:"a"`, "punctuation::", "punctuation.bracket:{",
				`property:"b"`, "punctuation::", "punctuation.bracket:[", "constant:true",
				"punctuation.bracket:]", "punctuation.bracket:}", "punctuation.bracket:}",
			},
		},
		{
			name: "escaped quotes in key and value",
			src:  `{"a\"b": "c\\"}`,
			want: []string{
				"punctuation.bracket:{", `property:"a\"b"`, "punctuation::", `string:"c\\"`,
				"punctuation.bracket:}",
			},
		},
		{
			name: "spaces between key and colon",
			src:  `{"k"  :  "v"}`,
			want: []string{
				"punctuation.bracket:{", `property:"k"`, "punctuation::", `string:"v"`,
				"punctuation.bracket:}",
			},
		},
		{
			name: "string values are not keys",
			src:  `["x", "y"]`,
			want: []string{
				"punctuation.bracket:[", `string:"x"`, "punctuation:,", `string:"y"`,
				"punctuation.bracket:]",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tokenText(tc.src, engineSpans(t, "json", tc.src))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("tokens\n got %q\nwant %q", got, tc.want)
			}
		})
	}
}
