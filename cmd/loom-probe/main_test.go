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

func TestProbeSummaryEndsWithNewline(t *testing.T) {
	var output strings.Builder
	writeResults(&output, []result{{sample: sample{name: "ZWJ family"}, row: 1, col: 3}})
	if got := output.String(); !strings.HasSuffix(got, "\n") {
		t.Fatalf("probe summary lacks trailing newline: %q", got)
	}
}
