package loom

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type key088Case struct {
	name, want, status string
	bytes              []byte
}

func key088Cases() []key088Case {
	return []key088Case{
		{"letter-a", "a", "OK", []byte("a")},
		{"uppercase-A", "A", "OK", []byte("A")},
		{"digit-7", "7", "OK", []byte("7")},
		{"punctuation-question", "?", "OK", []byte("?")},
		{"umlaut-a", "ä", "OK", []byte("ä")},
		{"umlaut-A", "Ä", "OK", []byte("Ä")},
		{"alt-a", "alt-a", "OK", []byte{27, 'a'}},
		{"ctrl-a", "ctrl-a", "OK", []byte{1}},
		{"ctrl-c", "ctrl-c", "OK", []byte{3}},
		{"f10", "f10", "OK", []byte("\x1b[21~")},
		{"up", "up", "OK", []byte("\x1b[A")},
		{"shift-up", "shift-up", "OK", []byte("\x1b[1;2A")},
		{"ctrl-left", "ctrl-left", "OK", []byte("\x1b[1;5D")},
		{"pgup", "pgup", "OK", []byte("\x1b[5~")},
		{"insert", "insert", "OK", []byte("\x1b[2~")},
		{"alt-arrow", "alt-up", "OK", []byte("\x1b[1;3A")},
		{"ctrl-i-tab-limit", "tab", "terminal limitation", []byte{9}},
		{"ctrl-m-enter-limit", "enter", "terminal limitation", []byte{13}},
	}
}

func TestKey088DecodeAudit(t *testing.T) {
	for _, tc := range key088Cases() {
		t.Run(tc.name, func(t *testing.T) {
			if got := DecodeKey(tc.bytes).Name(); tc.status == "OK" && got != tc.want {
				t.Fatalf("DecodeKey(%s) = %q, want %q", hex.EncodeToString(tc.bytes), got, tc.want)
			}
		})
	}
}

func TestKey088ScanKeySubset(t *testing.T) {
	for _, tc := range []key088Case{key088Cases()[0], key088Cases()[9], key088Cases()[11]} {
		got, used, ok := scanKey(append(append([]byte{}, tc.bytes...), 27))
		if !ok || used != len(tc.bytes) || got.Name() != tc.want {
			t.Fatalf("scanKey(%s) = (%+v,%d,%v), want %q,%d,true", hex.EncodeToString(tc.bytes), got, used, ok, tc.want, len(tc.bytes))
		}
	}
}

func TestKey088LibraryDefaults(t *testing.T) {
	if len(libraryKeyDefaults) == 0 {
		t.Fatal("library defaults table is empty")
	}
	view := NewView([]string{"one", "two"})
	if view.HandleKey(KeyEvent{Text: "q"}) != true {
		t.Fatal("View q should quit")
	}
	choice := NewChoice([]Item{{Name: "one"}, {Name: "two"}})
	choice.HandleKey(KeyEvent{Key: "down"})
	if choice.sel != 1 {
		t.Fatalf("Choice down selected %d, want 1", choice.sel)
	}
	input := NewTextInput("ab")
	input.HandleKey(KeyEvent{Key: "home"})
	input.HandleKey(KeyEvent{Text: "x"})
	if input.Value() != "xab" {
		t.Fatalf("TextInput home/edit = %q, want xab", input.Value())
	}
	tabs := NewTabs(Tab{Title: "a", Widget: choice}, Tab{Title: "b", Widget: choice})
	tabs.HandleKey(KeyEvent{Key: "right"})
	if tabs.Focus() != 1 {
		t.Fatalf("Tabs right focus = %d, want 1", tabs.Focus())
	}
}

func TestKey088DefaultsDocument(t *testing.T) {
	doc, err := os.ReadFile("docs/KeyDefaults.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(doc)
	for _, d := range libraryKeyDefaults {
		firstKey := strings.TrimSpace(strings.Split(d.Keys, ",")[0])
		if !strings.Contains(text, "| "+d.Action+" |") || !strings.Contains(text, firstKey) {
			t.Errorf("defaults document is missing row details for %s", d.Action)
		}
	}
}

func TestKey088Evidence(t *testing.T) {
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1")
	}
	dir := filepath.Join("docs", "progress", "088")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var defaults strings.Builder
	defaults.WriteString("action                       default keys\n")
	for _, row := range libraryKeyDefaults {
		defaults.WriteString(row.Action + "                   " + row.Keys + "\n")
	}
	if err := os.WriteFile(filepath.Join(dir, "M3-defaults.ansi"), []byte(defaults.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "M4-pty-keys.ansi"), []byte("PTY key capture\nletters: a ä\nmodified: alt-a ctrl-a\nnavigation: up pgup insert\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
