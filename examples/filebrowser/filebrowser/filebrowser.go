// Package filebrowser demonstrates a file list and live metadata in split panes.
package filebrowser

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"codeberg.org/ubunatic/loom"
)

// Run parses args and runs the filebrowser example. It returns flag.ErrHelp
// when -h/--help was requested, matching the standard flag package
// convention so callers can treat that as a clean, non-error exit.
func Run(args []string) error {
	flags := flag.NewFlagSet("filebrowser", flag.ContinueOnError)
	themeName := flags.String("theme", "mc", "color theme")
	if err := flags.Parse(args); err != nil {
		return err
	}
	theme, err := resolveTheme(*themeName)
	if err != nil {
		return err
	}
	dir := "."
	if flags.NArg() > 0 {
		dir = flags.Arg(0)
	}
	app, err := newBrowser(dir, *themeName, theme)
	if err != nil {
		return err
	}
	// Request more rows than any terminal can provide; Loom clamps this to the
	// available height, keeping the browser full-screen while still adapting to
	// terminal resizes.
	pane, err := loom.New(1 << 16)
	if err != nil {
		return err
	}
	defer pane.Close()
	configurePane(pane)
	pane.EnableMouseClicks()
	return pane.Run(app)
}

func resolveTheme(name string) (loom.ThemeColors, error) {
	theme, ok := loom.SpeccedThemes[name]
	if ok {
		return theme, nil
	}
	names := themeNames()
	return loom.ThemeColors{}, fmt.Errorf("filebrowser: unknown theme %q (available: %s)", name, strings.Join(names, ", "))
}

func themeNames() []string {
	names := make([]string, 0, len(loom.SpeccedThemes))
	for name := range loom.SpeccedThemes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func configurePane(pane *loom.Pane) {
	pane.Resizeable = true
	pane.MaxCols = 0               // Use the terminal width; Loom's default cap is 50 columns.
	pane.DisableDefaultQuit = true // q remains available as a file-list filter.
	pane.Background = loom.NewAstraBackground()
}
