package measure

import (
	"testing"
)

func TestClustersWidthParity(t *testing.T) {
	corpus := []string{
		"こんにちは", // CJK
		"👨‍👩‍👧‍👦", // Emoji ZWJ sequence
		"🏳️‍🌈", // Emoji ZWJ sequence with VS16
		"❤️", // VS16
		"\uFE0F", // Isolated VS16 (disagreed previously)
		"Text\uFE0E", // VS15
		"🤷🏽‍♂️", // Emoji with skin tone modifier and ZWJ
	}

	for _, text := range corpus {
		clusters := ClustersNew(text)
		sum := 0
		for _, c := range clusters {
			sum += ClusterWidth(c)
		}

		sw := StringWidthNew(text)

		if sum != sw {
			t.Errorf("Clusters width mismatch for %q: sum=%d, StringWidthNew=%d", text, sum, sw)
		}
	}
}
