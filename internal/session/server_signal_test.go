// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package session

import (
	"context"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

func TestServeSessionChildCanHandleHangup(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}
	t.Setenv("LOOM_SOCKET_DIR", t.TempDir())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second*ptytest.RaceFactor)
	defer cancel()
	// A shell cannot install a trap for a signal inherited as ignored. Exit 42
	// proves that the daemon leaked SIG_IGN instead of the child's default HUP.
	// Signal the daemon too: protecting it must not disable the child's HUP.
	child := `kill -HUP "$PPID"; trap 'exit 0' HUP; kill -HUP $$; exit 42`
	if err := serveSession(ctx, "child-hangup", Options{}, "/bin/sh", "-c", child); err != nil {
		t.Fatalf("child could not handle SIGHUP: %v", err)
	}
}
