// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package gallery

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/spec"
	"gopkg.in/yaml.v3"
)

func TestDemosMatchCatalogAndRender(t *testing.T) {
	var catalog struct {
		Widgets []struct {
			Name string `yaml:"name"`
		} `yaml:"widgets"`
	}
	data, err := spec.WidgetsYAML()
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		t.Fatal(err)
	}
	listed := map[string]bool{}
	for _, entry := range catalog.Widgets {
		listed[entry.Name] = true
	}
	for _, name := range Names() {
		qualified := "loom." + name
		if name == "Media" {
			qualified = "media.Widget"
		}
		if !listed[qualified] {
			t.Errorf("gallery name %q is missing from spec/widgets.yaml", qualified)
		}
		widget, err := New(name)
		if err != nil {
			t.Errorf("New(%q): %v", name, err)
			continue
		}
		rows := loom.Render(widget, 80, 24)
		if strings.TrimSpace(strings.Join(rows, "")) == "" {
			t.Errorf("%q rendered an empty frame", name)
		}
	}
}

func TestNewUnknownDemo(t *testing.T) {
	if _, err := New("Missing"); err == nil {
		t.Fatal("New accepted an unknown demo")
	}
}
