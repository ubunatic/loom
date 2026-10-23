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
	return &cobra.Command{
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
			return session.Run(cmd.Context(), os.Stdin, cmd.OutOrStdout(), cmd.ErrOrStderr(), command, args...)
		},
	}
}
