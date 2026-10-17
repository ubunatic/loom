package ptytest

import (
	"strings"
	"testing"
	"time"
)

// Wait must not return before the reader has drained the child's last
// bytes; tests assert on Raw right after it (issue 329).
func TestWaitDrainsOutputAfterExit(t *testing.T) {
	for range 20 {
		s := Start(t, 80, 24, "sh", "-c", "head -c 300000 /dev/zero | tr '\\0' a; printf END")
		if err := s.Wait(5 * time.Second); err != nil {
			t.Fatalf("child exit: %v", err)
		}
		if raw := string(s.Raw()); !strings.HasSuffix(raw, "END") {
			t.Fatalf("output not drained: %d bytes, tail %q", len(raw), raw[max(0, len(raw)-20):])
		}
	}
}
