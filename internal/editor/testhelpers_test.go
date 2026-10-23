// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package editor

import (
	"os"
	"testing"

	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/clitest"
)

func buildLoomBinary(t *testing.T) string { return clitest.BuildLoomBinary(t) }

func canvasPlainRow(canvas *loom.Canvas, row int) string {
	return clitest.CanvasPlainRow(canvas, row)
}

func canvasScreenText(canvas *loom.Canvas) string { return clitest.CanvasScreenText(canvas) }

func TestMain(m *testing.M) {
	_ = os.Setenv("LOOMCOLOR", "truecolor")
	os.Exit(m.Run())
}
