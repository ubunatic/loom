// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomoji

import (
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom"
	"github.com/spf13/cobra"
)

const measurementDataDir = "docs/data/loomoji-widths"

func newCommand(runPicker func() error, runMeasure func(TerminalProfile) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "loomoji",
		Short:         "Inline searchable emoji and symbol picker",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runPicker()
		},
	}
	var measure bool
	debugCmd := &cobra.Command{
		Use:           "debug",
		Short:         "Debug loomoji terminal rendering",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if !measure {
				return fmt.Errorf("loomoji debug: specify --measure")
			}
			return runMeasure(terminalProfileFromEnv())
		},
	}
	debugCmd.Flags().BoolVar(&measure, "measure", false, "record rendered glyph widths for this terminal")
	cmd.AddCommand(debugCmd)
	return cmd
}

func terminalProfileFromEnv() TerminalProfile {
	return TerminalProfile{Term: os.Getenv("TERM"), TermProgram: os.Getenv("TERM_PROGRAM")}
}

func runMeasure(profile TerminalProfile) error {
	widget, err := newMeasureSession(profile, measurementDataDir)
	if err != nil {
		return err
	}
	pane, err := loom.New(18)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	pane.DisableDefaultQuit = true
	if err := pane.Run(widget); err != nil {
		return err
	}
	return widget.Err()
}

func newMeasureSession(profile TerminalProfile, dataDir string) (*MeasureWidget, error) {
	jsonPath, textPath := profile.Paths(dataDir)
	store, err := LoadMeasurementStore(jsonPath)
	if err != nil {
		return nil, err
	}
	store.Profile = profile
	return NewMeasureWidget(store, jsonPath, textPath), nil
}
