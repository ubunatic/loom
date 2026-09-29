// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type widgetCatalogDocument struct {
	Widgets []struct {
		Name string `yaml:"name"`
		Docs string `yaml:"docs"`
	} `yaml:"widgets"`
	Exclusions []struct {
		Name string `yaml:"name"`
	} `yaml:"exclusions"`
}

func TestWidgetCatalogMatchesExportedWidgets(t *testing.T) {
	contents, err := os.ReadFile(filepath.Clean("spec/widgets.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var catalog widgetCatalogDocument
	if err := yaml.Unmarshal(contents, &catalog); err != nil {
		t.Fatal(err)
	}
	listed := make(map[string]bool, len(catalog.Widgets))
	for i, entry := range catalog.Widgets {
		if listed[entry.Name] {
			t.Errorf("widget catalog repeats %q", entry.Name)
		}
		if i > 0 && catalog.Widgets[i-1].Name >= entry.Name {
			t.Errorf("spec/widgets.yaml entries are not sorted by name: %q comes after %q", entry.Name, catalog.Widgets[i-1].Name)
		}
		if strings.Contains(entry.Docs, "§13") {
			t.Errorf("widget %q points docs to section 13 (the catalog section)", entry.Name)
		}
		listed[entry.Name] = true
	}
	// Named exclusions must still exist, so a stale entry cannot hide a rename.
	for _, exclusion := range catalog.Exclusions {
		name, ok := strings.CutPrefix(exclusion.Name, "loom.")
		if !ok {
			continue
		}
		if !rootPackageDeclares(t, name) {
			t.Errorf("widget catalog excludes %q, which is not declared in package loom", exclusion.Name)
		}
	}
	found := exportedWidgetTypes(t)
	for name := range found {
		if !listed[name] {
			t.Errorf("exported Widget implementation %q is missing from spec/widgets.yaml", name)
		}
	}
	for name := range listed {
		if !found[name] {
			t.Errorf("widget catalog entry %q does not name an exported Widget implementation", name)
		}
	}
}

func exportedWidgetTypes(t *testing.T) map[string]bool {
	t.Helper()
	command := exec.Command("go", "list", "-f", "{{.Dir}}|{{.ImportPath}}|{{.Name}}", "./...")
	output, err := command.Output()
	if err != nil {
		t.Fatalf("go list packages: %v", err)
	}
	found := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Split(line, "|")
		if len(fields) != 3 || strings.Contains(fields[1], "/examples/") || strings.Contains(fields[1], "/internal/") || strings.Contains(fields[1], "/cmd/") {
			continue
		}
		if fields[1] == "codeberg.org/ubunatic/loom" {
			fields[1] = "loom"
		}
		if err := collectPackageWidgets(fields[0], fields[2], found); err != nil {
			t.Fatalf("inspect %s: %v", fields[1], err)
		}
	}
	return found
}

func collectPackageWidgets(dir, packageName string, found map[string]bool) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return err
	}
	declared := map[string]bool{}
	methods := map[string]map[string]*ast.FuncDecl{}
	set := token.NewFileSet()
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(set, path, nil, 0)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			switch value := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range value.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if ok && typeSpec.Name.IsExported() {
						declared[typeSpec.Name.Name] = true
					}
				}
			case *ast.FuncDecl:
				if value.Recv == nil || value.Name == nil {
					continue
				}
				if value.Name.Name != "Draw" && value.Name.Name != "HandleKey" && value.Name.Name != "HandleMouse" {
					continue
				}
				if len(value.Recv.List) != 1 {
					continue
				}
				receiver := value.Recv.List[0].Type
				if pointer, ok := receiver.(*ast.StarExpr); ok {
					receiver = pointer.X
				}
				ident, ok := receiver.(*ast.Ident)
				if !ok {
					continue
				}
				if methods[ident.Name] == nil {
					methods[ident.Name] = map[string]*ast.FuncDecl{}
				}
				methods[ident.Name][value.Name.Name] = value
			}
		}
	}
	for name := range declared {
		methodSet := methods[name]
		if !implementsWidgetMethods(methodSet) {
			continue
		}
		qualified := packageName + "." + name
		if packageName == "loom" {
			qualified = "loom." + name
		}
		found[qualified] = true
	}
	return nil
}

func implementsWidgetMethods(methods map[string]*ast.FuncDecl) bool {
	// Embedded controls such as TextInput draw with an extra focus flag and
	// have no mouse handler; they are catalogued like full widgets.
	draw, key, mouse := methods["Draw"], methods["HandleKey"], methods["HandleMouse"]
	if draw == nil || key == nil {
		return false
	}
	params := draw.Type.Params.List
	if len(params) < 2 || len(params) > 3 || (draw.Type.Results != nil && len(draw.Type.Results.List) != 0) {
		return false
	}
	if len(params) == 3 && typeName(params[2].Type) != "bool" {
		return false
	}
	if typeName(draw.Type.Params.List[0].Type) != "Canvas" && typeName(draw.Type.Params.List[0].Type) != "loom.Canvas" {
		return false
	}
	if typeName(draw.Type.Params.List[1].Type) != "Rect" && typeName(draw.Type.Params.List[1].Type) != "loom.Rect" {
		return false
	}
	if mouse != nil && !handlerMatches(mouse, "MouseEvent", "loom.MouseEvent") {
		return false
	}
	return handlerMatches(key, "KeyEvent", "loom.KeyEvent")
}

func handlerMatches(method *ast.FuncDecl, event, qualifiedEvent string) bool {
	if len(method.Type.Params.List) != 1 || method.Type.Results == nil || len(method.Type.Results.List) != 1 {
		return false
	}
	parameter := typeName(method.Type.Params.List[0].Type)
	return (parameter == event || parameter == qualifiedEvent) && typeName(method.Type.Results.List[0].Type) == "bool"
}

func typeName(expression ast.Expr) string {
	if pointer, ok := expression.(*ast.StarExpr); ok {
		expression = pointer.X
	}
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.SelectorExpr:
		if pkg, ok := value.X.(*ast.Ident); ok {
			return pkg.Name + "." + value.Sel.Name
		}
	}
	return ""
}

func rootPackageDeclares(t *testing.T, name string) bool {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	set := token.NewFileSet()
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(set, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		if file.Scope.Lookup(name) != nil {
			return true
		}
	}
	return false
}
