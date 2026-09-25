package syntax

import "testing"

func TestCoordinateConversions(t *testing.T) {
	source := []byte("aé界\tZ")
	for _, tc := range []struct{ byteOffset, runeOffset int }{{0, 0}, {1, 1}, {3, 2}, {6, 3}, {7, 4}, {8, 5}, {99, 5}} {
		if got := ByteToRune(source, tc.byteOffset); got != tc.runeOffset {
			t.Errorf("ByteToRune(%d)=%d want %d", tc.byteOffset, got, tc.runeOffset)
		}
		if got := RuneToByte(source, tc.runeOffset); got != min(tc.byteOffset, len(source)) {
			t.Errorf("RuneToByte(%d)=%d", tc.runeOffset, got)
		}
	}
	if got := RuneToDisplayCol(source, 4); got != 8 {
		t.Fatalf("display column=%d want 8", got)
	}
	if RuneToByte(source, -1) != 0 || ByteToRune(source, -2) != 0 {
		t.Fatal("negative offsets must clamp to zero")
	}
}

func TestCoordinateConversionsUTF8BoundariesAndDisplay(t *testing.T) {
	source := []byte("Aé🙂\t界Z")
	for _, tc := range []struct{ byte, rune int }{{0, 0}, {1, 1}, {2, 1}, {3, 2}, {7, 3}, {8, 4}, {11, 5}, {12, 6}} {
		if got := ByteToRune(source, tc.byte); got != tc.rune {
			t.Errorf("ByteToRune(%d)=%d want %d", tc.byte, got, tc.rune)
		}
	}
	for _, tc := range []struct{ rune, byte, col int }{{0, 0, 0}, {1, 1, 1}, {2, 3, 2}, {3, 7, 4}, {4, 8, 8}, {5, 11, 10}, {6, 12, 11}, {99, 12, 11}} {
		if got := RuneToByte(source, tc.rune); got != tc.byte {
			t.Errorf("RuneToByte(%d)=%d want %d", tc.rune, got, tc.byte)
		}
		if got := RuneToDisplayCol(source, tc.rune); got != tc.col {
			t.Errorf("RuneToDisplayCol(%d)=%d want %d", tc.rune, got, tc.col)
		}
	}
}
