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
	"sync"
	"time"
)

type attachEvent struct {
	cols   int
	rows   int
	attach bool
}

type sessionServer struct {
	name     string
	listener net.Listener
	inputR   *io.PipeReader
	inputW   *io.PipeWriter
	output   *serverOutput
	attached chan attachEvent
	kill     chan struct{}
	killOnce sync.Once
	finished chan struct{}
	mu       sync.Mutex
	conn     net.Conn
	ready    bool
	tabCount int
	writeMu  sync.Mutex
}

type serverOutput struct {
	server *sessionServer
	direct bool
}

func newSessionServer(listener net.Listener) *sessionServer {
	inputR, inputW := io.Pipe()
	s := &sessionServer{
		listener: listener,
		inputR:   inputR,
		inputW:   inputW,
		attached: make(chan attachEvent, 1),
		kill:     make(chan struct{}),
	}
	s.output = &serverOutput{server: s}
	return s
}

func (s *sessionServer) writeMessage(message ipcMessage) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return nil
	}
	return writeMessage(conn, message)
}

func (s *sessionServer) writeOutput(message ipcMessage, direct bool) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	s.mu.Lock()
	conn, ready := s.conn, s.ready
	s.mu.Unlock()
	if conn == nil || (!ready && !direct) {
		return nil
	}
	return writeMessage(conn, message)
}

func (s *sessionServer) markReady() {
	s.mu.Lock()
	s.ready = true
	s.mu.Unlock()
}

func (s *sessionServer) setTabCount(count int) {
	s.mu.Lock()
	s.tabCount = count
	s.mu.Unlock()
}

func (s *sessionServer) attach(conn net.Conn, cols, rows int) error {
	s.mu.Lock()
	if s.conn != nil {
		s.mu.Unlock()
		return errors.New("session already has an attached client")
	}
	s.conn = conn
	s.ready = false
	s.mu.Unlock()
	select {
	case s.attached <- attachEvent{cols: cols, rows: rows, attach: true}:
		return nil
	case <-s.finished:
		s.detach(false)
		return os.ErrProcessDone
	}
}

func (s *sessionServer) detach(notify bool) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	s.mu.Lock()
	conn := s.conn
	s.conn = nil
	s.ready = false
	s.mu.Unlock()
	if conn == nil {
		return
	}
	if notify {
		_ = writeMessage(conn, ipcMessage{Type: msgDetached})
	}
	_ = conn.Close()
}

func (w *serverOutput) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if err := w.server.writeOutput(ipcMessage{Type: msgOutput, Data: append([]byte(nil), p...)}, w.direct); err != nil {
		w.server.detach(false)
		return len(p), nil // a lost client must never stop PTY output processing
	}
	return len(p), nil
}

// serveSession owns a detached session until its last tab exits or ctx ends.
func serveSession(ctx context.Context, name string, options Options, command string, args ...string) error {
	listener, err := listenSessionSocket(name)
	if err != nil {
		return err
	}
	server := newSessionServer(listener)
	server.name = name
	server.finished = make(chan struct{})
	defer func() {
		server.detach(false)
		_ = server.inputR.Close()
		_ = server.inputW.Close()
		_ = listener.Close()
	}()

	runDone := make(chan error, 1)
	go func() {
		err := runSessionWithServer(ctx, server, options, command, args...)
		runDone <- err
		close(server.finished)
	}()

	for {
		accepted := make(chan struct {
			conn net.Conn
			err  error
		}, 1)
		go func() {
			conn, acceptErr := listener.Accept()
			accepted <- struct {
				conn net.Conn
				err  error
			}{conn: conn, err: acceptErr}
		}()
		select {
		case err := <-runDone:
			return err
		case <-ctx.Done():
			return ctx.Err()
		case result := <-accepted:
			if result.err != nil {
				if ctx.Err() != nil || errors.Is(result.err, net.ErrClosed) {
					return ctx.Err()
				}
				return fmt.Errorf("accept session client: %w", result.err)
			}
			go func() { _ = server.handleClient(result.conn) }()
		}
	}
}

func (s *sessionServer) handleClient(conn net.Conn) error {
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	first, err := readMessage(conn)
	if err != nil {
		conn.Close()
		return err
	}
	if first.Type == msgStatus {
		s.mu.Lock()
		status := "detached"
		if s.conn != nil {
			status = "attached"
		}
		reply := ipcMessage{Type: msgStatusReply, Name: s.name, Status: status, Tabs: s.tabCount, PID: os.Getpid()}
		s.mu.Unlock()
		_ = writeMessage(conn, reply)
		return conn.Close()
	}
	if first.Type == msgDetach {
		s.mu.Lock()
		attached := s.conn != nil
		s.mu.Unlock()
		if !attached {
			_ = writeMessage(conn, ipcMessage{Type: msgError, Error: "session has no attached client"})
			return conn.Close()
		}
		s.detach(true)
		_ = writeMessage(conn, ipcMessage{Type: msgDetached})
		return conn.Close()
	}
	if first.Type == msgKill {
		s.killOnce.Do(func() { close(s.kill) })
		if s.finished != nil {
			<-s.finished
		}
		_ = writeMessage(conn, ipcMessage{Type: msgExit})
		return conn.Close()
	}
	if first.Type != msgAttach {
		_ = writeMessage(conn, ipcMessage{Type: msgError, Error: "first session message must be attach"})
		conn.Close()
		return errors.New("invalid initial session message")
	}
	if err := s.attach(conn, first.Cols, first.Rows); err != nil {
		_ = writeMessage(conn, ipcMessage{Type: msgError, Error: err.Error()})
		conn.Close()
		return err
	}
	_ = conn.SetReadDeadline(time.Time{})
	defer s.detach(false)
	for {
		message, err := readMessage(conn)
		if err != nil {
			return err
		}
		switch message.Type {
		case msgInput:
			if _, err := s.inputW.Write(message.Data); err != nil {
				return err
			}
		case msgResize:
			select {
			case s.attached <- attachEvent{cols: message.Cols, rows: message.Rows}:
			case <-s.finished:
				return os.ErrProcessDone
			}
		case msgDetach:
			s.detach(true)
			return nil
		case msgKill:
			s.killOnce.Do(func() { close(s.kill) })
			if s.finished != nil {
				<-s.finished
			}
			return os.ErrProcessDone
		default:
			_ = s.writeMessage(ipcMessage{Type: msgError, Error: "unexpected client message"})
		}
	}
}

func serverAttachEvents(server *sessionServer) <-chan attachEvent {
	if server == nil {
		return nil
	}
	return server.attached
}

func serverKillEvents(server *sessionServer) <-chan struct{} {
	if server == nil {
		return nil
	}
	return server.kill
}
