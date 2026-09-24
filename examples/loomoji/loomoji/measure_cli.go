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

func newCommand(runPicker func() error, runMeasure func(TerminalProfile, MeasureOptions) error) *cobra.Command {
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
	var measure, review, differsFromLoom, hasComment bool
	var width string
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
			filter, err := measureFilterFromWidth(width)
			if err != nil {
				return err
			}
			return runMeasure(terminalProfileFromEnv(), MeasureOptions{Review: review, Filter: filter, DiffersFromLoom: differsFromLoom, HasComment: hasComment})
		},
	}
	debugCmd.Flags().BoolVar(&measure, "measure", false, "record rendered glyph widths for this terminal")
	debugCmd.Flags().BoolVar(&review, "review", false, "review recorded glyph widths")
	debugCmd.Flags().StringVar(&width, "width", "", "review only: width 1, 2, 3, 4, or unsure")
	debugCmd.Flags().BoolVar(&differsFromLoom, "differs-from-loom", false, "review only: show widths that differ from Loom")
	debugCmd.Flags().BoolVar(&hasComment, "has-comment", false, "review only: show glyphs with comments")
	cmd.AddCommand(debugCmd)
	return cmd
}

func terminalProfileFromEnv() TerminalProfile {
	return TerminalProfile{Term: os.Getenv("TERM"), TermProgram: os.Getenv("TERM_PROGRAM"), VTEVersion: os.Getenv("VTE_VERSION")}
}

func measureFilterFromWidth(width string) (MeasureFilter, error) {
	switch width {
	case "":
		return MeasureFilterAll, nil
	case "1":
		return MeasureFilterWidth1, nil
	case "2":
		return MeasureFilterWidth2, nil
	case "3":
		return MeasureFilterWidth3, nil
	case "4":
		return MeasureFilterWidth4, nil
	case "unsure", "?":
		return MeasureFilterUnsure, nil
	default:
		return "", fmt.Errorf("loomoji debug: --width must be 1, 2, 3, 4, or unsure")
	}
}

func runMeasure(profile TerminalProfile, options MeasureOptions) error {
	widget, err := newMeasureSession(profile, measurementDataDir, options)
	if err != nil {
		return err
	}
	pane, err := loom.New(18)
	if err != nil {
		return err
	}
	defer pane.Close()
	configureMeasurePane(pane)
	if err := pane.Run(widget); err != nil {
		return err
	}
	return widget.Err()
}

// configureMeasurePane gives the measurement table the terminal's full width.
func configureMeasurePane(pane *loom.Pane) {
	pane.MaxCols = 0
	pane.Resizeable = true
	pane.DisableDefaultQuit = true
}

func newMeasureSession(profile TerminalProfile, dataDir string, options MeasureOptions) (*MeasureWidget, error) {
	jsonPath, textPath := profile.Paths(dataDir)
	store, err := LoadMeasurementStore(jsonPath)
	if err != nil {
		return nil, err
	}
	store.Profile = profile
	return NewMeasureWidgetWithOptions(store, jsonPath, textPath, options), nil
}
