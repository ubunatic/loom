// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
)

func TestYAMLRootFormsAndOrder(t *testing.T) {
	for _, form := range []string{"pane", "view", "views"} {
		t.Run(form, func(t *testing.T) {
			header := form + ":\n"
			if form == "views" {
				header += "  - name: main\n"
			}
			indent := "  "
			if form == "views" {
				indent = "    "
			}
			body := "order: [second, first]\nelements:\n  first:\n    type: view\n    static: FIRST\n  second:\n    type: view\n    static: SECOND\n"
			input := header + indent + strings.ReplaceAll(strings.TrimSuffix(body, "\n"), "\n", "\n"+indent) + "\n"
			for range 20 {
				if err := ValidateYAML(strings.NewReader(input)); err != nil {
					t.Fatal(err)
				}
				w, _, err := BuildWidget(strings.NewReader(input))
				if err != nil {
					t.Fatal(err)
				}
				rows := Render(w, 20, 2)
				if !strings.HasPrefix(rows[0], "SECOND") || !strings.HasPrefix(rows[1], "FIRST") {
					t.Fatalf("order lost: %q", rows)
				}
			}
			// Absence of explicit order has a documented, stable lexical fallback.
			input = strings.Replace(input, indent+"order: [second, first]\n", "", 1)
			w, _, err := BuildWidget(strings.NewReader(input))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(Render(w, 20, 2)[0], "FIRST") {
				t.Fatal("legacy map order not lexical")
			}
		})
	}
}

func TestYAMLRootSelection(t *testing.T) {
	input := "views:\n  - name: other\n    elements: {a: {type: view, static: OTHER}}\n  - name: last\n    elements: {a: {type: view, static: LAST}}\n"
	w, cfg, err := BuildWidget(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	root, err := cfg.RootView()
	if err != nil {
		t.Fatal(err)
	}
	if w.(*Router).Current() != root.Name || root.Name != "other" {
		t.Fatal("fallback root disagrees")
	}
	w, _, err = BuildWidget(strings.NewReader("app: {root: last}\n" + input))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(Render(w, 10, 1)[0], "LAST") {
		t.Fatal("explicit root ignored")
	}
}

func TestYAMLContractErrors(t *testing.T) {
	for _, tc := range []struct{ name, input, message string }{
		{"missing", "app: {}", "exactly one"},
		{"ambiguous", "view: {}\npane: {}", "exactly one"},
		{"ambiguous empty views", "view: {}\nviews: []", "exactly one"},
		{"unknown root", "app: {root: missing}\nviews: [{name: main}]", "app.root"},
		{"duplicate views", "views:\n  - name: main\n    elements: {a: {type: view}}\n  - name: main\n    elements: {a: {type: view}}", "duplicate"},
		{"missing view name", "views: [{elements: {a: {type: view}}}]", "name"},
		{"duplicate element", "view:\n  elements:\n    a: {type: view}\n    a: {type: view}", "already defined"},
		{"unknown order", "view: {order: [missing], elements: {a: {type: view}}}", "unknown element"},
		{"duplicate order", "view: {order: [a, a], elements: {a: {type: view}}}", "duplicate"},
		{"incomplete order", "view: {order: [a], elements: {a: {type: view}, b: {type: view}}}", "every element"},
		{"grid reference", "view:\n  grid: |\n    +---+\n    |Z  |\n    +---+\n  elements: {a: {type: view}}", "unknown element"},
		{"negative dimension", "app: {max_width: -1}\nview: {elements: {a: {type: view}}}", "negative"},
		{"trailing document", "view: {}\n---\nview: {}", "one YAML document"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, buildErr := BuildWidget(strings.NewReader(tc.input))
			validateErr := ValidateYAML(strings.NewReader(tc.input))
			if buildErr == nil || validateErr == nil || buildErr.Error() != validateErr.Error() || !strings.Contains(buildErr.Error(), tc.message) {
				t.Fatalf("build=%v validate=%v want %q", buildErr, validateErr, tc.message)
			}
		})
	}
}
