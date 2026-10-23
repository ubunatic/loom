// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/editor"
)

func editCommand() *cobra.Command {
	var configPath, themeName string
	var mouseGrab, altScreen bool
	cmd := &cobra.Command{
		Use:          "edit [file]",
		Short:        "Interactively edit a rich text or ANSI file",
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loom.LoadEditorConfig(configPath)
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("theme") {
				if !loom.ThemeExists(themeName) {
					return fmt.Errorf("unknown theme %q", themeName)
				}
				cfg.Theme = themeName
			}
			if cmd.Flags().Changed("mousegrab") {
				cfg.MouseGrab = mouseGrab
			}
			if cmd.Flags().Changed("altscreen") {
				cfg.AltScreen = altScreen
			}
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			return editor.Run(path, cfg)
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "path to editor settings YAML file")
	cmd.Flags().StringVarP(&themeName, "theme", "t", "", "color theme name")
	cmd.Flags().BoolVarP(&mouseGrab, "mousegrab", "m", loom.SpeccedDefaults.Editor.MouseGrab, "enable mouse tracking/capture")
	cmd.Flags().BoolVarP(&altScreen, "altscreen", "a", loom.SpeccedDefaults.Editor.AltScreen, "use alternate screen buffer")
	return cmd
}
