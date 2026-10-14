// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build windows

package measure

import (
	"fmt"
	"time"
)

func readProbeCursor(_ int, _ time.Duration) (int, int, error) {
	return 0, 0, fmt.Errorf("terminal cursor probing is unsupported on Windows; using split ZWJ width")
}
