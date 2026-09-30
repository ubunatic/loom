// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package usage

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Setenv("LOOMCOLOR", "truecolor"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
