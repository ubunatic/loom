// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom/layout"
	"codeberg.org/ubunatic/loom/measure"
)

func TestPillDimensions(t *testing.T) {
	pills := []ProviderPill{
		{Symbol: "●", Name: "mic", State: ProviderDone},
		{Symbol: "✳", Name: "claude", State: ProviderFetching},
		{Symbol: "֍", Name: "codex", State: ProviderPending},
		{Symbol: "Λ", Name: "agy", State: ProviderPending},
	}

	cluster := NewPillCluster(pills...)

	// "● mic" = 1+1+3 = 5
	// "✳ claude" = 1+1+6 = 8
	// "֍ codex" = 1+1+5 = 7
	// "Λ agy" = 1+1+3 = 5
	// gaps = 3 * 2 = 6
	// Total width = 5 + 8 + 7 + 5 + 6 = 31
	wantWidth := 31
	if got := cluster.ContentWidth(); got != wantWidth {
		t.Fatalf("ContentWidth = %d, want %d", got, wantWidth)
	}

	// Format text check
	formatted := FormatPillCluster(pills, 2, false)
	wantText := "● mic  ✳ claude  ֍ codex  Λ agy"
	if formatted != wantText {
		t.Fatalf("formatted = %q, want %q", formatted, wantText)
	}
	if w := measure.StringWidth(formatted); w != wantWidth {
		t.Fatalf("StringWidth(formatted) = %d, want %d", w, wantWidth)
	}
}

func TestPillClusterDraw(t *testing.T) {
	pills := []ProviderPill{
		{Symbol: "●", Name: "mic", State: ProviderDone},
		{Symbol: "✳", Name: "claude", State: ProviderFetching},
	}
	cluster := NewPillCluster(pills...)
	cluster.Align = layout.AlignStart

	c := NewCanvas(30, 2)
	cluster.Draw(c, c.Bounds())

	row := c.Row(0)
	if !strings.Contains(row, "● mic") || !strings.Contains(row, "✳ claude") {
		t.Fatalf("row output missing pill text: %q", row)
	}
}

func TestProviderStateStyles(t *testing.T) {
	if s := StyleForProviderState(ProviderDone); s.FG != colorGreen {
		t.Errorf("ProviderDone style.FG = %v, want %v", s.FG, colorGreen)
	}
	if s := StyleForProviderState(ProviderFetching); s.FG != colorYellow {
		t.Errorf("ProviderFetching style.FG = %v, want %v", s.FG, colorYellow)
	}
	if s := StyleForProviderState(ProviderFailed); s.FG != colorRed {
		t.Errorf("ProviderFailed style.FG = %v, want %v", s.FG, colorRed)
	}
	if s := StyleForProviderState(ProviderPending); !s.Dim {
		t.Errorf("ProviderPending style.Dim = false, want true")
	}
}
