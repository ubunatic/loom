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

func readChunk(_ int, _ time.Duration) ([]byte, error) {
	return nil, fmt.Errorf("terminal input polling is unsupported on Windows")
}
