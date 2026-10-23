package loom

import (
	"fmt"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"ubunatic.com/loom/spec"
)

func TestTableAndChoiceChromeFromSpec(t *testing.T) {
	var raw struct {
		Table  map[string]any `yaml:"table"`
		Choice map[string]any `yaml:"choice"`
	}
	data, err := spec.DefaultsYAML()
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	defs := SpeccedDefaults.Table
	for key, got := range map[string]string{
		"sort_asc_glyph": defs.SortAscGlyph, "sort_desc_glyph": defs.SortDescGlyph,
		"controls_hint_fmt": defs.ControlsHintFmt,
	} {
		if got == "" || fmt.Sprint(raw.Table[key]) != got {
			t.Fatalf("table.%s = %q, YAML = %v", key, got, raw.Table[key])
		}
	}
	placeholder, _ := raw.Choice["default_placeholder"].(string)
	if placeholder == "" || DefaultPlaceholder != placeholder || NewChoice(nil).Placeholder != placeholder {
		t.Fatal("Choice placeholder did not load from spec")
	}
	table := NewTable([]Column{{Header: "NAME"}}, []Row{{Key: "item", Cells: []string{"item"}}})
	if table.Placeholder != placeholder {
		t.Fatal("Table placeholder did not load from spec")
	}
	table.SortCol = 0
	for _, desc := range []bool{false, true} {
		table.SortDesc = desc
		glyph := defs.SortAscGlyph
		if desc {
			glyph = defs.SortDescGlyph
		}
		if got := table.headerText(0); got != "NAME"+glyph {
			t.Fatalf("header = %q, want configured sort glyph", got)
		}
		rows := Render(table, 80, 5)
		want := fmt.Sprintf(defs.ControlsHintFmt, glyph, "NAME")
		if table.SearchBar().Controls != want || !strings.Contains(rows[4], want) {
			t.Fatalf("render omitted controls %q: %q", want, rows[4])
		}
	}
}
