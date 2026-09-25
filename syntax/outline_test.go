package syntax

import (
	"reflect"
	"testing"
)

func TestGoOutlineAndFolds(t *testing.T) {
	src := `package main

type Config struct {
	Port int
}

func (c *Config) Validate() bool {
	return true
}

func main() {
	println("hello")
}
`
	engine := NewLexicalEngine("go")
	if err := engine.Parse([]byte(src)); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify Symbols
	syms := engine.Symbols()
	if len(syms) != 3 {
		t.Fatalf("expected 3 symbols, got %d: %+v", len(syms), syms)
	}

	if syms[0].Name != "Config" || syms[0].Kind != "type" || syms[0].Line != 2 {
		t.Errorf("sym[0] mismatch: %+v", syms[0])
	}
	if syms[1].Name != "(Config).Validate" || syms[1].Kind != "method" || syms[1].Line != 6 {
		t.Errorf("sym[1] mismatch: %+v", syms[1])
	}
	if syms[2].Name != "main" || syms[2].Kind != "function" || syms[2].Line != 10 {
		t.Errorf("sym[2] mismatch: %+v", syms[2])
	}

	// Verify Breadcrumb
	// Line 0 (package main) -> nil
	if b := engine.Breadcrumb(0, 0); len(b) != 0 {
		t.Errorf("expected empty breadcrumb at line 0, got %v", b)
	}
	// Line 3 (inside Config) -> Config
	if b := engine.Breadcrumb(3, 0); !reflect.DeepEqual(b, []string{"Config"}) {
		t.Errorf("expected [Config] at line 3, got %v", b)
	}
	// Line 7 (inside Validate) -> Config, Validate
	if b := engine.Breadcrumb(7, 0); !reflect.DeepEqual(b, []string{"Config", "Validate"}) {
		t.Errorf("expected [Config, Validate] at line 7, got %v", b)
	}
	// Line 11 (inside main) -> main
	if b := engine.Breadcrumb(11, 0); !reflect.DeepEqual(b, []string{"main"}) {
		t.Errorf("expected [main] at line 11, got %v", b)
	}

	// Verify Folds
	folds := engine.Folds()
	if len(folds) < 3 {
		t.Fatalf("expected at least 3 folds, got %d: %v", len(folds), folds)
	}
	// Config struct: lines 2 to 4
	if folds[0] != [2]int{2, 4} {
		t.Errorf("fold[0] = %v, want [2, 4]", folds[0])
	}
	// Validate method: lines 6 to 8
	if folds[1] != [2]int{6, 8} {
		t.Errorf("fold[1] = %v, want [6, 8]", folds[1])
	}
	// main func: lines 10 to 12
	if folds[2] != [2]int{10, 12} {
		t.Errorf("fold[2] = %v, want [10, 12]", folds[2])
	}
}

func TestMarkdownOutlineAndBreadcrumbs(t *testing.T) {
	src := `# Title
Some intro.

## Section 1
Details 1.

### Subsection 1.1
Deep dive.

## Section 2
Conclusion.
`
	engine := NewLexicalEngine("markdown")
	_ = engine.Parse([]byte(src))

	syms := engine.Symbols()
	if len(syms) != 4 {
		t.Fatalf("expected 4 symbols, got %d: %+v", len(syms), syms)
	}
	if syms[0].Name != "Title" || syms[1].Name != "Section 1" || syms[2].Name != "Subsection 1.1" || syms[3].Name != "Section 2" {
		t.Errorf("symbols mismatch: %+v", syms)
	}

	// Breadcrumb at line 7 (inside Subsection 1.1)
	b := engine.Breadcrumb(7, 0)
	wantB := []string{"Title", "Section 1", "Subsection 1.1"}
	if !reflect.DeepEqual(b, wantB) {
		t.Errorf("breadcrumb at line 7 = %v, want %v", b, wantB)
	}

	// Breadcrumb at line 10 (inside Section 2)
	b2 := engine.Breadcrumb(10, 0)
	wantB2 := []string{"Title", "Section 2"}
	if !reflect.DeepEqual(b2, wantB2) {
		t.Errorf("breadcrumb at line 10 = %v, want %v", b2, wantB2)
	}

	// Folds
	folds := engine.Folds()
	if len(folds) == 0 {
		t.Errorf("expected markdown heading folds, got none")
	}
}

func TestJSONOutlineAndFolds(t *testing.T) {
	src := `{
  "name": "loom",
  "config": {
    "port": 8080
  }
}`
	engine := NewLexicalEngine("json")
	_ = engine.Parse([]byte(src))

	syms := engine.Symbols()
	if len(syms) < 3 {
		t.Fatalf("expected at least 3 symbols, got %d: %+v", len(syms), syms)
	}
	if syms[0].Name != "name" || syms[1].Name != "config" || syms[2].Name != "port" {
		t.Errorf("symbols mismatch: %+v", syms)
	}

	folds := engine.Folds()
	if len(folds) < 2 {
		t.Fatalf("expected at least 2 folds, got %d: %v", len(folds), folds)
	}
	// Root object: lines 0 to 5
	if folds[0] != [2]int{0, 5} {
		t.Errorf("root fold = %v, want [0, 5]", folds[0])
	}
	// config object: lines 2 to 4
	if folds[1] != [2]int{2, 4} {
		t.Errorf("config fold = %v, want [2, 4]", folds[1])
	}
}

func TestYAMLOutlineAndFolds(t *testing.T) {
	src := `name: loom
server:
  host: localhost
  port: 8080
`
	engine := NewLexicalEngine("yaml")
	_ = engine.Parse([]byte(src))

	syms := engine.Symbols()
	if len(syms) != 4 {
		t.Fatalf("expected 4 symbols, got %d: %+v", len(syms), syms)
	}
	if syms[0].Name != "name" || syms[1].Name != "server" || syms[2].Name != "host" || syms[3].Name != "port" {
		t.Errorf("symbols mismatch: %+v", syms)
	}

	folds := engine.Folds()
	if len(folds) == 0 {
		t.Errorf("expected yaml indentation folds, got none")
	} else if folds[0] != [2]int{1, 3} {
		t.Errorf("server fold = %v, want [1, 3]", folds[0])
	}
}
