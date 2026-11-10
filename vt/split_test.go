package vt

import (
	"strings"
	"testing"
)

func TestB9Replay(t *testing.T) {
	v := NewVT(10, 2)
	v.Write([]byte("e"))
	v.Write([]byte("\u0301"))
	cells := v.Screen()
	if !strings.HasPrefix(cells[0], "e\u0301") {
		t.Errorf("Expected e\\u0301, got %q", cells[0])
	}
}

func TestZWJSequenceSplit(t *testing.T) {
	v := NewVT(10, 2)
	v.Write([]byte("👨"))
	v.Write([]byte("\u200D"))
	v.Write([]byte("🌾"))
	cells := v.Screen()
	if !strings.HasPrefix(cells[0], "👨\u200D🌾") {
		t.Errorf("Expected farmer, got %q", cells[0])
	}
}

func TestVariationSelectorSplit(t *testing.T) {
	v := NewVT(10, 2)
	v.Write([]byte("❤"))
	v.Write([]byte("\uFE0F"))
	cells := v.Screen()
	if !strings.HasPrefix(cells[0], "❤️") {
		t.Errorf("Expected heart with variation selector, got %q", cells[0])
	}
}

func TestWrapCombine(t *testing.T) {
	v := NewVT(2, 2)
	v.Write([]byte("a"))
	v.Write([]byte("e"))
	v.Write([]byte("\u0301"))
	cells := v.Screen()
	if !strings.HasPrefix(cells[0], "ae\u0301") {
		t.Errorf("Expected ae\\u0301 at 1st row, got %q", cells[0])
	}
}
