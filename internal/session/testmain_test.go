// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package session

import (
	"os"
	"testing"

	"ubunatic.com/loom/internal/clitest"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("LOOMCOLOR", "truecolor")
	os.Exit(clitest.RunTests(m))
}
