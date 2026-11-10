package measure

import (
	"strings"
	"testing"
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
		concat := strings.Join(clusters, "")
		if concat != s {
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

		if sumWidth != w {
			t.Fatalf("sum of cluster widths %d != StringWidth %d", sumWidth, w)
		}
	})
}
