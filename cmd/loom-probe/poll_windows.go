// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build windows

package main

import (
	"fmt"
	"time"
)

func readCursorReport(_ int, _ time.Duration) (int, int, error) {
	return 0, 0, fmt.Errorf("terminal cursor polling is unsupported on Windows")
}
