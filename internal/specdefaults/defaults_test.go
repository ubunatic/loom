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

func TestChangedEmbeddedKeyDefaults(t *testing.T) {
	root := filepath.Join("..", "..")
	dir := t.TempDir()
	for _, pattern := range []string{
		"go.mod", "go.sum", "*.go", "ansi/*.go",
		"specdefaults/*.go",
		"collector/*.go", "graph/*.go", "layout/*.go", "measure/*.go", "media/*.go", "spec/*.go", "spec/*.yaml", "spec/schemas/*.json", "syntax/*.go", "terminal/*.go",
	} {
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
			if rel == "spec/defaults.yaml" {
				var doc map[string]any
				if err := yaml.Unmarshal(data, &doc); err != nil {
					t.Fatal(err)
				}
				tabs := doc["tabs"].(map[string]any)["keys"].(map[string]any)
				tabs["previous"] = "h"
				tabs["next"] = "l"

				view := doc["view"].(map[string]any)["bindings"].(map[string]any)
				view["quit"] = []any{"x"}

				split := doc["split"].(map[string]any)["bindings"].(map[string]any)
				split["shrink"] = []any{"-"}

				choice := doc["choice"].(map[string]any)["bindings"].(map[string]any)
				choice["down"] = []any{"d"}

				tree := doc["tree"].(map[string]any)["bindings"].(map[string]any)
				tree["expand_or_enter"] = []any{"e"}

				menu := doc["menu"].(map[string]any)["bindings"].(map[string]any)
				menu["open_bar"] = []any{"f12"}

				form := doc["form"].(map[string]any)["bindings"].(map[string]any)
				form["cancel"] = []any{"q"}

				settings := doc["settings"].(map[string]any)["bindings"].(map[string]any)
				settings["quit"] = []any{"x"}

				mediaDoc := doc["media"].(map[string]any)
				mediaDoc["bindings"] = map[string]any{
					"play":     []any{"p"},
					"zoom_in":  []any{"z"},
					"zoom_out": []any{"-", "_"},
					"pan":      []any{"←/→/↑/↓", "left", "right", "up", "down"},
					"reset":    []any{"0"},
				}
				mediaDoc["labels"] = map[string]any{
					"play":     "Start/Stop",
					"zoom_in":  "Zoom in",
					"zoom_out": "Zoom out",
					"pan":      "Pan",
					"reset":    "Reset",
				}
				mediaDoc["zoom_factor"] = 2.0
				mediaDoc["pan_step"] = 0.25
				mediaDoc["loading_bar_width"] = 16

				data, err = yaml.Marshal(doc)
				if err != nil {
					t.Fatal(err)
				}
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
	"image"
	"strings"
	"ubunatic.com/loom"
	"ubunatic.com/loom/media"
)
func main() {
	tabs := loom.DefaultTabsKeys()
	if tabs.Previous != "h" || tabs.Next != "l" {
		panic("tabs default keys did not follow YAML: got previous=" + tabs.Previous + " next=" + tabs.Next)
	}
	v := loom.NewView([]string{"one", "two"})
	if !v.ConsumeKey(loom.KeyEvent{Key: "x"}).Quit {
		panic("view quit key did not follow YAML")
	}
	if v.ConsumeKey(loom.KeyEvent{Key: "q"}).Quit {
		panic("view original quit key should not have triggered")
	}
	split := loom.NewSplit(nil, nil)
	origRatio := split.Ratio
	split.ConsumeKey(loom.KeyEvent{Key: "-"})
	if split.Ratio >= origRatio {
		panic("split shrink key did not follow YAML")
	}
	choice := loom.NewChoice([]loom.Item{{Name: "one"}, {Name: "two"}})
	choice.ConsumeKey(loom.KeyEvent{Key: "d"})
	item, ok := choice.Selected()
	if !ok || item.Name != "two" {
		panic("choice down key did not follow YAML")
	}
	tree := loom.NewTree([]*loom.TreeNode{{ID: "root", Label: "root", Children: []*loom.TreeNode{{ID: "child", Label: "child"}}}})
	tree.ConsumeKey(loom.KeyEvent{Key: "e"})
	if !tree.VisibleNodes()[0].Expanded {
		panic("tree expand key did not follow YAML")
	}
	menu := loom.NewMenuBar(loom.Menu{Title: "File", Items: []loom.MenuItem{{Label: "Exit"}}})
	menu.ConsumeKey(loom.KeyEvent{Key: "f12"})
	if !menu.Open {
		panic("menu open_bar key did not follow YAML")
	}
	formCancelled := false
	form := loom.NewForm(nil)
	form.OnCancel = func() { formCancelled = true }
	form.ConsumeKey(loom.KeyEvent{Key: "q"})
	if !formCancelled {
		panic("form cancel key did not follow YAML")
	}
	settings := loom.NewSettings(nil)
	if !settings.ConsumeKey(loom.KeyEvent{Key: "x"}).Quit {
		panic("settings quit key did not follow YAML")
	}
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	mediaWidget, err := media.NewImage(img, media.ModeHalfblock)
	if err != nil {
		panic(err)
	}
	if !mediaWidget.ConsumeKey(loom.KeyEvent{Key: "z"}).Consumed {
		panic("media zoom_in key did not follow YAML")
	}
	if mediaWidget.ConsumeKey(loom.KeyEvent{Key: "+"}).Consumed {
		panic("media original zoom_in key should not have triggered")
	}
	mediaCanvas := loom.NewCanvas(40, 10)
	mediaWidget.Draw(mediaCanvas, loom.Rect{W: 40, H: 10})
	foundHelp := false
	for y := 0; y < 10; y++ {
		if strings.Contains(mediaCanvas.Row(y), "Start/Stop") {
			foundHelp = true
			break
		}
	}
	if !foundHelp {
		panic("media help label did not follow YAML")
	}
	if loom.SpeccedDefaults.Media.ZoomFactor != 2.0 || loom.SpeccedDefaults.Media.PanStep != 0.25 || loom.SpeccedDefaults.Media.LoadingBarWidth != 16 {
		panic("media numeric defaults did not follow YAML")
	}
	foundZoom := false
	for y := 0; y < 10; y++ {
		if strings.Contains(mediaCanvas.Row(y), "2.00x") {
			foundZoom = true
			break
		}
	}
	if !foundZoom {
		panic("media zoom factor did not follow YAML")
	}
	if !mediaWidget.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollUp, X: 5, Y: 5}).Consumed {
		panic("media mouse scroll zoom did not trigger")
	}
	mediaCanvas = loom.NewCanvas(40, 10)
	mediaWidget.Draw(mediaCanvas, loom.Rect{W: 40, H: 10})
	foundZoom = false
	for y := 0; y < 10; y++ {
		if strings.Contains(mediaCanvas.Row(y), "4.00x") {
			foundZoom = true
			break
		}
	}
	if !foundZoom {
		panic("media mouse scroll zoom factor did not follow YAML")
	}
}
`
	if err := os.MkdirAll(filepath.Join(dir, "probe"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "probe", "main.go"), []byte(probe), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", "./probe")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("changed embedded key defaults: %v\n%s", err, output)
	}
}
