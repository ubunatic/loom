package loom

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestResizeModeOrderFromSpec(t *testing.T) {
	var raw struct {
		ModeOrder []string `yaml:"mode_order"`
	}
	if err := yaml.Unmarshal(resizeSpecYAML, &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw.ModeOrder) == 0 || !reflect.DeepEqual(raw.ModeOrder, SpeccedResizeModeIDs) {
		t.Fatalf("presentation order = %v, YAML = %v", SpeccedResizeModeIDs, raw.ModeOrder)
	}
	if err := SpeccedResizeModes.validateModeOrder(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		order []string
	}{
		{"missing", nil},
		{"incomplete", raw.ModeOrder[:len(raw.ModeOrder)-1]},
		{"duplicate", append([]string{raw.ModeOrder[1]}, raw.ModeOrder[1:]...)},
		{"unknown", append([]string{"unknown-mode"}, raw.ModeOrder[1:]...)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			modes := SpeccedResizeModes
			modes.ModeOrder = tc.order
			if err := modes.validateModeOrder(); err == nil {
				t.Fatalf("accepted invalid mode order: %v", tc.order)
			}
		})
	}
}
