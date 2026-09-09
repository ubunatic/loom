// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "strings"

// TruncateText fits text to a terminal-cell budget, adding the caller's marker
// only when truncation is needed. It follows Canvas's supported text policy:
// terminal instructions are stripped, combining marks stay with their base,
// and wide glyphs are never split. Pass styles separately through Style.
// A marker wider than the budget is itself clipped, without recursive markers.
func TruncateText(text string, width int, marker string) string {
	if width <= 0 {
		return ""
	}
	clusters := textClusters(text)
	plain := strings.Join(clusters, "")
	if StringWidth(plain) <= width {
		return plain
	}
	end := fitClusters(textClusters(marker), width)
	return fitClusters(clusters, width-StringWidth(end)) + end
}

func fitClusters(clusters []string, width int) string {
	var result strings.Builder
	for _, cluster := range clusters {
		cells := StringWidth(cluster)
		if cells > width {
			break
		}
		result.WriteString(cluster)
		width -= cells
	}
	return result.String()
}
