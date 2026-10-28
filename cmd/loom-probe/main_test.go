package main

import (
	"strings"
	"testing"
)

func TestParseCursorReport(t *testing.T) {
	row, col, err := parseCursorReport("\x1b[7;9R")
	if err != nil {
		t.Fatalf("parseCursorReport() error = %v", err)
	}
	if row != 7 || col != 9 {
		t.Fatalf("parseCursorReport() = (%d, %d), want (7, 9)", row, col)
	}
}

func TestParseCursorReportRejectsInvalidReply(t *testing.T) {
	for _, reply := range []string{"", "\x1b[;9R", "\x1b[7;9", "7;9R", "\x1b[0;9R"} {
		if _, _, err := parseCursorReport(reply); err == nil {
			t.Errorf("parseCursorReport(%q) succeeded, want error", reply)
		}
	}
}

func TestCursorAdvanceFromColumnOne(t *testing.T) {
	if got := cursorAdvance(1, 3); got != 2 {
		t.Fatalf("cursorAdvance(1, 3) = %d, want 2", got)
	}
}

func TestParseKittyQueryReply(t *testing.T) {
	tests := map[string]kittySupport{
		"":                           {},
		"\x1b[?62;22c":               {answered: true},
		"\x1b[?0u\x1b[?62;22c":       {supported: true, flags: 0, answered: true},
		"\x1b[?1u\x1b[?1;2c":         {supported: true, flags: 1, answered: true},
		"\x1b[?1u":                   {supported: true, flags: 1},
		"junk\x1b[?31u\x1b[?65;1;9c": {supported: true, flags: 31, answered: true},
	}
	for reply, want := range tests {
		if got := parseKittyQueryReply(reply); got != want {
			t.Errorf("parseKittyQueryReply(%q) = %+v, want %+v", reply, got, want)
		}
	}
}

func TestProbeSummaryEndsWithNewline(t *testing.T) {
	var output strings.Builder
	writeResults(&output, []result{{sample: sample{name: "ZWJ family"}, row: 1, col: 3}})
	if got := output.String(); !strings.HasSuffix(got, "\n") {
		t.Fatalf("probe summary lacks trailing newline: %q", got)
	}
}
