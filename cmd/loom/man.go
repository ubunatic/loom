// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"ubunatic.com/loom"
)

func manCommand(root *cobra.Command) *cobra.Command {
	var install bool
	var dir string
	cmd := &cobra.Command{
		Use:   "man",
		Short: "Generate or install Loom manual pages",
		Long:  "Generate roff manual pages from Loom's Cobra commands. Without flags, print the loom(1) page to standard output. Use --install to install every page into the standard man directory, or --dir to choose a directory.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if install || dir != "" {
				return installManPages(root, dir, cmd.OutOrStdout())
			}
			return doc.GenMan(root, manHeader(), cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&install, "install", false, "install all manual pages into the standard man directory")
	cmd.Flags().StringVar(&dir, "dir", "", "install all manual pages into this directory (implies --install)")
	return cmd
}

func manHeader() *doc.GenManHeader {
	return &doc.GenManHeader{
		Title:   "LOOM",
		Section: "1",
		Source:  "loom " + loom.Version,
		Manual:  "Loom Manual",
	}
}

func defaultManDir() string {
	return defaultManDirFor(os.Geteuid(), os.Getenv("XDG_DATA_HOME"), homeDir())
}

func homeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

func defaultManDirFor(euid int, xdgDataHome, home string) string {
	if euid == 0 {
		return "/usr/local/share/man/man1"
	}
	if xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "man", "man1")
	}
	if home != "" {
		return filepath.Join(home, ".local", "share", "man", "man1")
	}
	return "/usr/local/share/man/man1"
}

func installManPages(root *cobra.Command, targetDir string, out io.Writer) error {
	if targetDir == "" {
		targetDir = defaultManDir()
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("creating man directory %s: %w", targetDir, err)
	}
	if err := doc.GenManTree(root, manHeader(), targetDir); err != nil {
		return fmt.Errorf("generating man pages in %s: %w", targetDir, err)
	}
	_, err := fmt.Fprintf(out, "Installed Loom man pages to %s\nRun 'man loom' to view.\n", targetDir)
	return err
}
