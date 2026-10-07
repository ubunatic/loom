// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

//go:embed spec/schemas/editor.schema.json
var editorSchemaJSON []byte

var (
	editorSchemaOnce sync.Once
	editorSchema     *jsonschema.Resolved
	editorSchemaErr  error
)

// EditorConfig holds preferences loaded from editor.yaml or spec defaults.
type EditorConfig struct {
	Theme     string `yaml:"theme" json:"theme"`
	MouseGrab bool   `yaml:"mousegrab" json:"mousegrab"`
	AltScreen bool   `yaml:"altscreen" json:"altscreen"`
}

// DefaultEditorConfigPath returns the XDG config path for editor.yaml (~/.config/loom/editor.yaml).
func DefaultEditorConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "loom", "editor.yaml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "loom", "editor.yaml")
	}
	return filepath.Join(home, ".config", "loom", "editor.yaml")
}

// LoadEditorConfig loads, parses, and validates editor preferences.
// If path is empty, DefaultEditorConfigPath() is checked. Missing default path returns spec defaults.
func LoadEditorConfig(path string) (EditorConfig, error) {
	defaults := EditorConfig{
		Theme:     SpeccedDefaults.Editor.Theme,
		MouseGrab: SpeccedDefaults.Editor.MouseGrab,
		AltScreen: SpeccedDefaults.Editor.AltScreen,
	}

	explicitPath := path != ""
	if path == "" {
		path = DefaultEditorConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && !explicitPath {
			return defaults, nil
		}
		return defaults, fmt.Errorf("editor config %s: %w", path, err)
	}

	if err := ValidateEditorConfigYAML(data); err != nil {
		return defaults, fmt.Errorf("editor config %s: %w", path, err)
	}

	cfg := defaults
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return defaults, fmt.Errorf("editor config %s: parse yaml: %w", path, err)
	}

	if cfg.Theme != "" && !ThemeExists(cfg.Theme) {
		return defaults, fmt.Errorf("editor config %s: unknown theme %q", path, cfg.Theme)
	}

	return cfg, nil
}

// ValidateEditorConfigYAML validates raw YAML bytes against spec/schemas/editor.schema.json.
func ValidateEditorConfigYAML(data []byte) error {
	editorSchemaOnce.Do(func() {
		var schema jsonschema.Schema
		if err := json.Unmarshal(editorSchemaJSON, &schema); err != nil {
			editorSchemaErr = fmt.Errorf("parse editor schema: %w", err)
			return
		}
		resolved, err := schema.Resolve(nil)
		if err != nil {
			editorSchemaErr = fmt.Errorf("resolve editor schema: %w", err)
			return
		}
		editorSchema = resolved
	})
	if editorSchemaErr != nil {
		return editorSchemaErr
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var node yaml.Node
	if err := decoder.Decode(&node); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	var rawValue any
	if err := node.Decode(&rawValue); err != nil {
		return fmt.Errorf("invalid YAML structure: %w", err)
	}

	jsonData, err := json.Marshal(rawValue)
	if err != nil {
		return fmt.Errorf("convert YAML to JSON: %w", err)
	}

	var document map[string]any
	if err := json.Unmarshal(jsonData, &document); err != nil {
		return fmt.Errorf("convert JSON to map: %w", err)
	}

	if err := editorSchema.Validate(document); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}
