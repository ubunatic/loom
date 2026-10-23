// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package editor

import (
	"testing"

	"ubunatic.com/loom"
)

func TestEditCommandPaneSettings(t *testing.T) {
	pane := &loom.Pane{MaxCols: loom.DefaultMaxCols}
	cfg := loom.EditorConfig{MouseGrab: true, AltScreen: true}
	configureEditPane(pane, cfg)
	if pane.MaxCols != 0 {
		t.Errorf("edit pane MaxCols = %d, want 0", pane.MaxCols)
	}
	if !pane.Resizeable {
		t.Error("edit pane Resizeable = false, want true")
	}
	if pane.DisableGlobalF10Quit {
		t.Error("edit pane DisableGlobalF10Quit = true, want false: the Pane owns F10")
	}
	if !pane.ResizeConfig.AltScreen {
		t.Errorf("pane AltScreen = false, want true")
	}
}
