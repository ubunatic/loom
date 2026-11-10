// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package session

import (
	"context"
	"net"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

func TestBackpressure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := newSessionServer(nil)
	done := make(chan error, 1)

	// Create a child that produces output.
	script := "while :; do echo lots_of_data; done"

	go func() {
		done <- runSessionWithServer(ctx, server, Options{}, "bash", "-c", script)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second * ptytest.RaceFactor):
			t.Error("session server did not stop")
		}
	})

	// Attach a client but never read from it.
	client, handler := net.Pipe()
	defer client.Close()

	handlerDone := make(chan error, 1)
	go func() { handlerDone <- server.handleClient(handler) }()

	// Send attach message.
	if err := writeMessage(client, ipcMessage{Type: msgAttach, Cols: 80, Rows: 24}); err != nil {
		t.Fatal(err)
	}

	// Wait for the server to acknowledge the connection and try to write output.
	// We read nothing from the client, so net.Pipe() blocks writes instantly.
	// Since server's writeOutput holds writeMu and writes to this client,
	// writeMu will block. Then detach() will also block.

	// We wait until detach blocks. We wait for a short duration to ensure
	// the attach is processed and the server has started writing, using polling loop.

	detachChan := make(chan struct{})
	go func() {
		// Use a ticker for polling wait
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			server.mu.Lock()
			conn := server.conn
			server.mu.Unlock()
			if conn != nil {
				break
			}
			<-ticker.C
		}

		server.detach(true)
		close(detachChan)
	}()

	select {
	case <-detachChan:
		// Passed? Then the bug didn't reproduce (it didn't block)
		// Or the bug is fixed.
		// We actually WANT it to timeout and fail on current code.
	case <-time.After(2 * time.Second * ptytest.RaceFactor):
		t.Fatal("server.detach() blocked")
	}
}
