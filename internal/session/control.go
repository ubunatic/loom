// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package session

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

// SessionInfo describes a live session daemon.
type SessionInfo struct {
	Name   string
	Tabs   int
	Status string
	PID    int
}

// Attach connects a foreground terminal to a persistent session.
func Attach(ctx context.Context, name string, inReader io.Reader, outWriter io.Writer) error {
	return attachSession(ctx, name, inReader, outWriter)
}

// SessionStatus queries a session daemon and reports its current state.
func SessionStatus(name string) (SessionInfo, error) {
	reply, err := sessionRequest(name, ipcMessage{Type: msgStatus})
	if err != nil {
		return SessionInfo{}, err
	}
	if reply.Type != msgStatusReply {
		return SessionInfo{}, fmt.Errorf("unexpected status response %q", reply.Type)
	}
	return SessionInfo{Name: reply.Name, Tabs: reply.Tabs, Status: reply.Status, PID: reply.PID}, nil
}

// Detach requests that the foreground client disconnect from a session.
func Detach(name string) error {
	_, err := sessionRequest(name, ipcMessage{Type: msgDetach})
	return err
}

// Kill terminates a session daemon and the child process groups it owns.
func Kill(name string) error {
	_, err := sessionRequest(name, ipcMessage{Type: msgKill})
	return err
}

// ListSessions returns active session daemons, sorted by name.
func ListSessions() ([]SessionInfo, error) {
	dir, err := socketDirectory()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read session socket directory: %w", err)
	}
	var sessions []SessionInfo
	for _, entry := range entries {
		name := strings.TrimSuffix(strings.TrimPrefix(entry.Name(), "session-"), ".sock")
		if name == entry.Name() || !validSessionName.MatchString(name) {
			continue
		}
		info, statusErr := SessionStatus(name)
		if statusErr != nil {
			if errors.Is(statusErr, syscall.ECONNREFUSED) {
				path := filepath.Join(dir, entry.Name())
				_ = os.Remove(path)
				continue
			}
			continue
		}
		sessions = append(sessions, info)
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].Name < sessions[j].Name })
	return sessions, nil
}

func sessionRequest(name string, request ipcMessage) (ipcMessage, error) {
	path, err := sessionSocketPath(name)
	if err != nil {
		return ipcMessage{}, err
	}
	conn, err := net.DialTimeout("unix", path, 500*time.Millisecond)
	if err != nil {
		return ipcMessage{}, fmt.Errorf("connect to session %q: %w", name, err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if err := writeMessage(conn, request); err != nil {
		return ipcMessage{}, err
	}
	reply, err := readMessage(conn)
	if err != nil {
		return ipcMessage{}, err
	}
	if reply.Type == msgError {
		return ipcMessage{}, errors.New(reply.Error)
	}
	return reply, nil
}
