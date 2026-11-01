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
			if strings.Contains(line, "os.Getenv(\"LOOM_SESSION_") {
				start := strings.Index(line, "os.Getenv(\"LOOM_SESSION_") + len("os.Getenv(\"")
				end := strings.Index(line[start:], "\"")
				if end == -1 {
					continue
				}
				envVar := line[start : start+end]

				if !allowList[envVar] {
					t.Errorf("%s:%d: reads %s which is not in the allow list", file, i+1, envVar)
				}
			}
		}
	}
}
