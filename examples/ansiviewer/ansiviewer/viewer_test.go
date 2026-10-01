// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansiviewer

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	"ubunatic.com/loom"
)

func TestBrowserListsTextAndANSI(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "plain.txt"), []byte("hello\nworld\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "color.ansi"), []byte("\x1b[31mred\x1b[0m\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.files) != 2 || b.files[0].Name() != "color.ansi" {
		t.Fatalf("files = %v", b.files)
	}
	b.selectFile(0)
	c := loom.NewCanvas(30, 4)
	b.Draw(c, c.Bounds())
	if got := c.Row(0); got == "" {
		t.Fatal("empty rendered row")
	}
}

func TestPreviewScrollbarAutoRendersOnlyOnOverflow(t *testing.T) {
	long := &browser{lines: []string{strings.Repeat("x", 20), "one", "two", "three", "four", "five", "six", "seven"}, kind: KindText}
	canvas := loom.NewCanvas(12, 4)
	long.Draw(canvas, canvas.Bounds())
	if got := canvas.Get(11, 0).Text; got != loom.SpeccedDefaults.Scrollbar.ForegroundChar {
		t.Fatalf("overflow preview cell = %q, want scrollbar thumb", got)
	}
	if got := canvas.Get(10, 0).Text; got != "x" {
		t.Fatalf("preview content width did not reserve scrollbar column: cell = %q", got)
	}

	short := &browser{lines: []string{strings.Repeat("x", 20), "two"}, kind: KindText}
	canvas.Clear()
	short.Draw(canvas, canvas.Bounds())
	if got := canvas.Get(11, 0).Text; got != "x" {
		t.Fatalf("short preview did not use final column for content: got %q", got)
	}
	for y := 0; y < 4; y++ {
		if got := canvas.Get(11, y).Text; got == loom.SpeccedDefaults.Scrollbar.ForegroundChar || got == loom.SpeccedDefaults.Scrollbar.BackgroundChar {
			t.Fatalf("short preview row %d unexpectedly draws scrollbar %q", y, got)
		}
	}
}

func TestPreviewScrollbarTrackClickAndThumbDrag(t *testing.T) {
	b := &browser{lines: []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven"}, kind: KindText}
	canvas := loom.NewCanvas(12, 4)
	b.Draw(canvas, canvas.Bounds())
	b.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 11, Y: 3})
	if b.offset == 0 {
		t.Fatal("bottom track click did not move preview offset")
	}

	b.Draw(canvas, canvas.Bounds())
	thumbY := 0
	for y := 0; y < 4; y++ {
		if canvas.Get(11, y).Text == loom.SpeccedDefaults.Scrollbar.ForegroundChar {
			thumbY = y
		}
	}
	start := b.offset
	b.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 11, Y: thumbY})
	b.ConsumeMouse(loom.MouseEvent{Action: loom.MouseDrag, Button: loom.MouseLeft, X: 11, Y: 0})
	if b.offset != 0 || b.offset == start {
		t.Fatalf("dragging preview scrollbar thumb moved offset to %d, want 0 from %d", b.offset, start)
	}
	b.ConsumeMouse(loom.MouseEvent{Action: loom.MouseRelease, Button: loom.MouseLeft, X: 11, Y: 0})
}

func TestBrowserUsesSharedNavigationAndKeepsANSISelection(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "alpha.txt"), []byte("alpha preview"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "beta.ansi"), []byte("\x1b[31mred preview\x1b[0m"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir)
	if err != nil {
		t.Fatal(err)
	}
	framed := newFramedBrowser(b, nil)
	if framed.frame.Boxes[0].Child != b.navigation {
		t.Fatal("frame does not host the shared navigation pane")
	}
	stableChoice := b.navigation.List()
	b.navigation.ConsumeKey(loom.KeyEvent{Key: "down"})
	b.syncSelection()
	if b.navigation.List() != stableChoice {
		t.Fatal("navigation replaced its stable Choice")
	}
	if b.kind != KindANSI || !strings.Contains(strings.Join(b.lines, "\n"), "red preview") {
		t.Fatalf("ANSI preview after shared selection = %q, %v", b.kind, b.lines)
	}
	b.navigation.ConsumeKey(loom.KeyEvent{Key: "/"})
	b.navigation.ConsumeKey(loom.KeyEvent{Text: "alpha"})
	if got := b.navigation.List().Query(); got != "alpha" {
		t.Fatalf("shared filter query = %q, want alpha", got)
	}
}

func TestPlainTextViewClipsAndPansUnicode(t *testing.T) {
	c := loom.NewCanvas(24, 2)
	c.Write(0, 0, "LEFT", loom.Style{})
	c.Write(18, 0, "RIGHT", loom.Style{})
	before := c.Row(0)
	b := &browser{lines: []string{"界CJK long text"}, kind: KindText}
	b.Draw(c, loom.Rect{X: 4, Y: 0, W: 6, H: 1})
	after := c.Row(0)
	if !strings.Contains(after, "LEFT") || !strings.Contains(after, "RIGHT") {
		t.Fatalf("outside cells changed: before=%q after=%q", before, after)
	}
	if !strings.Contains(after, "界") || !strings.Contains(after, "CJK") {
		t.Fatalf("bounded text content missing: %q", after)
	}

	b.offsetX = 2
	c.Clear()
	b.Draw(c, loom.Rect{W: 4, H: 1})
	if got := c.Get(0, 0).Text; got != "C" {
		t.Fatalf("first horizontally panned text cell = %q, want C", got)
	}
}

func TestBrowserPansPreviewInTwoDimensions(t *testing.T) {
	b := &browser{lines: []string{strings.Repeat("x", 40), "line 1", "line 2", "line 3", "line 4"}, kind: KindText}
	b.Draw(loom.NewCanvas(12, 2), loom.Rect{W: 12, H: 2})
	for _, key := range []string{"right", "l", "down", "j"} {
		b.ConsumeKey(loom.KeyEvent{Key: key})
	}
	if b.offsetX != 2 || b.offset != 2 {
		t.Fatalf("after right/down panning offset=(%d,%d), want (2,2)", b.offsetX, b.offset)
	}
	b.ConsumeKey(loom.KeyEvent{Key: "left"})
	b.ConsumeKey(loom.KeyEvent{Key: "up"})
	if b.offsetX != 1 || b.offset != 1 {
		t.Fatalf("after left/up panning offset=(%d,%d), want (1,1)", b.offsetX, b.offset)
	}
	b.ConsumeKey(loom.KeyEvent{Key: "home"})
	if b.offset != 0 {
		t.Fatalf("home vertical offset=%d, want 0", b.offset)
	}
	b.ConsumeKey(loom.KeyEvent{Key: "end"})
	if b.offset != len(b.lines)-2 {
		t.Fatalf("end vertical offset=%d, want %d", b.offset, len(b.lines)-2)
	}
}

func TestFramedBrowserStatusIncludesPanningHint(t *testing.T) {
	framed := newFramedBrowser(&browser{}, nil)
	if !strings.Contains(framed.frame.Status, "hjkl pan") {
		t.Fatalf("status = %q, missing panning hint", framed.frame.Status)
	}
}

func TestPlainTextViewPreservesZWJFamily(t *testing.T) {
	family := "👨‍👩‍👧‍👦"
	canvas := loom.NewCanvas(12, 1)
	b := &browser{lines: []string{family}, kind: KindText}
	b.Draw(canvas, loom.Rect{X: 0, Y: 0, W: 12, H: 1})

	if got := canvas.Get(0, 0).Text; got != family {
		t.Fatalf("replayed cell text = %q, want family sequence %q", got, family)
	}
	if row := canvas.Row(0); !strings.Contains(row, family) {
		t.Fatalf("replayed row = %q, want family sequence %q", row, family)
	}
}

func TestClassifyBinaryAndImage(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "data.bin")
	if err := os.WriteFile(bin, []byte{'x', 0, 'y'}, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Classify(bin); got != KindBinary {
		t.Fatalf("binary kind = %q", got)
	}
	if got := Classify(filepath.Join(dir, "photo.png")); got != KindImage {
		t.Fatalf("image kind = %q", got)
	}
}

func TestClassifyANSIByExtension(t *testing.T) {
	if got := Classify("sample.ansi"); got != KindANSI {
		t.Fatalf("ANSI kind = %q", got)
	}
}

func TestBrowserMetadataAndScroll(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "data.bin"), []byte{'x', 0}, 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.metadata, "binary") || len(b.lines) != 2 {
		t.Fatalf("metadata=%q lines=%v", b.metadata, b.lines)
	}
	for i := 0; i < 20; i++ {
		b.lines = append(b.lines, "line")
	}
	b.Draw(loom.NewCanvas(20, 4), loom.Rect{W: 20, H: 4})
	b.ConsumeKey(loom.KeyEvent{Key: "pgdn"})
	if b.offset != 4 {
		t.Fatalf("pgdn offset=%d, want viewport page 4", b.offset)
	}
}

func TestBrowserEscapeGoesToParentDirectory(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}

	b, err := newBrowser(child)
	if err != nil {
		t.Fatal(err)
	}
	if quit := b.ConsumeKey(loom.KeyEvent{Key: "esc"}).Quit; quit {
		t.Fatal("escape from a child directory should navigate to its parent")
	}
	if b.dir != root {
		t.Fatalf("directory after escape = %q, want %q", b.dir, root)
	}
	if b.files[b.selected].Name() != "child" {
		t.Fatalf("selection after escape = %q, want child", b.files[b.selected].Name())
	}

	b.navigation.SetRoot("")
	b.navigation.ConsumeKey(loom.KeyEvent{Key: "esc"})
	b.syncSelection()
	if quit := b.ConsumeKey(loom.KeyEvent{Key: "esc"}).Quit; quit {
		t.Fatal("escape from the filesystem root should not quit")
	}
	if quit := b.ConsumeKey(loom.KeyEvent{Key: "backspace"}).Quit; quit {
		t.Fatal("backspace from the filesystem root should not quit")
	}
}

func TestBrowserFilterAndMouseSelection(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"alpha.txt", "beta.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	b, err := newBrowser(dir)
	if err != nil {
		t.Fatal(err)
	}
	framed := newFramedBrowser(b, &astraToggle{})
	if framed.ConsumeKey(loom.KeyEvent{Key: "/"}).Quit {
		t.Fatal("slash filtering quit the browser")
	}
	if framed.ConsumeKey(loom.KeyEvent{Text: "beta"}).Quit {
		t.Fatal("typing a filter query quit the browser")
	}
	if got := b.navigation.List().Query(); got != "beta" {
		t.Fatalf("filter query = %q, want beta", got)
	}
	list := b.navigation.List()
	if list.FilteredItem(0).Name != "beta.txt" || list.FilteredItem(1).Name != "" {
		t.Fatalf("filtered files = [%q, %q], want [beta.txt]", list.FilteredItem(0).Name, list.FilteredItem(1).Name)
	}

	c := loom.NewCanvas(30, 8)
	b.navigation.Draw(c, loom.Rect{W: 30, H: 8})
	b.navigation.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, Y: 2})
	b.syncSelection()
	entry, ok := b.navigation.Selected()
	if !ok || entry.Name != "beta.txt" {
		t.Fatalf("mouse selection = %+v, ok=%v; want beta.txt", entry, ok)
	}
	if result := framed.ConsumeKey(loom.KeyEvent{Key: "esc"}); result.Quit || !result.Consumed {
		t.Fatalf("framed escape = quit:%v consumed:%v, want quit:false consumed:true", result.Quit, result.Consumed)
	}
}

func TestRecordWritesOneSnapshotAfterDelay(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "record.txt"), []byte("recorded\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := Record(context.Background(), &out, dir, time.Millisecond, 40, 8); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "record.txt") {
		t.Fatalf("recording does not show file: %q", out.String())
	}
}

func TestRecordCommandReapsSelfExitingChild(t *testing.T) {
	var out bytes.Buffer
	if err := RecordCommand(context.Background(), &out, time.Second, "sh", "-c", "printf '\\033[2Jself-exit'"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "self-exit") {
		t.Fatalf("capture = %q", out.String())
	}
}

func TestRecordCommandStopsLongLivedChild(t *testing.T) {
	var out bytes.Buffer
	if err := RecordCommand(context.Background(), &out, 20*time.Millisecond, "sh", "-c", "printf long-lived; trap 'exit 0' TERM; while :; do sleep 1; done"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "long-lived") {
		t.Fatalf("capture = %q", out.String())
	}
}

func TestRecordLongLivedKeepsPreTeardownScreen(t *testing.T) {
	var out bytes.Buffer
	if err := RecordCommand(context.Background(), &out, 40*time.Millisecond, "sh", "-c", "printf 'harnez usage'; trap 'exit 0' TERM; while :; do sleep 1; done"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "harnez usage") {
		t.Fatalf("pre-teardown screen = %q", out.String())
	}
}

func TestRunRecordWritesToStdout(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"--record", "1ms", "-o", "-", "--", "sh", "-c", "printf cli"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "cli") {
		t.Fatalf("CLI recording = %q", out.String())
	}
}

func TestRecordCommandWritesRenderedScreenOnly(t *testing.T) {
	var out bytes.Buffer
	if err := RecordCommand(context.Background(), &out, time.Second, "sh", "-c", "printf '\\033[2J\\033[31mred\\033[0m'"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "red") {
		t.Fatalf("screen = %q", out.String())
	}
	if !strings.Contains(out.String(), "\x1b[31m") {
		t.Fatalf("ANSI styling was discarded: %q", out.String())
	}
}

func TestRecordCommandStripsTerminalStateButKeepsANSIStyles(t *testing.T) {
	var out bytes.Buffer
	command := "printf '\033[?1049h\033]0;title\033\\\r\033[31mred\033[0m\033[?1049l'"
	if err := RecordCommand(context.Background(), &out, time.Second, "sh", "-c", command); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "\r\x1b[31mred\x1b[0m" {
		t.Fatalf("recording = %q, want preserved carriage return and color", got)
	}
}

func TestRecordCommandPreservesUsageANSIFixture(t *testing.T) {
	testRecordFixture(t, "harnez-usage.ansi")
}

func TestRecordCommandPreservesMCANSIFixtures(t *testing.T) {
	for _, name := range []string{"mc-julia256.ansi", "mc-mc46.ansi"} {
		t.Run(name, func(t *testing.T) {
			testRecordFixture(t, name)
		})
	}
}

func testRecordFixture(t *testing.T, name string) {
	t.Helper()
	fixture := filepath.Join("..", "..", "..", "docs", "data", name)
	want, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := RecordCommand(context.Background(), &out, time.Second, "cat", fixture); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), want) {
		t.Fatalf("recorded %s differs: got %d bytes, want %d", name, out.Len(), len(want))
	}
	if !bytes.Contains(out.Bytes(), []byte("\x1b[")) {
		t.Fatalf("recorded %s lost ANSI control sequences", name)
	}
}

func TestOpenRecorderPTYUsesRequestedGeometry(t *testing.T) {
	master, slave, err := openRecorderPTY(115, 37)
	if err != nil {
		t.Fatal(err)
	}
	defer master.Close()
	defer slave.Close()

	ws, err := unix.IoctlGetWinsize(int(slave.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Col != 115 || ws.Row != 37 {
		t.Fatalf("recorder PTY geometry = %dx%d, want 115x37", ws.Col, ws.Row)
	}
}

func TestRulerPadsPreviewWithDefaultBackground(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir)
	if err != nil {
		t.Fatal(err)
	}
	framed := newFramedBrowser(b, &astraToggle{})
	framed.ConsumeKey(loom.KeyEvent{Text: "r"})
	if !b.ruler {
		t.Fatal("r did not turn the ruler on")
	}
	b.lines, b.kind, b.offset = []string{"hello"}, KindText, 0
	c := loom.NewCanvas(30, 6)
	area := loom.Rect{W: 30, H: 6}
	c.PaintSurface(area, loom.Style{BG: loom.ColorIndex(236)})
	b.Draw(c, area)
	if got := c.Get(1, 1).Text; got != "h" {
		t.Fatalf("content starts with %q at (1,1), want h", got)
	}
	for _, p := range [][2]int{{0, 0}, {10, 0}, {0, 3}, {29, 3}, {15, 5}} {
		if got := c.Get(p[0], p[1]).Style.BG; got != loom.ColorReset() {
			t.Fatalf("ruler cell %v background = %+v, want terminal default", p, got)
		}
	}
	if got := c.Get(10, 0).Text; got != "1" {
		t.Fatalf("ruler mark at column 10 = %q, want 1", got)
	}
	framed.ConsumeKey(loom.KeyEvent{Text: "r"})
	c = loom.NewCanvas(30, 6)
	b.Draw(c, area)
	if got := c.Get(0, 0).Text; got != "h" {
		t.Fatalf("ruler off: content at (0,0) = %q, want h", got)
	}
}
