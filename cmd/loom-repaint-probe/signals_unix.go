// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"
)

func notifyResizeSignals(channel chan<- os.Signal) {
	signal.Notify(channel, syscall.SIGWINCH)
}

func notifyInterruptSignals(channel chan<- os.Signal) {
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
}
