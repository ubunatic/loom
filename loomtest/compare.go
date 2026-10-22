// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomtest

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
)

var updateReference = flag.Bool("update", false, "update ANSI references from rendered widgets")

// CompareOptions controls which cell fields CompareWidget checks.
type CompareOptions struct {
	IgnoreColour bool
}

// CompareWidget calls render and compares its ANSI cells to reference. Run the
// test binary with -update to deliberately rewrite the reference.
func CompareWidget(t testing.TB, reference string, width, height int, render func() string, options ...CompareOptions) {
	t.Helper()
	got := render()
	if *updateReference {
		if err := os.WriteFile(reference, []byte(got), 0o644); err != nil {
			t.Fatalf("write reference %s: %v", reference, err)
		}
		return
	}
	want, err := os.ReadFile(reference)
	if err != nil {
		t.Fatalf("read reference %s: %v", reference, err)
	}
	if string(want) == got {
		return
	}
	option := CompareOptions{}
	if len(options) > 0 {
		option = options[0]
	}
	report := compareANSIReport(string(want), got, width, height, option.IgnoreColour)
	if strings.Contains(report, "differences: 0\n") {
		return
	}
	t.Errorf("reference %s differs; run go test with -update to accept it\n%s", reference, report)
}

// CompareANSI is the original name for CompareWidget.
//
// Deprecated: use CompareWidget.
func CompareANSI(t testing.TB, reference string, width, height int, render func() string, options ...CompareOptions) {
	CompareWidget(t, reference, width, height, render, options...)
}

func compareANSIReport(wantText, gotText string, width, height int, ignoreColour bool) string {
	want := parseComparisonScreen(wantText, width, height)
	got := parseComparisonScreen(gotText, width, height)
	var report strings.Builder
	differences := 0
	for y := 0; y < max(want.Rows, got.Rows); y++ {
		for x := 0; x < max(want.Cols, got.Cols); x++ {
			w, g := want.Cell(x, y), got.Cell(x, y)
			wfg, wbg := w.Style.Effective()
			gfg, gbg := g.Style.Effective()
			if w.Rune == g.Rune && (ignoreColour || (wfg == gfg && wbg == gbg)) {
				continue
			}
			differences++
			fmt.Fprintf(&report, "row=%d col=%d want={glyph=%q fg=%s bg=%s} got={glyph=%q fg=%s bg=%s}\n", y+1, x+1, string(w.Rune), formatColor(wfg), formatColor(wbg), string(g.Rune), formatColor(gfg), formatColor(gbg))
		}
	}
	fmt.Fprintf(&report, "differences: %d\n", differences)
	return report.String()
}

func parseComparisonScreen(text string, width, height int) *VT {
	screen := NewVT(width, height)
	// Reference files use LF as a row separator. A terminal LF advances the
	// row without returning to column zero, so normalize row breaks to CRLF
	// before feeding the terminal emulator.
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\n", "\r\n")
	screen.Write([]byte(text)) //nolint:errcheck
	return screen
}

func formatColor(color Color) string {
	switch color.Type {
	case ColorDefault:
		return "default"
	case ColorIndexed:
		return fmt.Sprintf("index:%d", color.Index)
	case ColorDirect:
		return fmt.Sprintf("#%02x%02x%02x", color.R, color.G, color.B)
	default:
		return "default"
	}
}
