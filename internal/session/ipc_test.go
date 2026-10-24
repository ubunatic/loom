// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package session

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIPCMessageRoundTrip(t *testing.T) {
	want := ipcMessage{Type: msgAttach, Name: "dev", Cols: 100, Rows: 40}
	var buf bytes.Buffer
	if err := writeMessage(&buf, want); err != nil {
		t.Fatal(err)
	}
	got, err := readMessage(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != want.Type || got.Name != want.Name || got.Cols != want.Cols || got.Rows != want.Rows {
		t.Fatalf("readMessage() = %#v, want %#v", got, want)
	}
}

type fragmentReader struct {
	r io.Reader
	n int
}

func (r fragmentReader) Read(p []byte) (int, error) {
	if len(p) > r.n {
		p = p[:r.n]
	}
	return r.r.Read(p)
}

func TestIPCMessageFragmentedFrame(t *testing.T) {
	var buf bytes.Buffer
	want := ipcMessage{Type: msgOutput, Data: []byte("fragmented output")}
	if err := writeMessage(&buf, want); err != nil {
		t.Fatal(err)
	}
	got, err := readMessage(fragmentReader{r: &buf, n: 1})
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != want.Type || !bytes.Equal(got.Data, want.Data) {
		t.Fatalf("readMessage() = %#v, want %#v", got, want)
	}
}

func TestIPCMessageRejectsMalformedFrames(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: []byte{0, 0, 0, 0}},
		{name: "oversize", data: []byte{0, 0x10, 0, 1}},
		{name: "truncated header", data: []byte{0, 0}},
		{name: "truncated body", data: []byte{0, 0, 0, 2, '{'}},
		{name: "unknown type", data: framedJSON(`{"type":"bogus"}`)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := readMessage(bytes.NewReader(test.data)); err == nil {
				t.Fatal("readMessage() succeeded, want error")
			}
		})
	}
}

func TestIPCWriteMessageHandlesFragmentedAndShortWrites(t *testing.T) {
	var buf bytes.Buffer
	if err := writeMessage(fragmentWriter{w: &buf, n: 1}, ipcMessage{Type: msgStatus}); err != nil {
		t.Fatal(err)
	}
	if _, err := readMessage(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatal(err)
	}
	if err := writeMessage(zeroWriter{}, ipcMessage{Type: msgStatus}); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("writeMessage() error = %v, want %v", err, io.ErrShortWrite)
	}
}

type fragmentWriter struct {
	w io.Writer
	n int
}

func (w fragmentWriter) Write(p []byte) (int, error) {
	if len(p) > w.n {
		p = p[:w.n]
	}
	return w.w.Write(p)
}

type zeroWriter struct{}

func (zeroWriter) Write([]byte) (int, error) { return 0, nil }

func TestIPCMessageTypes(t *testing.T) {
	types := []messageType{msgAttach, msgInput, msgResize, msgDetach, msgKill, msgOutput, msgStderr, msgStatus, msgStatusReply, msgExit, msgError}
	for _, typ := range types {
		if !validMessageType(typ) {
			t.Errorf("validMessageType(%q) = false", typ)
		}
	}
	if validMessageType("unknown") {
		t.Fatal("validMessageType(unknown) = true")
	}
}

func TestIPCSocketDirectoryResolution(t *testing.T) {
	t.Setenv("LOOM_SOCKET_DIR", "/custom/loom-sockets")
	t.Setenv("XDG_RUNTIME_DIR", "/run/user/1000")
	t.Setenv("HOME", "/home/tester")
	dir, err := socketDirectory()
	if err != nil || dir != "/custom/loom-sockets" {
		t.Fatalf("socketDirectory() = %q, %v", dir, err)
	}

	t.Setenv("LOOM_SOCKET_DIR", "")
	dir, err = socketDirectory()
	if err != nil || dir != "/run/user/1000/loom" {
		t.Fatalf("socketDirectory() with XDG = %q, %v", dir, err)
	}

	t.Setenv("XDG_RUNTIME_DIR", "")
	dir, err = socketDirectory()
	if err != nil || dir != filepath.Join("/home/tester", ".local", "state", "loom") {
		t.Fatalf("socketDirectory() with HOME = %q, %v", dir, err)
	}
}

func TestIPCSessionSocketPathValidatesNameAndLength(t *testing.T) {
	t.Setenv("LOOM_SOCKET_DIR", t.TempDir())
	for _, name := range []string{"", "..", "../escape", "has space", "a/b", strings.Repeat("a", 65)} {
		if _, err := sessionSocketPath(name); err == nil {
			t.Errorf("sessionSocketPath(%q) succeeded", name)
		}
	}
	if path, err := sessionSocketPath("dev-1_test"); err != nil || filepath.Base(path) != "session-dev-1_test.sock" {
		t.Fatalf("sessionSocketPath(valid) = %q, %v", path, err)
	}
	t.Setenv("LOOM_SOCKET_DIR", filepath.Join(string(filepath.Separator), strings.Repeat("x", 105)))
	if _, err := sessionSocketPath("dev"); err == nil {
		t.Fatal("sessionSocketPath() accepted an overlong path")
	}
}

func TestIPCListenerRejectsLiveSocket(t *testing.T) {
	t.Setenv("LOOM_SOCKET_DIR", t.TempDir())
	listener, err := listenSessionSocket("live")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	if _, err := listenSessionSocket("live"); err == nil || !strings.Contains(err.Error(), "already active") {
		t.Fatalf("second listen error = %v, want active socket error", err)
	}
}

func TestIPCListenerRemovesStaleSocket(t *testing.T) {
	t.Setenv("LOOM_SOCKET_DIR", t.TempDir())
	path, err := sessionSocketPath("stale")
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	listener.(*net.UnixListener).SetUnlinkOnClose(false)
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSocket == 0 {
		t.Fatal("expected stale socket path")
	}

	newListener, err := listenSessionSocket("stale")
	if err != nil {
		t.Fatal(err)
	}
	defer newListener.Close()
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("new listener socket missing: %v", err)
	}
}

func TestIPCListenerUsesPrivatePermissions(t *testing.T) {
	t.Setenv("LOOM_SOCKET_DIR", filepath.Join(t.TempDir(), "sockets"))
	listener, err := listenSessionSocket("private")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	path, _ := sessionSocketPath("private")
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if dirInfo.Mode().Perm() != 0o700 {
		t.Errorf("socket directory permissions = %o, want 700", dirInfo.Mode().Perm())
	}
	socketInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if socketInfo.Mode().Perm() != 0o600 {
		t.Errorf("socket permissions = %o, want 600", socketInfo.Mode().Perm())
	}
}

func framedJSON(body string) []byte {
	data := []byte(body)
	frame := make([]byte, 4+len(data))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(data)))
	copy(frame[4:], data)
	return frame
}

func TestIPCListenerCloseRemovesSocket(t *testing.T) {
	t.Setenv("LOOM_SOCKET_DIR", t.TempDir())
	listener, err := listenSessionSocket("cleanup")
	if err != nil {
		t.Fatal(err)
	}
	path, _ := sessionSocketPath("cleanup")
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("socket remains after close: %v", err)
	}
}
