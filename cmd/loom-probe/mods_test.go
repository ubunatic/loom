package main

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func writeState(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestModMaskNames(t *testing.T) {
	tests := []struct {
		mask modMask
		want []string
	}{
		{0x00, nil},
		{0x01 | 0x02, []string{"lctrl", "rctrl"}},
		{0xFF, []string{"lctrl", "rctrl", "lalt", "ralt", "lsuper", "rsuper", "lshift", "rshift"}},
	}
	for bit, name := range modNames {
		tests = append(tests, struct {
			mask modMask
			want []string
		}{modMask(1 << bit), []string{name}})
	}
	for _, tt := range tests {
		if got := tt.mask.Names(); !slices.Equal(got, tt.want) {
			t.Errorf("modMask(0x%02x).Names() = %v, want %v", uint8(tt.mask), got, tt.want)
		}
	}
}

func TestReadModMaskMissingFile(t *testing.T) {
	_, err := readModMask(filepath.Join(t.TempDir(), "absent"))
	if !errors.Is(err, errModsUnavailable) {
		t.Fatalf("readModMask(missing) error = %v, want errModsUnavailable", err)
	}
}

func TestReadModMaskEmptyFile(t *testing.T) {
	path := writeState(t, t.TempDir(), "modifiers", nil)
	m, err := readModMask(path)
	if err == nil {
		t.Fatalf("readModMask(empty) = 0x%02x, want error", uint8(m))
	}
	if errors.Is(err, errModsUnavailable) {
		t.Fatalf("readModMask(empty) error = %v, want empty-file error", err)
	}
}

func TestReadModMaskUsesFirstByte(t *testing.T) {
	path := writeState(t, t.TempDir(), "modifiers", []byte{0x41, 0xFF, 0x00})
	m, err := readModMask(path)
	if err != nil || m != 0x41 {
		t.Fatalf("readModMask() = (0x%02x, %v), want (0x41, nil)", uint8(m), err)
	}
	if got := m.Names(); !slices.Equal(got, []string{"lctrl", "lshift"}) {
		t.Fatalf("Names() = %v, want [lctrl lshift]", got)
	}
}

func TestReadModMaskRereadsReplacedFile(t *testing.T) {
	dir := t.TempDir()
	path := writeState(t, dir, "modifiers", []byte{0x00})
	if m, err := readModMask(path); err != nil || m != 0 {
		t.Fatalf("first read = (0x%02x, %v), want (0x00, nil)", uint8(m), err)
	}
	tmp := writeState(t, dir, "modifiers.tmp", []byte{0x02})
	if err := os.Rename(tmp, path); err != nil {
		t.Fatal(err)
	}
	if m, err := readModMask(path); err != nil || m != 0x02 {
		t.Fatalf("read after rename = (0x%02x, %v), want (0x02, nil)", uint8(m), err)
	}
}

func TestFormatModsLine(t *testing.T) {
	got := formatModsLine([]byte("1"), 0x01, nil)
	if !strings.HasSuffix(got, "mask=0x01 lctrl") || !strings.Contains(got, `"1"`) {
		t.Errorf("formatModsLine(mask) = %q", got)
	}
	got = formatModsLine([]byte("1"), 0, nil)
	if !strings.HasSuffix(got, "mask=0x00 -") {
		t.Errorf("formatModsLine(zero) = %q", got)
	}
	got = formatModsLine([]byte("1"), 0, errModsUnavailable)
	if !strings.HasSuffix(got, "mask=n/a") {
		t.Errorf("formatModsLine(err) = %q", got)
	}
}
