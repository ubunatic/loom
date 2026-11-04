// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package editor

import (
	"os"
	"path/filepath"

	"ubunatic.com/loom"
	"ubunatic.com/loom/media"
)

// Options controls how the editor starts.
type Options struct {
	// ShowFiles starts with the file pane shown and focused.
	ShowFiles bool
	// Dir roots the file pane. Empty means the opened file's directory, or the working directory.
	Dir string
}

// resolveTarget turns a command argument into the file to open and the start
// options. A directory opens no file, roots the file pane there and shows it.
// Any other path, including a missing one, is returned unchanged for loading.
func resolveTarget(path string, opts Options) (string, Options, error) {
	if path == "" {
		return path, opts, nil
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return path, opts, nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", opts, err
	}
	opts.ShowFiles = true
	opts.Dir = abs
	return "", opts, nil
}

// Run opens path in the interactive editor using cfg and opts. See resolveTarget
// for how a directory path is handled.
func Run(path string, cfg loom.EditorConfig, opts Options) error {
	path, opts, err := resolveTarget(path, opts)
	if err != nil {
		return err
	}
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
	view, err := newEditViewWith(edit, path, cfg, opts)
	if err != nil {
		return err
	}
	view.attach(pane)
	return pane.Run(view)
}
