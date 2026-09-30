// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/gallery"
	"codeberg.org/ubunatic/loom/spec"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type widgetCatalog struct {
	Widgets    []widgetEntry `yaml:"widgets"`
	Exclusions []struct {
		Name   string `yaml:"name"`
		Reason string `yaml:"reason"`
	} `yaml:"exclusions"`
}

type widgetEntry struct {
	Name         string   `yaml:"name"`
	Package      string   `yaml:"package"`
	Purpose      string   `yaml:"purpose"`
	Category     string   `yaml:"category"`
	Capabilities []string `yaml:"capabilities"`
	Example      string   `yaml:"example"`
	Source       string   `yaml:"source"`
	Docs         string   `yaml:"docs"`
	Ticket       string   `yaml:"ticket"`
}

func widgetsCommand() *cobra.Command {
	var show, list bool
	command := &cobra.Command{
		Use:   "widgets [name]",
		Short: "List library widgets, show usage, or run live demos",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeWidgetNames(toComplete), cobra.ShellCompDirectiveNoFileComp
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if show {
				return nil
			}
			return cobra.MaximumNArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if show {
				return showWidgetDemos(args)
			}
			catalog, err := readWidgetCatalog()
			if err != nil {
				return err
			}
			if len(args) == 0 || list {
				categories := []string{"input", "display", "layout", "infra"}
				for _, cat := range categories {
					for _, entry := range catalog.Widgets {
						if entry.Category == cat {
							fmt.Fprintf(cmd.OutOrStdout(), "%s [%s] — %s\n", entry.Name, entry.Category, entry.Purpose)
						}
					}
				}
				return nil
			}
			name := args[0]
			var matches []widgetEntry
			for _, entry := range catalog.Widgets {
				if name == entry.Name || name == strings.TrimPrefix(entry.Name, entry.Package+".") {
					matches = append(matches, entry)
				}
			}
			switch len(matches) {
			case 0:
				return fmt.Errorf("unknown widget %q", name)
			case 1:
				writeWidget(cmd.OutOrStdout(), matches[0])
				return nil
			}
			names := make([]string, len(matches))
			for i, entry := range matches {
				names[i] = entry.Name
			}
			return fmt.Errorf("widget %q is ambiguous: %s", name, strings.Join(names, ", "))
		},
	}
	command.Flags().BoolVar(&list, "list", false, "print the widget catalog and exit")
	command.Flags().BoolVar(&show, "show", false, "run live widget demos")
	_ = command.RegisterFlagCompletionFunc("show", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeWidgetNames(toComplete), cobra.ShellCompDirectiveNoFileComp
	})
	return command
}

func completeWidgetNames(prefix string) []string {
	catalog, err := readWidgetCatalog()
	if err != nil {
		return nil
	}
	available := make(map[string]bool)
	for _, name := range gallery.Names() {
		available[name] = true
	}
	var names []string
	for _, entry := range catalog.Widgets {
		short := strings.TrimPrefix(entry.Name, entry.Package+".")
		if !available[short] {
			continue
		}
		for _, name := range []string{short, entry.Name} {
			if strings.HasPrefix(name, prefix) {
				names = append(names, name)
			}
		}
	}
	sort.Strings(names)
	return names
}

var runWidgetPane = func(widget loom.Widget) error {
	pane, err := loom.New(24)
	if err != nil {
		return err
	}
	defer pane.Close()
	return pane.Run(widget)
}

func showWidgetDemos(names []string) error {
	if len(names) == 0 {
		return runWidgetPane(gallery.NewAll())
	}
	if len(names) == 1 {
		widget, err := gallery.New(names[0])
		if err != nil {
			return err
		}
		return runWidgetPane(widget)
	}
	tabs := make([]loom.Tab, 0, len(names))
	for _, name := range names {
		widget, err := gallery.New(name)
		if err != nil {
			return err
		}
		tabs = append(tabs, loom.Tab{Title: strings.TrimPrefix(name, "loom."), Widget: widget})
	}
	return runWidgetPane(loom.NewTabs(tabs...))
}

func readWidgetCatalog() (widgetCatalog, error) {
	data, err := spec.WidgetsYAML()
	if err != nil {
		return widgetCatalog{}, fmt.Errorf("read embedded widget catalog: %w", err)
	}
	var catalog widgetCatalog
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return widgetCatalog{}, fmt.Errorf("decode embedded widget catalog: %w", err)
	}
	return catalog, nil
}

func writeWidget(out io.Writer, entry widgetEntry) {
	fmt.Fprintf(out, "%s [%s] — %s\n", entry.Name, entry.Category, entry.Purpose)
	if len(entry.Capabilities) > 0 {
		fmt.Fprintf(out, "Capabilities: %s\n", strings.Join(entry.Capabilities, "; "))
	}
	fmt.Fprintln(out, "Example:")
	for _, line := range strings.Split(entry.Example, "\n") {
		fmt.Fprintf(out, "  %s\n", line)
	}
	fmt.Fprintf(out, "Source: %s\n", entry.Source)
	if entry.Docs != "" {
		fmt.Fprintf(out, "Docs: %s\n", entry.Docs)
	}
	if entry.Ticket != "" {
		fmt.Fprintf(out, "Ticket: %s\n", entry.Ticket)
	}
}
