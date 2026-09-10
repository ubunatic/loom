// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package layout

import (
	"math"
	"testing"
)

func TestPlanRejectsGapOverflow(t *testing.T) {
	items := []Item{{Visible: true}, {Visible: true}, {Visible: true}}
	if _, err := Plan(0, math.MaxInt, items); err == nil {
		t.Fatal("expected gap overflow/insufficient-space error")
	}
}
