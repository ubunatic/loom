// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build linux

package view

import (
	"bytes"
	"os"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	"ubunatic.com/loom/internal/ptytest"
	"os/signal"
)

func TestRelayDuplexAndEINTR(t *testing.T) {
	// Use socketpair to get full-duplex pipes
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	if err != nil {
		t.Fatal(err)
	}
	masterEnd := os.NewFile(uintptr(fds[0]), "master")
	childEnd := os.NewFile(uintptr(fds[1]), "child")
	defer masterEnd.Close()
	defer childEnd.Close()

	stdinFds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	if err != nil {
		t.Fatal(err)
	}
	stdinEnd := os.NewFile(uintptr(stdinFds[0]), "stdin")
	observerEnd := os.NewFile(uintptr(stdinFds[1]), "observer")
	defer stdinEnd.Close()
	defer observerEnd.Close()

	var rendered bool
	var mu sync.Mutex
	renderCond := sync.NewCond(&mu)

	dsrCalled := make(chan struct{})

	relay := &Relay{
		Master: masterEnd,
		Stdin:  stdinEnd,
		DSR: func() []byte {
			select {
			case <-dsrCalled:
				return nil
			default:
				close(dsrCalled)
				return []byte("DSR")
			}
		},
		Render: func() {
			mu.Lock()
			rendered = true
			renderCond.Signal()
			mu.Unlock()
		},
	}

	errc := make(chan error, 1)
	go func() {
		errc <- relay.Run()
	}()

	// Read DSR that relay wrote to masterEnd
	buf := make([]byte, 1024)
	n, err := childEnd.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf[:n], []byte("DSR")) {
		t.Errorf("Expected DSR to be written to master, got: %s", buf[:n])
	}

	// 1. Observer -> Child (observerEnd -> stdinEnd -> Relay -> masterEnd -> childEnd)
	_, err = observerEnd.Write([]byte("input"))
	if err != nil {
		t.Fatal(err)
	}
	n, err = childEnd.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf[:n], []byte("input")) {
		t.Errorf("Expected 'input' to reach child, got: %s", buf[:n])
	}

	// 2. Child -> Observer (childEnd -> masterEnd -> Relay -> Render())
	_, err = childEnd.Write([]byte("output"))
	if err != nil {
		t.Fatal(err)
	}

	// Wait for render (using Condition variable instead of sleep)
	mu.Lock()
	done := make(chan struct{})
	go func() {
		mu.Lock()
		defer mu.Unlock()
		for !rendered {
			renderCond.Wait()
		}
		close(done)
	}()
	mu.Unlock()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond * ptytest.RaceFactor):
		t.Errorf("Expected child output to trigger Render()")
	}

	// 3. Test EINTR survival
	// Catch SIGUSR1 before sending it to avoid killing the process.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, unix.SIGUSR1)

	unix.Kill(unix.Getpid(), unix.SIGUSR1)

	// Wait for signal delivery via channel
	select {
	case <-sigCh:
	case <-time.After(100 * time.Millisecond * ptytest.RaceFactor):
		t.Errorf("Failed to receive SIGUSR1")
	}

	signal.Stop(sigCh)

	// Ensure relay is still alive
	_, err = observerEnd.Write([]byte("alive"))
	if err != nil {
		t.Fatal(err)
	}
	n, err = childEnd.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf[:n], []byte("alive")) {
		t.Errorf("Expected relay to survive EINTR, got: %s", buf[:n])
	}

	// 4. Test Stop on 0x03
	_, err = observerEnd.Write([]byte{0x03})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-errc:
		if err != nil {
			t.Errorf("Expected normal exit on 0x03, got: %v", err)
		}
	case <-time.After(100 * time.Millisecond * ptytest.RaceFactor):
		t.Errorf("Relay did not stop after receiving 0x03")
	}
}
