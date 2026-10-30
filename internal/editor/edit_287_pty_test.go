package editor

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/ptytest"
	"ubunatic.com/loom/measure"
)

// syncHelpPTY gives an observable acknowledgement after every queued report:
// F1 opens the help overlay only once earlier input has been processed.
func syncHelpPTY(t *testing.T, s *ptytest.Session) {
	t.Helper()
	timeout := 3 * time.Second * ptytest.RaceFactor
	s.Send("\x1bOP")
	s.WaitForUpdateState(timeout, func(rows []string) bool {
		return strings.Contains(strings.Join(rows, "\n"), "RichTextEdit Help")
	})
	s.Send("\x1b")
	s.WaitForUpdateState(timeout, func(rows []string) bool {
		return !strings.Contains(strings.Join(rows, "\n"), "RichTextEdit Help")
	})
}

func searchRowPTY(t *testing.T, s *ptytest.Session, label string) int {
	t.Helper()
	for row, text := range s.Screen() {
		if strings.Contains(text, label) {
			return row
		}
	}
	t.Fatalf("missing %q on screen:\n%s", label, strings.Join(s.Screen(), "\n"))
	return -1
}

func requireDocumentIntactPTY(t *testing.T, path, content string) {
	t.Helper()
	// loom edit saves ANSI styling even for an input named doc.txt. Compare
	// its canonical serialization rather than expecting plain input bytes.
	var expected loom.RichDocument
	expected.FromANSI(content)
	if data, err := os.ReadFile(path); err != nil || string(data) != expected.ToANSI()+"\n" {
		t.Fatalf("search changed the saved document: %q, %v", data, err)
	}
}

// Regression guard: the reported repro on the installed binary's code path.
func Test287EditPTYSearchQueryClickDoesNotEditDocument(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	content := "first word\nsecond line\n"
	s, path := startMouseEditPTY(t, content)
	timeout := 3 * time.Second * ptytest.RaceFactor
	title := loom.SpeccedDefaults.Editor.SearchTitle

	s.Send("\x06\t") // open search, then Tab to the editor
	s.WaitForUpdateState(timeout, func(rows []string) bool {
		return strings.Contains(strings.Join(rows, "\n"), "Regex")
	})
	titleRow := searchRowPTY(t, s, title)
	queryRow := titleRow + 1
	byteAt := strings.Index(s.Screen()[titleRow], title)
	runeX := len([]rune(s.Screen()[titleRow][:byteAt]))
	cellX := measure.StringWidth(s.Screen()[titleRow][:byteAt])
	inQuery := func(rows []string) bool { return strings.Contains(string([]rune(rows[queryRow])[runeX:]), "X") }
	s.Send(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", cellX+1, queryRow+1, cellX+1, queryRow+1))
	s.Send("X\x13")
	s.WaitForUpdateState(timeout, inQuery)
	syncHelpPTY(t, s)
	if !inQuery(s.Screen()) {
		t.Fatalf("X did not land in the query field: %q", s.Screen()[queryRow])
	}
	if text := strings.Join(s.Screen(), "\n"); strings.Contains(text, "· Modified") {
		t.Fatalf("typing into the query modified the document:\n%s", text)
	}
	requireDocumentIntactPTY(t, path, content)
}

// Regression guard: mode clicks switch mode and never reach the document.
func Test287EditPTYSearchModeClicks(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	content := "first word\nsecond line\n"
	s, path := startMouseEditPTY(t, content)
	timeout := 3 * time.Second * ptytest.RaceFactor
	defs := loom.SpeccedDefaults.Editor
	marked := func(label string) func([]string) bool {
		return func(rows []string) bool {
			return strings.Contains(strings.Join(rows, "\n"), label+defs.ModeMarker)
		}
	}

	s.Send("\x06")
	s.WaitForUpdateState(timeout, marked(defs.NormalLabel))
	row := searchRowPTY(t, s, defs.RegexLabel)
	clickEditPTY(t, s, row, defs.RegexLabel)
	s.WaitForUpdateState(timeout, marked(defs.RegexLabel))
	clickEditPTY(t, s, row, defs.NormalLabel)
	s.WaitForUpdateState(timeout, marked(defs.NormalLabel))
	s.Send("\x13")
	syncHelpPTY(t, s)
	if text := strings.Join(s.Screen(), "\n"); strings.Contains(text, "· Modified") {
		t.Fatalf("mode clicks modified the document:\n%s", text)
	}
	requireDocumentIntactPTY(t, path, content)
}

// Expected red before M2: reopening shows the match count of the old text.
func Test287EditPTYReopenAfterEditShowsCurrentCount(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s, _ := startMouseEditPTY(t, "alpha beta\n")
	timeout := 3 * time.Second * ptytest.RaceFactor
	screen := func(rows []string) string { return strings.Join(rows, "\n") }

	s.Send("\x06alpha")
	s.WaitForUpdateState(timeout, func(rows []string) bool { return strings.Contains(screen(rows), "1 / 1 matches") })
	s.Send("\x06") // close search
	s.WaitForUpdateState(timeout, func(rows []string) bool { return !strings.Contains(screen(rows), "Regex") })
	s.Send("\x1b[Halpha ") // Home, then insert a second "alpha"
	s.WaitForUpdateState(timeout, func(rows []string) bool { return strings.Contains(screen(rows), "· Modified") })
	s.Send("\x06")
	s.WaitForUpdateState(timeout, func(rows []string) bool { return strings.Contains(screen(rows), "Regex") })
	s.WaitForState(timeout, func(rows []string) bool { return strings.Contains(screen(rows), " matches") })
	if text := screen(s.Screen()); !strings.Contains(text, "1 / 2 matches") {
		t.Fatalf("counter after reopen does not match the edited text:\n%s", text)
	}
}
