// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package gallery

import "ubunatic.com/loom"

func newHintBarDemo() loom.Widget {
	bar := loom.NewHintBar()
	bar.Entries = []loom.HintEntry{
		{Key: "F1", Binding: "f1", Label: "Help", Action: func() loom.EventResult {
			bar.Entries[0].Label = "Clicked"
			return loom.Handled()
		}},
		{Key: "F7", Binding: "f7", Label: "View", Action: func() loom.EventResult {
			if bar.Entries[1].Label == "View" {
				bar.Entries[1].Label = "Edit"
			} else {
				bar.Entries[1].Label = "View"
			}
			return loom.Handled()
		}},
	}
	return bar
}
