// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command validate-spec validates Loom YAML documents against their JSON
// Schemas without requiring a Python runtime.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

type specPair struct {
	document string
	schema   string
}

var specPairs = []specPair{
	{document: "spec/defaults.yaml", schema: "spec/schemas/defaults.schema.json"},
	{document: "spec/box.yaml", schema: "spec/schemas/box.schema.json"},
	{document: "spec/themes.yaml", schema: "spec/schemas/themes.schema.json"},
	{document: "spec/backgrounds.yaml", schema: "spec/schemas/backgrounds.schema.json"},
	{document: "spec/resize.yaml", schema: "spec/schemas/resize.schema.json"},
	{document: "examples/monitor/monitor/spec/watch.yaml", schema: "spec/schemas/watch.schema.json"},
	{document: "examples/monitor/monitor/spec/monitor.yaml", schema: "spec/schemas/monitor.schema.json"},
	{document: "testdata/fixtures/empty-shell.yaml", schema: "spec/schemas/monitor.schema.json"},
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fail(err)
	}
	for _, pair := range specPairs {
		if err := validatePair(root, pair); err != nil {
			fail(err)
		}
		fmt.Printf("validated %s (including negative controls)\n", pair.document)
	}
}

func validatePair(root string, pair specPair) error {
	schema, err := loadSchema(filepath.Join(root, pair.schema))
	if err != nil {
		return fmt.Errorf("%s: %w", pair.schema, err)
	}
	document, err := loadDocument(filepath.Join(root, pair.document))
	if err != nil {
		return fmt.Errorf("%s: %w", pair.document, err)
	}
	if err := schema.Validate(document); err != nil {
		return fmt.Errorf("%s: %w", pair.document, err)
	}

	if err := expectInvalid(schema, pair, document, "misspelled_property", func(invalid map[string]any) error {
		invalid["misspelled_property"] = true
		return nil
	}); err != nil {
		return err
	}

	if _, themesOK := document["themes"].(map[string]any); themesOK {
		for _, change := range []string{"invalid_rgb_short", "invalid_rgb_chars", "out_of_range_index", "invalid_color_name"} {
			if err := expectInvalid(schema, pair, document, change, func(invalid map[string]any) error {
				invalidThemes := invalid["themes"].(map[string]any)
				plain := invalidThemes["plain"].(map[string]any)
				switch change {
				case "invalid_rgb_short":
					plain["normal_fg"] = "#12345"
				case "invalid_rgb_chars":
					plain["normal_fg"] = "#gggggg"
				case "out_of_range_index":
					plain["normal_fg"] = 256
				case "invalid_color_name":
					plain["normal_fg"] = "notacolor"
				}
				return nil
			}); err != nil {
				return err
			}
		}
	}

	if _, modesOK := document["modes"].(map[string]any); modesOK {
		for _, change := range []string{"missing_required_mode", "invalid_default_type", "unknown_mode_field", "empty_title", "invalid_guard_n", "invalid_rate_step", "invalid_margin_rows", "invalid_min_percent"} {
			if err := expectInvalid(schema, pair, document, change, func(invalid map[string]any) error {
				invalidModes := invalid["modes"].(map[string]any)
				switch change {
				case "missing_required_mode":
					delete(invalidModes, "coalesce")
				case "invalid_default_type":
					coalesce := invalidModes["coalesce"].(map[string]any)
					coalesce["default"] = "not_a_bool"
				case "unknown_mode_field":
					coalesce := invalidModes["coalesce"].(map[string]any)
					coalesce["unknown_field"] = true
				case "empty_title":
					coalesce := invalidModes["coalesce"].(map[string]any)
					coalesce["title"] = ""
				case "invalid_guard_n":
					wg := invalidModes["width_guard"].(map[string]any)
					wg["guard_n"] = 0
				case "invalid_rate_step":
					ag := invalidModes["adaptive_guard"].(map[string]any)
					ag["rate_step"] = 0
				case "invalid_margin_rows":
					invalid["auto_fullscreen"].(map[string]any)["margin_rows"] = -1
				case "invalid_min_percent":
					invalid["auto_fullscreen"].(map[string]any)["min_percent"] = 101
				}
				return nil
			}); err != nil {
				return err
			}
		}
	}

	view, ok := document["view"].(map[string]any)
	if !ok {
		return nil
	}
	for _, change := range []string{"negative_width", "unknown_box_field", "missing_id", "ambiguous_root"} {
		if err := expectInvalid(schema, pair, document, change, func(invalid map[string]any) error {
			invalidView := invalid["view"].(map[string]any)
			boxes := invalidView["frame"].(map[string]any)["boxes"].([]any)
			box := boxes[0].(map[string]any)
			switch change {
			case "negative_width":
				box["width"] = -1
			case "unknown_box_field":
				box["widht"] = 31
			case "missing_id":
				delete(box, "id")
			case "ambiguous_root":
				pane, err := clone(invalidView)
				if err != nil {
					return err
				}
				invalid["pane"] = pane
			}
			return nil
		}); err != nil {
			return err
		}
	}

	dynamic, err := clone(document)
	if err != nil {
		return fmt.Errorf("%s: clone document: %w", pair.document, err)
	}
	dynamicView := dynamic["view"].(map[string]any)
	dynamicBox := dynamicView["frame"].(map[string]any)["boxes"].([]any)[0].(map[string]any)
	dynamicBox["dynamic"] = true
	delete(dynamicBox, "width")
	delete(dynamicBox, "height")
	if err := schema.Validate(dynamic); err != nil {
		return fmt.Errorf("%s: dynamic box: %w", pair.document, err)
	}

	frame := view["frame"].(map[string]any)
	box := frame["boxes"].([]any)[0].(map[string]any)
	_, rowsOK := box["rows"].(map[string]any)
	if !rowsOK {
		return nil
	}
	for _, change := range []string{
		"invalid_align", "zero_column_width", "negative_column_width", "negative_gap",
		"unknown_rows_field", "unknown_column_field", "non_string_value",
	} {
		if err := expectInvalid(schema, pair, document, change, func(invalid map[string]any) error {
			invalidView := invalid["view"].(map[string]any)
			invalidBox := invalidView["frame"].(map[string]any)["boxes"].([]any)[0].(map[string]any)
			invalidRows := invalidBox["rows"].(map[string]any)
			columns := invalidRows["columns"].([]any)
			column := columns[0].(map[string]any)
			switch change {
			case "invalid_align":
				column["align"] = "center"
			case "zero_column_width":
				column["width"] = 0
			case "negative_column_width":
				column["width"] = -1
			case "negative_gap":
				invalidRows["gap"] = -1
			case "unknown_rows_field":
				invalidRows["unknown_field"] = true
			case "unknown_column_field":
				column["unknown_field"] = true
			case "non_string_value":
				invalidRows["values"].([]any)[0].([]any)[0] = 123
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func loadSchema(path string) (*jsonschema.Resolved, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var schema jsonschema.Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("decode schema: %w", err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		return nil, fmt.Errorf("resolve schema: %w", err)
	}
	return resolved, nil
}

func loadDocument(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var node yaml.Node
	if err := decoder.Decode(&node); err != nil {
		return nil, fmt.Errorf("decode YAML: %w", err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("decode YAML: %w", err)
		}
		return nil, fmt.Errorf("expected one YAML document")
	}
	var value any
	if err := node.Decode(&value); err != nil {
		return nil, fmt.Errorf("decode YAML value: %w", err)
	}
	jsonData, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("convert YAML to JSON: %w", err)
	}
	var document map[string]any
	if err := json.Unmarshal(jsonData, &document); err != nil {
		return nil, fmt.Errorf("decode JSON value: %w", err)
	}
	return document, nil
}

func expectInvalid(schema *jsonschema.Resolved, pair specPair, document map[string]any, name string, mutate func(map[string]any) error) error {
	invalid, err := clone(document)
	if err != nil {
		return fmt.Errorf("%s: clone document for %s: %w", pair.document, name, err)
	}
	if err := mutate(invalid); err != nil {
		return fmt.Errorf("%s: prepare %s: %w", pair.document, name, err)
	}
	if err := schema.Validate(invalid); err == nil {
		return fmt.Errorf("%s: accepts %s", pair.schema, name)
	}
	return nil
}

func clone(value map[string]any) (map[string]any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var copy map[string]any
	if err := json.Unmarshal(data, &copy); err != nil {
		return nil, err
	}
	return copy, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
