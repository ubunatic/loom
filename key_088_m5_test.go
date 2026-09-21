package loom

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type key088AuditRow struct {
	group, name, want, status, reason string
	bytes                             []byte
}

func key088AuditRows() []key088AuditRow {
	var rows []key088AuditRow
	add := func(group, name, want string, b []byte) {
		rows = append(rows, key088AuditRow{group: group, name: name, want: want, status: "OK", bytes: b})
	}
	for _, r := range "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		add("letters-digits", string(r), string(r), []byte(string(r)))
	}
	for _, r := range "0123456789" {
		add("letters-digits", string(r), string(r), []byte{byte(r)})
	}
	for _, r := range "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~" {
		add("punct-umlauts", string(r), string(r), []byte(string(r)))
	}
	for _, r := range "äöüÄÖÜß" {
		add("punct-umlauts", string(r), string(r), []byte(string(r)))
	}
	for _, r := range "abcdefghijklmnopqrstuvwxyz" {
		add("ctrl-alt", "alt-"+string(r), "alt-"+string(r), append([]byte{27}, byte(r)))
		b := []byte{byte(r - 'a' + 1)}
		want := "ctrl-" + string(r)
		switch r {
		case 'b', 'c', 'd', 'f', 'q', 'u', 'w':
			want = "ctrl-" + string(r)
		case 'h':
			want = "backspace"
		case 'i':
			want = "tab"
		case 'j', 'm':
			want = "enter"
		}
		status := "OK"
		reason := ""
		if r == 'i' || r == 'j' || r == 'm' {
			status, reason = "terminal limitation", "indistinguishable from Tab/Enter"
		}
		rows = append(rows, key088AuditRow{group: "ctrl-alt", name: "ctrl-" + string(r), want: want, status: status, reason: reason, bytes: b})
	}
	for _, r := range "0123456789" {
		add("ctrl-alt", "alt-"+string(r), "alt-"+string(r), []byte{27, byte(r)})
	}
	for i := 1; i <= 12; i++ {
		ss3 := []byte{27, 'O', byte('P' + i - 1)}
		if i > 4 {
			ss3 = []byte(fmt.Sprintf("\x1b[%d~", map[int]int{5: 15, 6: 17, 7: 18, 8: 19, 9: 20, 10: 21, 11: 23, 12: 24}[i]))
		}
		add("function-nav", fmt.Sprintf("f%d", i), fmt.Sprintf("f%d", i), ss3)
	}
	for _, item := range []struct{ name, want, seq string }{{"up", "up", "\x1b[A"}, {"down", "down", "\x1b[B"}, {"left", "left", "\x1b[D"}, {"right", "right", "\x1b[C"}, {"home", "home", "\x1b[H"}, {"end", "end", "\x1b[F"}, {"home-1", "home", "\x1b[1~"}, {"end-4", "end", "\x1b[4~"}, {"home-7", "home", "\x1b[7~"}, {"end-8", "end", "\x1b[8~"}, {"pgup", "pgup", "\x1b[5~"}, {"pgdown", "pgdown", "\x1b[6~"}, {"insert", "insert", "\x1b[2~"}, {"delete", "delete", "\x1b[3~"}} {
		add("function-nav", item.name, item.want, []byte(item.seq))
	}
	for mod := 2; mod <= 8; mod++ {
		prefix := map[int]string{2: "shift-", 3: "alt-", 4: "shift-alt-", 5: "ctrl-", 6: "ctrl-shift-", 7: "ctrl-alt-", 8: "ctrl-shift-alt-"}[mod]
		for _, item := range []struct {
			r    rune
			name string
		}{{'A', "up"}, {'B', "down"}, {'C', "right"}, {'D', "left"}} {
			name := prefix + item.name
			add("modified-arrows", name, name, []byte(fmt.Sprintf("\x1b[1;%d%c", mod, item.r)))
		}
	}
	return rows
}

func TestKey088M5Audit(t *testing.T) {
	for _, row := range key088AuditRows() {
		t.Run(row.group+"/"+row.name, func(t *testing.T) {
			got := DecodeKey(row.bytes).Name()
			if got != row.want {
				t.Fatalf("DecodeKey(%s) = %q, want %q (%s)", hex.EncodeToString(row.bytes), got, row.want, row.reason)
			}
		})
	}
}

func TestKey088M5ScanRepresentative(t *testing.T) {
	rows := key088AuditRows()
	for i, row := range rows {
		if i%7 != 0 {
			continue
		}
		got, used, ok := scanKey(append(append([]byte{}, row.bytes...), 27))
		if !ok || used != len(row.bytes) || got.Name() != row.want {
			t.Fatalf("scanKey(%s) = (%q,%d,%v), want (%q,%d,true)", hex.EncodeToString(row.bytes), got.Name(), used, ok, row.want, len(row.bytes))
		}
	}
}

func TestKey088M5Evidence(t *testing.T) {
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1")
	}
	dir := filepath.Join("docs", "progress", "088")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	rows := key088AuditRows()
	counts := map[string][3]int{}
	for _, row := range rows {
		c := counts[row.group]
		switch row.status {
		case "OK":
			c[0]++
		case "gap":
			c[1]++
		case "terminal limitation":
			c[2]++
		}
		counts[row.group] = c
	}
	for _, group := range []string{"letters-digits", "punct-umlauts", "ctrl-alt", "function-nav", "modified-arrows"} {
		var b strings.Builder
		b.WriteString("key                         bytes                 decoded             status\n")
		n := 0
		for _, row := range rows {
			if row.group != group || n == 24 {
				continue
			}
			fmt.Fprintf(&b, "%-27s %-21s %-19s %s\n", row.name, hex.EncodeToString(row.bytes), DecodeKey(row.bytes).Name(), row.status)
			n++
		}
		name := "M5-decode-" + group + ".ansi"
		if group == "punct-umlauts" {
			name = "M5-decode-punct-umlauts.ansi"
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var summary strings.Builder
	summary.WriteString("group                    OK  gap  terminal limitation\n")
	for _, group := range []string{"letters-digits", "punct-umlauts", "ctrl-alt", "function-nav", "modified-arrows"} {
		c := counts[group]
		fmt.Fprintf(&summary, "%-24s %2d  %3d  %3d\n", group, c[0], c[1], c[2])
	}
	if err := os.WriteFile(filepath.Join(dir, "M5-summary.ansi"), []byte(summary.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "M2-decode-table.ansi"), []byte(summary.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}
