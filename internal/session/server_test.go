// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package session

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

func TestServerAttachDetachRepaintAndInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := newSessionServer(nil)
	done := make(chan error, 1)
	go func() {
		done <- runSessionWithServer(ctx, server, Options{}, "bash", "-c", "printf 'READY\\n'; read -r line; printf 'GOT:%s\\n' \"$line\"; stty size; while :; do sleep 1; done")
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("session server did not stop")
		}
	})

	client, handler := attachTestClient(t, server, 80, 24)
	defer client.Close()
	readServerStatus(t, client)
	initial := readOutputUntilAll(t, client, "\x1b[2J", "READY")
	if !strings.Contains(initial, "READY") {
		t.Fatalf("attach output lacked full repaint: %q", initial)
	}
	if err := writeMessage(client, ipcMessage{Type: msgResize, Cols: 100, Rows: 40}); err != nil {
		t.Fatal(err)
	}
	readOutputUntil(t, client, "\x1b[2J") // resize has been applied and the active screen repainted

	if err := writeMessage(client, ipcMessage{Type: msgInput, Data: []byte("from-client\n")}); err != nil {
		t.Fatal(err)
	}
	got := readOutputUntilAll(t, client, "GOT:from-client", "39 100")
	if !strings.Contains(got, "GOT:from-client") || !strings.Contains(got, "39 100") {
		t.Fatalf("child output missing relayed input: %q", got)
	}

	if err := writeMessage(client, ipcMessage{Type: msgDetach}); err != nil {
		t.Fatal(err)
	}
	detached := readUntilType(t, client, msgDetached)
	if detached.Type != msgDetached {
		t.Fatalf("detach response = %#v", detached)
	}
	if err := <-handler; err != nil {
		t.Fatalf("client handler after explicit detach: %v", err)
	}

	client2, _ := attachTestClient(t, server, 80, 24)
	defer client2.Close()
	readServerStatus(t, client2)
	repaint := readOutputUntilAll(t, client2, "\x1b[2J", "GOT:from-client")
	if !strings.Contains(repaint, "READY") {
		t.Fatalf("reattach repaint did not restore prior screen: %q", repaint)
	}
}

func TestServerPrefixDDetachesWithoutStoppingChild(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := newSessionServer(nil)
	done := make(chan error, 1)
	go func() {
		done <- runSessionWithServer(ctx, server, Options{}, "bash", "-c", "printf 'PREFIX_READY\\n'; while :; do sleep 1; done")
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("session server did not stop")
		}
	})

	client, _ := attachTestClient(t, server, 80, 24)
	defer client.Close()
	beforeStatus := readServerStatus(t, client)
	if !strings.Contains(beforeStatus, "PREFIX_READY") {
		readOutputUntil(t, client, "PREFIX_READY")
	}
	if err := writeMessage(client, ipcMessage{Type: msgInput, Data: []byte{2, 'd'}}); err != nil {
		t.Fatal(err)
	}
	_ = readUntilType(t, client, msgDetached)
	select {
	case err := <-done:
		t.Fatalf("prefix detach stopped the session: %v", err)
	default:
	}
}

func TestServerClientRelaysInputAndDetaches(t *testing.T) {
	t.Setenv("LOOM_SOCKET_DIR", t.TempDir())
	listener, err := listenSessionSocket("client-test")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	serverDone := make(chan error, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			serverDone <- acceptErr
			return
		}
		defer conn.Close()
		attach, readErr := readMessage(conn)
		if readErr != nil {
			serverDone <- readErr
			return
		}
		if attach.Type != msgAttach || attach.Cols <= 0 || attach.Rows <= 0 {
			serverDone <- io.ErrUnexpectedEOF
			return
		}
		input, readErr := readMessage(conn)
		if readErr != nil {
			serverDone <- readErr
			return
		}
		if input.Type != msgInput || string(input.Data) != "typed" {
			serverDone <- io.ErrUnexpectedEOF
			return
		}
		detach, readErr := readMessage(conn)
		if readErr != nil {
			serverDone <- readErr
			return
		}
		if detach.Type != msgDetach {
			serverDone <- io.ErrUnexpectedEOF
			return
		}
		serverDone <- nil
	}()
	if err := attachSession(context.Background(), "client-test", strings.NewReader("typed"), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if err := <-serverDone; err != nil {
		t.Fatal(err)
	}
}

func TestServerLastTabExitCleansSocket(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LOOM_SOCKET_DIR", dir)
	t.Setenv("LOOM_SESSION_SERVER_NAME", "finished")
	path := filepath.Join(dir, "session-finished.sock")
	err := RunWithOptions(context.Background(), bytes.NewReader(nil), io.Discard, io.Discard, Options{}, "bash", "-c", "exit 0")
	if err != nil {
		t.Fatalf("server returned error: %v", err)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("session socket still exists after last tab exit: %v", err)
	}
}

func TestServeSessionIgnoresWaitAttachEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LOOM_SOCKET_DIR", dir)
	t.Setenv("LOOM_SESSION_WAIT_ATTACH", "1")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second*ptytest.RaceFactor)
	defer cancel()
	if err := serveSession(ctx, "envgate", Options{}, "bash", "-c", "exit 0"); err != nil {
		t.Fatalf("server returned error: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(dir, "session-envgate.sock")); !os.IsNotExist(err) {
		t.Fatalf("session socket still exists after last tab exit: %v", err)
	}
}

func TestServeSessionWaitForAttachHoldsChildUntilAttach(t *testing.T) {
	t.Setenv("LOOM_SOCKET_DIR", t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- serveSession(ctx, "wait-test", Options{WaitForAttach: true}, "bash", "-c", "printf 'STARTED\\n'; while :; do sleep 1; done")
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second * ptytest.RaceFactor):
			t.Error("session server did not stop")
		}
	})

	path, err := sessionSocketPath("wait-test")
	if err != nil {
		t.Fatal(err)
	}

	timer := time.NewTimer(5 * time.Second * ptytest.RaceFactor)
	defer timer.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	var status SessionInfo
	for {
		status, err = SessionStatus("wait-test")
		if err == nil {
			break
		}
		select {
		case <-ticker.C:
		case <-timer.C:
			t.Fatalf("timed out waiting for session marker %s", path)
		}
	}
	if status.Tabs != 0 {
		t.Fatalf("tabs before attach = %d, want 0 while waiting for a client", status.Tabs)
	}

	select {
	case err := <-done:
		t.Fatalf("session ended while waiting for a client: %v", err)
	default:
	}

	// Output only reaches an attached client, so seeing it proves the child
	// started after the attach released the gate.
	client, err := net.Dial("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if err := writeMessage(client, ipcMessage{Type: msgAttach, Cols: 80, Rows: 24}); err != nil {
		t.Fatal(err)
	}
	_ = readServerStatus(t, client)
	if got := readOutputUntil(t, client, "STARTED"); !strings.Contains(got, "STARTED") {
		t.Fatalf("child output after attach = %q", got)
	}
}

func TestServerControlStatusListDetachAndKill(t *testing.T) {
	t.Setenv("LOOM_SOCKET_DIR", t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- serveSession(ctx, "control-test", Options{}, "bash", "-c", "printf 'CONTROL_READY\\n'; while :; do sleep 1; done")
	}()

	path, err := sessionSocketPath("control-test")
	if err != nil {
		t.Fatal(err)
	}

	timer := time.NewTimer(5 * time.Second * ptytest.RaceFactor)
	defer timer.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	var status SessionInfo
	for {
		status, err = SessionStatus("control-test")
		if err == nil {
			break
		}
		select {
		case <-ticker.C:
		case <-timer.C:
			t.Fatalf("session did not become ready: %v", err)
		}
	}
	if status.Name != "control-test" || status.Status != "detached" || status.Tabs != 1 || status.PID <= 0 {
		t.Fatalf("SessionStatus() = %#v", status)
	}
	items, err := ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Name != status.Name {
		t.Fatalf("ListSessions() = %#v", items)
	}

	client, err := net.Dial("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeMessage(client, ipcMessage{Type: msgAttach, Cols: 80, Rows: 24}); err != nil {
		t.Fatal(err)
	}
	_ = readServerStatus(t, client)
	if err := Detach("control-test"); err != nil {
		t.Fatal(err)
	}
	if got := readUntilType(t, client, msgDetached); got.Type != msgDetached {
		t.Fatalf("detach message = %#v", got)
	}
	_ = client.Close()
	status, err = SessionStatus("control-test")
	if err != nil || status.Status != "detached" {
		t.Fatalf("status after detach = %#v, %v", status, err)
	}
	if err := Kill("control-test"); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("daemon exit: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("daemon did not exit after kill")
	}
}

func attachTestClient(t *testing.T, server *sessionServer, cols, rows int) (net.Conn, <-chan error) {
	t.Helper()
	client, handler := net.Pipe()
	done := make(chan error, 1)
	go func() { done <- server.handleClient(handler) }()
	if err := writeMessage(client, ipcMessage{Type: msgAttach, Cols: cols, Rows: rows}); err != nil {
		t.Fatal(err)
	}
	return client, done
}

func readServerStatus(t *testing.T, conn net.Conn) string {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var before strings.Builder
	for {
		message, err := readMessage(conn)
		if err != nil {
			t.Fatalf("read server attach status: %v", err)
		}
		if message.Type == msgStatus && message.Status == "modes" {
			return before.String()
		}
		if message.Type != msgOutput {
			t.Fatalf("server attach message before mode status = %#v", message)
		}
		before.Write(message.Data)
	}
}

func readOutputUntil(t *testing.T, conn net.Conn, text string) string {
	return readOutputUntilAll(t, conn, text)
}

func readOutputUntilAll(t *testing.T, conn net.Conn, texts ...string) string {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var output strings.Builder
	containsAll := func() bool {
		for _, text := range texts {
			if !strings.Contains(output.String(), text) {
				return false
			}
		}
		return true
	}
	for !containsAll() {
		message, err := readMessage(conn)
		if err != nil {
			t.Fatalf("read output containing %q: %v; got %q", texts, err, output.String())
		}
		if message.Type != msgOutput {
			t.Fatalf("message while waiting for %q = %#v", texts, message)
		}
		output.Write(message.Data)
	}
	return output.String()
}

func readUntilType(t *testing.T, conn net.Conn, typ messageType) ipcMessage {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		message, err := readMessage(conn)
		if err != nil {
			t.Fatalf("read message type %q: %v", typ, err)
		}
		if message.Type == typ {
			return message
		}
		if message.Type != msgOutput {
			t.Fatalf("message while waiting for %q = %#v", typ, message)
		}
	}
}
