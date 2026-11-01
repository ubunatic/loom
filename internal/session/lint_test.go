package session_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvVarReads(t *testing.T) {
	allowList := map[string]bool{
		"LOOM_SESSION_SERVER_NAME": true, // Consumed and unset during setup
	}

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			for _, fn := range []string{"os.Getenv(\"", "os.LookupEnv(\""} {
				idx := strings.Index(line, fn+"LOOM_SESSION_")
				if idx < 0 {
					continue
				}
				start := idx + len(fn)
				end := strings.Index(line[start:], "\"")
				if end == -1 {
					continue
				}
				envVar := line[start : start+end]
				if !allowList[envVar] {
					t.Errorf("%s:%d: %s reads %s which is not in the allow list", file, i+1, strings.TrimSuffix(fn, "(\""), envVar)
				}
			}
		}
	}
}
