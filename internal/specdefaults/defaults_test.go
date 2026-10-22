// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package specdefaults

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Build with deliberately different embedded values, without importing loom's
// root package (which installs its own emoji configuration).
func TestChangedEmbeddedDefaults(t *testing.T) {
	root := filepath.Join("..", "..")
	dir := t.TempDir()
	for _, pattern := range []string{"go.mod", "go.sum", "collector/*.go", "graph/*.go", "measure/*.go", "spec/*.go", "spec/*.yaml"} {
		files, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			t.Fatal(err)
		}
		for _, source := range files {
			if strings.HasSuffix(source, "_test.go") {
				continue
			}
			rel, err := filepath.Rel(root, source)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			if rel == "spec/defaults.yaml" || rel == "spec/emoji.yaml" {
				data = changedDefaults(t, rel, data)
			}
			target := filepath.Join(dir, rel)
			if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(target, data, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	const probe = `package main
import (
 "strings"
 "time"
 "ubunatic.com/loom/collector"
 "ubunatic.com/loom/graph"
 "ubunatic.com/loom/measure"
)
func main() {
 retention, err := (collector.Spec{}).RetentionDuration()
 if err != nil || retention != 37*time.Minute || collector.DefaultRetention != retention { panic("retention did not follow YAML") }
 if graph.MaxWidth != 13 || graph.DefaultBackgroundANSI != "44" { panic("graph exports did not follow YAML") }
 if got := graph.RenderBar(100, graph.BarOptions{NoWrapper:true, ANSI:true}); !strings.HasPrefix(got, "\x1b[44m") || measure.StringWidth(got) != 13 { panic("graph rendering did not follow YAML") }
 checkEmoji := func() {
  s := measure.ActiveEmojiSpec()
  if s.VS16DefaultWidth != 3 || s.ZWJDefaultWidth != 3 || s.FlagDefaultWidth != 3 || s.VS16VTEMode != "default" || s.VS16NonVTEMode != "pad-1" { panic("emoji defaults did not follow YAML") }
  if measure.StringWidth("🇩🇪") != 3 { panic("emoji measurement did not follow YAML") }
  if measure.VTEMode("a\ufe0f") != "default" || measure.NonVTEMode("a\ufe0f") != "pad-1" { panic("emoji render modes did not follow YAML") }
 }
 checkEmoji()
 measure.LoadEmojiSpec(measure.EmojiSpec{})
 checkEmoji()
 if err := measure.LoadEmojiSpecYAML([]byte("{}")); err != nil { panic(err) }
 checkEmoji()
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(probe), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "LOOM_ZWJ=join")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("changed embedded defaults: %v\n%s", err, output)
	}
}

func changedDefaults(t *testing.T, path string, data []byte) []byte {
	t.Helper()
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if path == "spec/defaults.yaml" {
		doc["collector"].(map[string]any)["default_retention"] = "37m"
		graph := doc["graph"].(map[string]any)
		graph["default_background_ansi"] = "44"
		graph["bar"].(map[string]any)["max_width"] = 13
	} else {
		for _, key := range []string{"vs16_default_width", "zwj_default_width", "flag_default_width"} {
			doc[key] = 3
		}
		doc["vs16_vte_mode"] = "default"
		doc["vs16_non_vte_mode"] = "pad-1"
	}
	data, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
