// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package session

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const maxIPCMessageSize = 1 << 20

var validSessionName = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

type messageType string

const (
	msgAttach      messageType = "attach"
	msgInput       messageType = "input"
	msgResize      messageType = "resize"
	msgDetach      messageType = "detach"
	msgKill        messageType = "kill"
	msgOutput      messageType = "output"
	msgStderr      messageType = "stderr"
	msgStatus      messageType = "status"
	msgStatusReply messageType = "status_reply"
	msgExit        messageType = "exit"
	msgError       messageType = "error"
)

// ipcMessage is one framed request or response on a session socket.
type ipcMessage struct {
	Type   messageType `json:"type"`
	Name   string      `json:"name,omitempty"`
	Data   []byte      `json:"data,omitempty"`
	Cols   int         `json:"cols,omitempty"`
	Rows   int         `json:"rows,omitempty"`
	PID    int         `json:"pid,omitempty"`
	Status string      `json:"status,omitempty"`
	Code   int         `json:"code,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// socketDirectory resolves the directory used for session sockets.
func socketDirectory() (string, error) {
	if dir := os.Getenv("LOOM_SOCKET_DIR"); dir != "" {
		return filepath.Clean(dir), nil
	}
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return filepath.Join(dir, "loom"), nil
	}
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, ".local", "state", "loom"), nil
	}
	if uid := os.Getuid(); uid >= 0 {
		return filepath.Join(os.TempDir(), "loom-"+strconv.Itoa(uid)), nil
	}
	return "", fmt.Errorf("resolve session socket directory: %w", err)
}

// sessionSocketPath returns the socket path for a validated session name.
func sessionSocketPath(name string) (string, error) {
	if !validSessionName.MatchString(name) {
		return "", fmt.Errorf("invalid session name %q", name)
	}
	dir, err := socketDirectory()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "session-"+name+".sock")
	// Linux sockaddr_un.sun_path has 108 bytes including the NUL terminator.
	if len(path) >= 108 {
		return "", fmt.Errorf("session socket path is too long (%d bytes)", len(path))
	}
	return path, nil
}

// listenSessionSocket creates a private listener, removing only stale sockets.
func listenSessionSocket(name string) (net.Listener, error) {
	path, err := sessionSocketPath(name)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create session socket directory: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return nil, fmt.Errorf("secure session socket directory: %w", err)
	}
	if err := removeStaleSocket(path); err != nil {
		return nil, err
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("listen on session socket: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		listener.Close()
		os.Remove(path)
		return nil, fmt.Errorf("secure session socket: %w", err)
	}
	return &sessionListener{Listener: listener, path: path}, nil
}

type sessionListener struct {
	net.Listener
	path      string
	closeOnce sync.Once
	closeErr  error
}

func (l *sessionListener) Close() error {
	l.closeOnce.Do(func() {
		closeErr := l.Listener.Close()
		removeErr := os.Remove(l.path)
		if errors.Is(removeErr, os.ErrNotExist) {
			removeErr = nil
		}
		l.closeErr = errors.Join(closeErr, removeErr)
	})
	return l.closeErr
}

func removeStaleSocket(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect session socket: %w", err)
	}
	if info.Mode()&os.ModeSocket == 0 {
		return fmt.Errorf("session socket path exists and is not a socket: %s", path)
	}
	conn, err := net.DialTimeout("unix", path, 100*time.Millisecond)
	if err == nil {
		conn.Close()
		return fmt.Errorf("session socket is already active: %s", path)
	}
	if !errors.Is(err, syscall.ECONNREFUSED) {
		return fmt.Errorf("probe session socket %s: %w", path, err)
	}
	// Recheck the inode before unlinking so a concurrent replacement is retained.
	current, statErr := os.Lstat(path)
	if errors.Is(statErr, os.ErrNotExist) {
		return nil
	}
	if statErr != nil {
		return fmt.Errorf("recheck session socket: %w", statErr)
	}
	if !os.SameFile(info, current) {
		return fmt.Errorf("session socket changed during stale check: %s", path)
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove stale session socket: %w", err)
	}
	return nil
}

// readMessage reads one length-prefixed JSON message from r.
func readMessage(r io.Reader) (ipcMessage, error) {
	var message ipcMessage
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return message, err
	}
	size := binary.BigEndian.Uint32(header[:])
	if size == 0 || size > maxIPCMessageSize {
		return message, fmt.Errorf("invalid IPC message size %d", size)
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(r, data); err != nil {
		return message, fmt.Errorf("read IPC message body: %w", err)
	}
	if err := json.Unmarshal(data, &message); err != nil {
		return message, fmt.Errorf("decode IPC message: %w", err)
	}
	if !validMessageType(message.Type) {
		return message, fmt.Errorf("unknown IPC message type %q", message.Type)
	}
	return message, nil
}

// writeMessage writes one length-prefixed JSON message to w.
func writeMessage(w io.Writer, message ipcMessage) error {
	if !validMessageType(message.Type) {
		return fmt.Errorf("unknown IPC message type %q", message.Type)
	}
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("encode IPC message: %w", err)
	}
	if len(data) == 0 || len(data) > maxIPCMessageSize {
		return fmt.Errorf("invalid IPC message size %d", len(data))
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(data)))
	if err := writeAll(w, header[:]); err != nil {
		return fmt.Errorf("write IPC message header: %w", err)
	}
	if err := writeAll(w, data); err != nil {
		return fmt.Errorf("write IPC message body: %w", err)
	}
	return nil
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if n > 0 {
			data = data[n:]
		}
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}

func validMessageType(t messageType) bool {
	return strings.Contains(" attach input resize detach kill output stderr status status_reply exit error ", " "+string(t)+" ")
}
