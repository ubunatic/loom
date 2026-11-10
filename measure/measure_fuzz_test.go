package measure

import (
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

func FuzzMeasure(f *testing.F) {
	for _, seed := range []string{
		"a",
		"👨‍👩‍👧",
		"❤️‍🔥",
		"中",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		w := StringWidth(s)
		if w < 0 {
			t.Fatalf("negative width: %d", w)
		}

		clusters := Clusters(s)
		// Clusters returns printable clusters only: it drops control and
		// non-format Cf runes by design, so concatenation holds only for
		// valid text without them.
		// A leading mark has no base to attach to and is dropped too.
		first, _ := utf8.DecodeRuneInString(s)
		dropped := !utf8.ValidString(s) || strings.IndexFunc(s, func(r rune) bool {
			return unicode.IsControl(r) || (unicode.Is(unicode.Cf, r) && !isFormatRune(r))
		}) >= 0 || unicode.In(first, unicode.Mn, unicode.Me) || isEmojiModifier(first)
		concat := strings.Join(clusters, "")
		if !dropped && concat != s {
			t.Fatalf("clusters do not concat back: %q vs %q", concat, s)
		}

		sumWidth := 0
		for _, c := range clusters {
			cw := StringWidth(c)
			if cw < 0 {
				t.Fatalf("negative cluster width: %d", cw)
			}
			sumWidth += cw
		}

		// Dropped runes can split or join clusters differently, so the
		// sum holds only for the same text as the concat check.
		if !dropped && sumWidth != w {
			t.Fatalf("sum of cluster widths %d != StringWidth %d", sumWidth, w)
		}
	})
}
