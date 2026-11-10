package measure

import (
	"strings"
	"testing"
)

func TestGraphemeBreakRules(t *testing.T) {
	// 10 hand-written test cases since network: unavailable for fetching GraphemeBreakTest.txt
	cases := []struct {
		desc string
		in   string
		want []string
	}{
		{"GB9c Devanagari Consonant+Linker+Consonant", "\u0915\u094D\u0937", []string{"\u0915\u094D\u0937"}},
		{"GB9c Devanagari Consonant+Linker+Extend+Consonant", "\u0915\u094D\u0308\u0937", []string{"\u0915\u094D\u0308\u0937"}},
		{"GB9c Bengali Consonant+Linker+Consonant", "\u0995\u09CD\u09B7", []string{"\u0995\u09CD\u09B7"}},
		{"GB9c Malayalam Consonant+Linker+Consonant", "\u0D15\u0D4D\u0D37", []string{"\u0D15\u0D4D\u0D37"}},
		{"Skin tone modifier base", "👍🏽", []string{"👍🏽"}},
		{"Skin tone modifier and ZWJ", "🤷🏽‍♂️", []string{"🤷🏽‍♂️"}},
		{"Isolated VS16", "\uFE0F", []string{"\uFE0F"}},
		{"Text with VS16", "a\uFE0F", []string{"a\uFE0F"}},
		{"Hash with VS16", "#\uFE0F", []string{"#\uFE0F"}},
		{"Heart with VS16", "❤\uFE0F", []string{"❤\uFE0F"}},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := ClustersNew(tc.in)
			if strings.Join(got, "|") != strings.Join(tc.want, "|") {
				t.Errorf("ClustersNew(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
