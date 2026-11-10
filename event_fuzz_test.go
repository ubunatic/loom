package loom

import (
	"testing"
)

func FuzzInput(f *testing.F) {
	for _, seed := range []string{
		"a",
		"\x1b[<0;10;5M",
		"\x1b[M 00",
		"\x1b[200~",
		"\x1b[201~",
		"\x1b[",
		"\x1bO",
		"\x1b[27;5;9~",
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		DecodeMouse(b)

		rest := b
		for len(rest) > 0 {
			_, n, ok := ScanKey(rest)
			if n < 0 || n > len(rest) {
				t.Fatalf("n=%d out of bounds for len=%d", n, len(rest))
			}
			if ok && n == 0 {
				t.Fatalf("ok=true but n=0")
			}
			if n == 0 {
				break
			}
			rest = rest[n:]
		}
		bulkConsumed := len(b) - len(rest)

		buf := make([]byte, 0, len(b))
		chunkedConsumed := 0
		for i := 0; i < len(b); i++ {
			buf = append(buf, b[i])
			for len(buf) > 0 {
				_, n, _ := ScanKey(buf)
				if n == 0 {
					break
				}
				buf = buf[n:]
				chunkedConsumed += n
			}
		}

		if chunkedConsumed != bulkConsumed {
			t.Fatalf("chunked consumed %d, but bulk consumed %d", chunkedConsumed, bulkConsumed)
		}
	})
}
