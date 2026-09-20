package loom

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func popupTopRow(t *testing.T, title string, width int) string {
	t.Helper()
	p := NewPopup(title, NewView(nil))
	p.Width = width
	p.Height = 3
	c := NewCanvas(width, 3)
	p.Draw(c, c.Bounds())
	return strings.TrimSuffix(c.Row(0), "\x1b[0m")
}

func TestPopupUTF8TitleTruncation(t *testing.T) {
	title := "📖📖📖📖"
	row := popupTopRow(t, title, 8)

	if !utf8.ValidString(row) {
		t.Fatalf("popup title row is not valid UTF-8: %q", row)
	}
	if got := StringWidth(row); got != 8 {
		t.Fatalf("popup title row display width = %d, want 8: %q", got, row)
	}
	if strings.Contains(row, title) {
		t.Fatalf("popup title was not truncated: %q", row)
	}
	for _, cell := range strings.Fields(row) {
		if !utf8.ValidString(cell) {
			t.Errorf("rendered field is not valid UTF-8: %q", cell)
		}
	}
}

func TestPopupUTF8TitleFitsWithoutTruncation(t *testing.T) {
	title := "📖é界"
	row := popupTopRow(t, title, 10)

	if StringWidth(title)+2 > 8 {
		t.Fatalf("test title does not exceed the intended interior budget")
	}
	if !strings.Contains(row, " "+title+" ") {
		t.Fatalf("popup title was truncated despite fitting: %q", row)
	}
	if got := StringWidth(row); got != 10 {
		t.Fatalf("popup title row display width = %d, want 10: %q", got, row)
	}
}

func TestPopupASCIITitleRenderingUnchanged(t *testing.T) {
	got := popupTopRow(t, "Title", 12)
	want := "┌ Title ───┐"
	if got != want {
		t.Fatalf("ASCII popup title row = %q, want %q", got, want)
	}
}
