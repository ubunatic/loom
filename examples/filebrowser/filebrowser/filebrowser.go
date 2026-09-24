// Package filebrowser demonstrates a file list and live metadata in split panes.
package filebrowser

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"codeberg.org/ubunatic/loom"
)

// Run parses args and runs the filebrowser example. It returns flag.ErrHelp
// when -h/--help was requested, matching the standard flag package
// convention so callers can treat that as a clean, non-error exit.
func Run(args []string) error {
	app, err := NewWidget(args)
	if err != nil {
		return err
	}
	pane, err := loom.New(1 << 16)
	if err != nil {
		return err
	}
	defer pane.Close()
	configurePane(pane)
	return pane.Run(app)
}

// NewWidget builds the filebrowser root widget from command-line arguments,
// without creating or running a Pane.
func NewWidget(args []string) (loom.Widget, error) {
	flags := flag.NewFlagSet("filebrowser", flag.ContinueOnError)
	themeName := flags.String("theme", "mc", "color theme")
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	theme, err := resolveTheme(*themeName)
	if err != nil {
		return nil, err
	}
	dir := "."
	if flags.NArg() > 0 {
		dir = flags.Arg(0)
	}
	return newBrowser(dir, *themeName, theme)
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
	pane.MaxCols = 0 // Use the terminal width; Loom's default cap is 50 columns.
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		pane.Background = loom.NewAstraBackground()
	}
}
