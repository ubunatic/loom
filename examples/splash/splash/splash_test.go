// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package splash

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func stripANSI(s string) string {
	var b strings.Builder
	esc := false
	csi := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			esc = true
		case esc && r == '[':
			csi = true
			esc = false
		case csi && r >= 0x40 && r <= 0x7e:
			csi = false
		case esc:
			esc = false
		case !csi:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func TestSplashShowOnceOutput(t *testing.T) {
	var buf bytes.Buffer
	err := Execute(context.Background(), []string{"-w", "60", "-H", "12"}, &buf)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	out := stripANSI(buf.String())
	if !strings.Contains(out, "harnez usage") {
		t.Errorf("output missing title: %q", out)
	}
	if !strings.Contains(out, "fetching claude...") {
		t.Errorf("output missing step text: %q", out)
	}
	if !strings.Contains(out, "● mic  ✳ claude  ֍ codex  Λ agy") {
		t.Errorf("output missing status pills: %q", out)
	}
	if !strings.Contains(out, "Esc to skip") {
		t.Errorf("output missing footer: %q", out)
	}
}
