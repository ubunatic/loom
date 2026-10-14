// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build windows

package main

import (
	"os"
	"os/signal"
)

func notifyResizeSignals(_ chan<- os.Signal) {}

func notifyInterruptSignals(channel chan<- os.Signal) {
	signal.Notify(channel, os.Interrupt)
}
