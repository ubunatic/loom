// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"ubunatic.com/loom/internal/session"
)

func sessionCommand() *cobra.Command {
	debugDefault := os.Getenv("LOOM_SESSION_DEBUG") == "1"
	debugDirDefault := os.Getenv("LOOM_DEBUG_DIR")
	var debug bool
	var debugDir string
	var sessionName string
	var command *cobra.Command
	command = &cobra.Command{
		Use:          "session [-- <cmd> [args...]]",
		Short:        "Run, attach to, and manage persistent terminal sessions",
		Long:         session.LongHelp(),
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("debug") {
				debugDefault = debug
			}
			if cmd.Flags().Changed("debug-dir") {
				debugDirDefault = debugDir
			}
			options := session.Options{Debug: debugDefault, DebugDir: debugDirDefault}
			if os.Getenv("LOOM_SESSION_SERVER_NAME") != "" {
				program := ""
				if len(args) > 0 {
					program, args = args[0], args[1:]
				}
				return session.RunWithOptions(cmd.Context(), os.Stdin, cmd.OutOrStdout(), cmd.ErrOrStderr(), options, program, args...)
			}
			if len(args) == 0 {
				name := sessionName
				if name == "" {
					name = "default"
				}
				if _, err := session.SessionStatus(name); err == nil {
					return session.Attach(cmd.Context(), name, os.Stdin, cmd.OutOrStdout())
				}
				return startAndAttach(cmd.Context(), name, options, "")
			}
			program, args := args[0], args[1:]
			name := sessionName
			if name == "" {
				name = uniqueCommandSessionName()
			}
			if _, err := session.SessionStatus(name); err == nil {
				return fmt.Errorf("session %q already exists", name)
			}
			return startAndAttach(cmd.Context(), name, options, program, args...)
		},
	}
	command.PersistentFlags().StringVarP(&sessionName, "session", "s", "", "session name (default: default for interactive sessions)")
	command.Flags().BoolVar(&debug, "debug", debugDefault, "enable session stream diagnostics")
	command.Flags().StringVar(&debugDir, "debug-dir", debugDirDefault, "directory for session diagnostics")

	command.AddCommand(&cobra.Command{
		Use: "attach", Aliases: []string{"a"}, Short: "Attach to a running session", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			name, err := resolveSessionName(sessionName)
			if err != nil {
				return err
			}
			return session.Attach(cmd.Context(), name, os.Stdin, cmd.OutOrStdout())
		},
	})
	command.AddCommand(&cobra.Command{
		Use: "detach", Short: "Detach the foreground client", Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			name, err := resolveSessionName(sessionName)
			if err != nil {
				return err
			}
			return session.Detach(name)
		},
	})
	command.AddCommand(&cobra.Command{
		Use: "list", Aliases: []string{"ls"}, Short: "List active sessions", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			items, err := session.ListSessions()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(items) == 0 {
				_, err = io.WriteString(out, "No active sessions.\n")
				return err
			}
			_, _ = fmt.Fprintln(out, "NAME\tTABS\tSTATUS")
			for _, item := range items {
				if _, err := fmt.Fprintf(out, "%s\t%d\t%s\n", item.Name, item.Tabs, item.Status); err != nil {
					return err
				}
			}
			return nil
		},
	})
	command.AddCommand(&cobra.Command{
		Use: "kill", Short: "Kill a session and its child processes", Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			name, err := resolveSessionName(sessionName)
			if err != nil {
				return err
			}
			return session.Kill(name)
		},
	})
	return command
}

func startAndAttach(ctx context.Context, name string, options session.Options, program string, args ...string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find loom executable: %w", err)
	}
	childArgs := []string{"session", "--"}
	if program != "" {
		childArgs = append(childArgs, program)
	}
	childArgs = append(childArgs, args...)
	child := exec.Command(executable, childArgs...)
	child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open daemon stdio: %w", err)
	}
	defer devNull.Close()
	child.Stdin, child.Stdout, child.Stderr = devNull, devNull, devNull
	child.Env = setEnvironment(os.Environ(), "LOOM_SESSION_SERVER_NAME", name)
	child.Env = setEnvironment(child.Env, "LOOM_SESSION_WAIT_ATTACH", "1")
	if options.Debug {
		child.Env = setEnvironment(child.Env, "LOOM_SESSION_DEBUG", "1")
	}
	if options.DebugDir != "" {
		child.Env = setEnvironment(child.Env, "LOOM_DEBUG_DIR", options.DebugDir)
	}
	if err := child.Start(); err != nil {
		return fmt.Errorf("start session daemon: %w", err)
	}
	childDone := make(chan error, 1)
	go func() { childDone <- child.Wait() }()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		if _, err := session.SessionStatus(name); err == nil {
			return session.Attach(ctx, name, os.Stdin, os.Stdout)
		}
		select {
		case err := <-childDone:
			if err == nil {
				return errors.New("session daemon exited before becoming ready")
			}
			return fmt.Errorf("session daemon failed to start: %w", err)
		case <-deadline.C:
			return fmt.Errorf("timed out waiting for session %q daemon", name)
		case <-ticker.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func setEnvironment(environment []string, key, value string) []string {
	filtered := environment[:0]
	prefix := key + "="
	for _, item := range environment {
		if !strings.HasPrefix(item, prefix) {
			filtered = append(filtered, item)
		}
	}
	return append(filtered, key+"="+value)
}

func uniqueCommandSessionName() string {
	return "cmd-" + strconv.Itoa(os.Getpid()) + "-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

func resolveSessionName(requested string) (string, error) {
	if requested != "" {
		return requested, nil
	}
	items, err := session.ListSessions()
	if err != nil {
		return "", err
	}
	for _, item := range items {
		if item.Name == "default" {
			return item.Name, nil
		}
	}
	if len(items) == 1 {
		return items[0].Name, nil
	}
	if len(items) == 0 {
		return "", errors.New("no active session; start one with 'loom session'")
	}
	return "", errors.New("multiple active sessions; specify one with --session")
}
