// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package editor

import (
	"ubunatic.com/loom"
	"ubunatic.com/loom/media"
)

// Run opens path in the interactive editor using cfg.
func Run(path string, cfg loom.EditorConfig) error {
	edit := loom.NewRichTextEdit(nil)
	if path != "" {
		loaded, err := loom.NewRichTextEditFromFileWith(path, media.MarkupOptions(path))
		if err != nil {
			return err
		}
		edit = loaded
	}
	edit.ShowFileBar = loom.SpeccedDefaults.Editor.ShowFileBar

	pane, err := loom.New(loom.SpeccedDefaults.Editor.InitialHeight)
	if err != nil {
		return err
	}
	defer pane.Close()
	configureEditPane(pane, cfg)
	view, err := newEditView(edit, path, cfg)
	if err != nil {
		return err
	}
	view.attach(pane)
	return pane.Run(view)
}
