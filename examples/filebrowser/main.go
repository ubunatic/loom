// Command filebrowser demonstrates a file list and live metadata in split panes.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"codeberg.org/ubunatic/loom"
)

func run(args []string) error {
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
	pane, err := loom.New(20)
	if err != nil {
		return err
	}
	defer pane.Close()
	configurePane(pane)
	pane.EnableMouseClicks()
	return pane.Run(app)
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
}
