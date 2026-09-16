// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package treemap

import (
	"fmt"

	"codeberg.org/ubunatic/loom/graph"
)

// Theme flag mapping and the demo palette are presentation choices. The
// renderer itself accepts any TreemapOptions supplied by a library caller.
func treemapTheme(theme int) graph.TreemapTheme {
	if theme == 2 {
		return graph.TreemapThemeNumbered
	}
	return graph.TreemapThemeClassic
}

func parseTheme(theme int, ansi bool) error {
	if theme != 1 && theme != 2 {
		return fmt.Errorf("treemap: --theme must be 1 or 2, got %d", theme)
	}
	if theme == 2 && !ansi {
		return fmt.Errorf("treemap: --theme 2 requires --ansi (it has no color to render its edges with)")
	}
	return nil
}

func demoPalette() (backgrounds, foregrounds []string) {
	return []string{"41", "42", "43", "44", "45", "46", "100", "47"},
		[]string{"97", "30", "30", "97", "97", "30", "97", "30"}
}
