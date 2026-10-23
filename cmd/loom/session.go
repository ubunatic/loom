// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package main

import (
	"os"

	"github.com/spf13/cobra"
	"ubunatic.com/loom/internal/session"
)

func sessionCommand() *cobra.Command {
	debugDefault := os.Getenv("LOOM_SESSION_DEBUG") == "1"
	debugDirDefault := os.Getenv("LOOM_DEBUG_DIR")
	var debug bool
	var debugDir string
	command := &cobra.Command{
		Use:          "session [-- <cmd> [args...]]",
		Short:        "Run a command in a PTY wrapper with screenshot and quit hotkeys",
		Long:         session.LongHelp(),
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			command := ""
			if len(args) > 0 {
				command = args[0]
				args = args[1:]
			}
			if cmd.Flags().Changed("debug") {
				debugDefault = debug
			}
			if cmd.Flags().Changed("debug-dir") {
				debugDirDefault = debugDir
			}
			return session.RunWithOptions(cmd.Context(), os.Stdin, cmd.OutOrStdout(), cmd.ErrOrStderr(), session.Options{Debug: debugDefault, DebugDir: debugDirDefault}, command, args...)
		},
	}
	command.Flags().BoolVar(&debug, "debug", debugDefault, "enable session stream diagnostics")
	command.Flags().StringVar(&debugDir, "debug-dir", debugDirDefault, "directory for session diagnostics")
	return command
}
