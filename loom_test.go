// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"
	"time"

	"codeberg.org/ubunatic/loom"
)

// ── Widget interface conformance ──────────────────────────────────────────────

// widgetCheck compiles only if T implements Widget. Used as a static check.
func widgetCheck[T loom.Widget](_ T) {}

func TestWidgetConformance(t *testing.T) {
	widgetCheck(loom.NewChoice(nil))
	widgetCheck(loom.NewView(nil))
	widgetCheck(loom.NewStack(loom.Vertical))
	widgetCheck(loom.NewGrid(2))
	widgetCheck(loom.NewPopup("", loom.NewView(nil)))
	widgetCheck(&loom.Notif{})
}

// ── Canvas ────────────────────────────────────────────────────────────────────

func TestCanvasSetGet(t *testing.T) {
	c := loom.NewCanvas(5, 3)
	cell := loom.Cell{Text: "X", Style: loom.Reset}
	c.Set(2, 1, cell)
	got := c.Get(2, 1)
	if got.Text != "X" {
		t.Errorf("Get(2,1).Text = %q, want %q", got.Text, "X")
	}
}

func TestCanvasOutOfBoundsIgnored(t *testing.T) {
	c := loom.NewCanvas(4, 4)
	c.Set(-1, 0, loom.Cell{Text: "!"}) // must not panic
	c.Set(10, 10, loom.Cell{Text: "!"})
	got := c.Get(-1, 0)
	if got.Text != " " {
		t.Errorf("Get(-1,0).Text = %q, want blank", got.Text)
	}
}

func TestCanvasFill(t *testing.T) {
	c := loom.NewCanvas(6, 4)
	c.Fill(loom.Rect{X: 1, Y: 1, W: 3, H: 2}, loom.Cell{Text: "#"})
	for y := 1; y <= 2; y++ {
		for x := 1; x <= 3; x++ {
			if got := c.Get(x, y).Text; got != "#" {
				t.Errorf("Get(%d,%d).Text = %q, want #", x, y, got)
			}
		}
	}
	// outside fill rect must remain blank
	if got := c.Get(0, 0).Text; got != " " {
		t.Errorf("Get(0,0).Text = %q, want blank", got)
	}
}

func TestCanvasWrite(t *testing.T) {
	c := loom.NewCanvas(10, 1)
	n := c.Write(2, 0, "hello", loom.Reset)
	if n != 5 {
		t.Errorf("Write returned %d, want 5", n)
	}
	if got := c.Get(2, 0).Text; got != "h" {
		t.Errorf("Get(2,0).Text = %q, want h", got)
	}
	if got := c.Get(6, 0).Text; got != "o" {
		t.Errorf("Get(6,0).Text = %q, want o", got)
	}
}

func TestCanvasWriteClipsAtEdge(t *testing.T) {
	c := loom.NewCanvas(5, 1)
	n := c.Write(3, 0, "hello", loom.Reset) // only "he" fits
	if n != 2 {
		t.Errorf("Write returned %d, want 2 (clipped)", n)
	}
}

func TestCanvasRowContainsText(t *testing.T) {
	c := loom.NewCanvas(5, 1)
	c.Write(0, 0, "AB", loom.Reset)
	row := c.Row(0)
	if row == "" {
		t.Error("Row(0) is empty, expected ANSI string with content")
	}
}

func TestCanvasClear(t *testing.T) {
	c := loom.NewCanvas(4, 2)
	c.Set(0, 0, loom.Cell{Text: "X"})
	c.Clear()
	if got := c.Get(0, 0).Text; got != " " {
		t.Errorf("Get(0,0).Text after Clear = %q, want blank", got)
	}
}

func TestCanvasBackgroundCompositionProtectsForegroundAndCursor(t *testing.T) {
	c := loom.NewCanvas(4, 1)
	c.Write(0, 0, "A", loom.Reset)
	c.Set(1, 0, loom.Cell{Text: " ", Style: loom.Style{BG: loom.ColorIndex(4)}})
	c.CursorX, c.CursorY = 2, 0
	c.ComposeBackground(testBackground{}, c.Bounds(), time.Time{})
	if got := c.Get(0, 0).Text; got != "A" {
		t.Fatalf("foreground text overwritten: %q", got)
	}
	if got := c.Get(1, 0).Style.BG; got != loom.ColorIndex(4) {
		t.Fatalf("styled blank overwritten: %+v", got)
	}
	if got := c.Get(2, 0).Text; got != " " {
		t.Fatalf("cursor cell painted: %q", got)
	}
	if c.CursorX != 2 || c.CursorY != 0 {
		t.Fatalf("cursor moved to (%d,%d)", c.CursorX, c.CursorY)
	}
	if got := c.Get(3, 0).Text; got != "*" {
		t.Fatalf("eligible cell = %q, want *", got)
	}
}

func TestCanvasBackgroundCompositionPassesThroughTransparentBlankCells(t *testing.T) {
	c := loom.NewCanvas(3, 1)
	c.Fill(c.Bounds(), loom.Cell{Text: " ", Style: loom.Style{FG: loom.ColorRGB(200, 200, 200)}})
	c.Set(1, 0, loom.Cell{Text: " ", Style: loom.Style{BG: loom.ColorIndex(24)}})
	c.ComposeBackground(testBackground{}, c.Bounds(), time.Time{})

	if got := c.Get(0, 0).Text; got != "*" {
		t.Fatalf("transparent blank cell = %q, want background glyph", got)
	}
	if got := c.Get(1, 0).Text; got == "*" {
		t.Fatal("colored blank cell was overwritten by background")
	}
}

func TestCanvasBackgroundCompositionHonorsExplicitBlankClaim(t *testing.T) {
	c := loom.NewCanvas(1, 1)
	c.Set(0, 0, loom.Cell{Claim: true})
	c.ComposeBackground(testBackground{}, c.Bounds(), time.Time{})
	if got := c.Get(0, 0).Text; got != " " {
		t.Fatalf("claimed blank cell = %q, want blank", got)
	}
}

func TestCanvasPaintSurfaceAllowsDecorationAndPreservesColor(t *testing.T) {
	c := loom.NewCanvas(1, 1)
	c.PaintSurface(c.Bounds(), loom.Style{BG: loom.ColorIndex(7)})
	c.ComposeBackground(testBackground{}, c.Bounds(), time.Time{})
	if got := c.Get(0, 0); got.Text != "*" || got.Style.BG != loom.ColorIndex(7) {
		t.Fatalf("surface decoration = %+v, want star with inherited surface", got)
	}
}

func TestCanvasBackgroundCompositionPreservesNestedSurfaceInheritance(t *testing.T) {
	parent := loom.NewCanvas(1, 1)
	parent.Set(0, 0, loom.Cell{Text: " ", Style: loom.Style{BG: loom.ColorIndex(7)}})
	child := loom.NewCanvas(1, 1)
	child.ComposeBackground(testBackground{}, child.Bounds(), time.Time{})
	parent.Set(0, 0, child.Get(0, 0))
	if got := parent.Get(0, 0).Style.BG; got != loom.ColorIndex(7) {
		t.Fatalf("inherited background = %+v, want parent surface", got)
	}
}

func TestAstraBackgroundCadenceAndQuantizedFrames(t *testing.T) {
	background := loom.NewAstraBackground()
	interval := background.BackgroundInterval()
	if interval <= 0 {
		t.Fatalf("background interval = %s, want positive", interval)
	}

	first := loom.NewCanvas(32, 4)
	second := loom.NewCanvas(32, 4)
	background.DrawBackgroundAt(first, first.Bounds(), time.Unix(0, interval.Nanoseconds()))
	background.DrawBackgroundAt(second, second.Bounds(), time.Unix(0, interval.Nanoseconds()+1))
	for y := 0; y < first.Rows(); y++ {
		for x := 0; x < first.Cols(); x++ {
			if got, want := second.Get(x, y), first.Get(x, y); got != want {
				t.Fatalf("frame changed inside one cadence interval at (%d,%d): got %+v, want %+v", x, y, got, want)
			}
		}
	}
}

var _ loom.AnimatedBackground = loom.NewAstraBackground()
var _ loom.BackgroundCadence = loom.NewAstraBackground()

type testBackground struct{}

func (testBackground) DrawBackground(c *loom.Canvas, r loom.Rect) {
	for x := r.X; x < r.X+r.W; x++ {
		c.Set(x, r.Y, loom.Cell{Text: "*"})
	}
}

// ── Rect ─────────────────────────────────────────────────────────────────────

func TestRectContains(t *testing.T) {
	r := loom.Rect{X: 2, Y: 3, W: 4, H: 2}
	cases := []struct {
		x, y int
		want bool
	}{
		{2, 3, true},
		{5, 4, true},
		{6, 3, false}, // X == X+W
		{2, 5, false}, // Y == Y+H
		{0, 0, false},
	}
	for _, tc := range cases {
		if got := r.Contains(tc.x, tc.y); got != tc.want {
			t.Errorf("Contains(%d,%d) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

// ── DecodeKey ─────────────────────────────────────────────────────────────────

func TestDecodeKeyArrowsCSI(t *testing.T) {
	cases := []struct {
		b    []byte
		want string
	}{
		{[]byte{27, '[', 'A'}, "up"},
		{[]byte{27, '[', 'B'}, "down"},
		{[]byte{27, '[', 'C'}, "right"},
		{[]byte{27, '[', 'D'}, "left"},
	}
	for _, tc := range cases {
		got := loom.DecodeKey(tc.b)
		if got.Key != tc.want {
			t.Errorf("DecodeKey(%v).Key = %q, want %q", tc.b, got.Key, tc.want)
		}
	}
}

func TestDecodeKeyHomeEndDelete(t *testing.T) {
	cases := []struct {
		b    []byte
		want string
	}{
		{[]byte{27, '[', 'H'}, "home"},
		{[]byte{27, 'O', 'H'}, "home"},
		{[]byte{27, '[', 'F'}, "end"},
		{[]byte{27, 'O', 'F'}, "end"},
		{[]byte{27, '[', '1', '~'}, "home"},
		{[]byte{27, '[', '7', '~'}, "home"},
		{[]byte{27, '[', '4', '~'}, "end"},
		{[]byte{27, '[', '8', '~'}, "end"},
		{[]byte{27, '[', '3', '~'}, "delete"},
	}
	for _, tc := range cases {
		got := loom.DecodeKey(tc.b)
		if got.Key != tc.want {
			t.Errorf("DecodeKey(%v).Key = %q, want %q", tc.b, got.Key, tc.want)
		}
	}
}

// ZSH ZLE leaves DECCKM active after zle -I, sending \x1bO instead of \x1b[.
// Both prefixes must decode to the same key names. See docs/TuiInput.md §3.
func TestDecodeKeyArrowsApplicationMode(t *testing.T) {
	cases := []struct {
		b    []byte
		want string
	}{
		{[]byte{27, 'O', 'A'}, "up"},
		{[]byte{27, 'O', 'B'}, "down"},
		{[]byte{27, 'O', 'C'}, "right"},
		{[]byte{27, 'O', 'D'}, "left"},
	}
	for _, tc := range cases {
		got := loom.DecodeKey(tc.b)
		if got.Key != tc.want {
			t.Errorf("DecodeKey(%v).Key = %q, want %q (DECCKM not handled)", tc.b, got.Key, tc.want)
		}
	}
}

func TestDecodeKeyControlCodes(t *testing.T) {
	cases := []struct {
		b    byte
		want string
	}{
		{3, "ctrl-c"},
		{4, "ctrl-d"},
		{9, "tab"},
		{10, "enter"},
		{13, "enter"},
		{17, "ctrl-q"},
		{23, "ctrl-w"},
		{27, "esc"},
		{127, "backspace"},
		{8, "backspace"},
	}
	for _, tc := range cases {
		got := loom.DecodeKey([]byte{tc.b})
		if got.Key != tc.want {
			t.Errorf("DecodeKey([%d]).Key = %q, want %q", tc.b, got.Key, tc.want)
		}
	}
}

func TestDecodeKeyText(t *testing.T) {
	got := loom.DecodeKey([]byte("hi"))
	if got.Key != "" || got.Text != "hi" {
		t.Errorf("DecodeKey(hi) = {%q,%q}, want {%q,%q}", got.Key, got.Text, "", "hi")
	}
}

// ── Choice ────────────────────────────────────────────────────────────────────

func TestChoiceNavigation(t *testing.T) {
	items := []loom.Item{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	c := loom.NewChoice(items)

	c.HandleKey(loom.KeyEvent{Key: "down"})
	if item, ok := c.Selected(); !ok || item.Name != "b" {
		t.Errorf("after down: selected = %+v, ok=%v; want b", item, ok)
	}

	c.HandleKey(loom.KeyEvent{Key: "up"})
	if item, ok := c.Selected(); !ok || item.Name != "a" {
		t.Errorf("after up: selected = %+v, ok=%v; want a", item, ok)
	}
}

func TestChoiceFilter(t *testing.T) {
	items := []loom.Item{{Name: "git"}, {Name: "make"}, {Name: "find"}}
	c := loom.NewChoice(items)
	c.HandleKey(loom.KeyEvent{Text: "g"})
	item, ok := c.Selected()
	if !ok || item.Name != "git" {
		t.Errorf("after filter 'g': selected = %+v, ok=%v; want git", item, ok)
	}
}

func TestChoiceAbort(t *testing.T) {
	c := loom.NewChoice([]loom.Item{{Name: "x"}})
	quit := c.HandleKey(loom.KeyEvent{Key: "esc"})
	if !quit {
		t.Error("esc should return quit=true")
	}
	if !c.Aborted() {
		t.Error("Aborted() should be true after esc")
	}
}

func TestChoiceEnterQuits(t *testing.T) {
	c := loom.NewChoice([]loom.Item{{Name: "x"}})
	quit := c.HandleKey(loom.KeyEvent{Key: "enter"})
	if !quit {
		t.Error("enter with no OnSelect should return quit=true")
	}
	if c.Aborted() {
		t.Error("Aborted() should be false after enter")
	}
	if item, ok := c.Selected(); !ok || item.Name != "x" {
		t.Errorf("Selected() = %+v, %v; want x, true", item, ok)
	}
}

func TestChoiceDraw(t *testing.T) {
	items := []loom.Item{{Name: "a", Desc: "alpha"}, {Name: "b"}}
	c := loom.NewChoice(items)
	cv := loom.NewCanvas(20, 5)
	c.Draw(cv, cv.Bounds()) // must not panic
}

// ── View ──────────────────────────────────────────────────────────────────────

func TestViewDraw(t *testing.T) {
	v := loom.NewView([]string{"hello", "world"})
	c := loom.NewCanvas(10, 3)
	v.Draw(c, c.Bounds()) // must not panic
	// first line should start with 'h'
	if got := c.Get(0, 0).Text; got != "h" {
		t.Errorf("Get(0,0).Text = %q, want h", got)
	}
}

func TestViewScrollKey(t *testing.T) {
	v := loom.NewView([]string{"a", "b", "c"})
	v.HandleKey(loom.KeyEvent{Key: "down"})
	if v.Scroll != 1 {
		t.Errorf("Scroll = %d after down, want 1", v.Scroll)
	}
	v.HandleKey(loom.KeyEvent{Key: "up"})
	if v.Scroll != 0 {
		t.Errorf("Scroll = %d after up, want 0", v.Scroll)
	}
}

// ── Stack ─────────────────────────────────────────────────────────────────────

func TestStackDraw(t *testing.T) {
	a := loom.NewView([]string{"A"})
	b := loom.NewView([]string{"B"})
	s := loom.NewStack(loom.Vertical, a, b)
	c := loom.NewCanvas(10, 6)
	s.Draw(c, c.Bounds()) // must not panic
}

func TestStackTabCyclesFocus(t *testing.T) {
	a := loom.NewChoice([]loom.Item{{Name: "a1"}})
	b := loom.NewChoice([]loom.Item{{Name: "b1"}})
	s := loom.NewStack(loom.Horizontal, a, b)
	// Tab should cycle without quitting.
	quit := s.HandleKey(loom.KeyEvent{Key: "tab"})
	if quit {
		t.Error("tab should not quit")
	}
}

// ── Grid ──────────────────────────────────────────────────────────────────────

func TestGridDraw(t *testing.T) {
	children := []loom.Widget{
		loom.NewView([]string{"1"}),
		loom.NewView([]string{"2"}),
		loom.NewView([]string{"3"}),
		loom.NewView([]string{"4"}),
	}
	g := loom.NewGrid(2, children...)
	c := loom.NewCanvas(20, 8)
	g.Draw(c, c.Bounds()) // must not panic
}

func TestGridArrowNavigation(t *testing.T) {
	children := []loom.Widget{
		loom.NewView([]string{"1"}),
		loom.NewView([]string{"2"}),
		loom.NewView([]string{"3"}),
		loom.NewView([]string{"4"}),
	}
	g := loom.NewGrid(2, children...)
	g.HandleKey(loom.KeyEvent{Key: "right"}) // focus → 1
	g.HandleKey(loom.KeyEvent{Key: "down"})  // focus → 3
	if g.Focus() != 3 {
		t.Errorf("focus = %d, want 3", g.Focus())
	}
}

func TestGridOnSelect(t *testing.T) {
	children := []loom.Widget{
		loom.NewView([]string{"a"}),
		loom.NewView([]string{"b"}),
	}
	var selected int = -1
	g := loom.NewGrid(2, children...)
	g.OnSelect = func(i int) { selected = i }

	// Enter with OnSelect must not quit and must call the callback.
	quit := g.HandleKey(loom.KeyEvent{Key: "enter"})
	if quit {
		t.Error("enter with OnSelect should not quit")
	}
	if selected != 0 {
		t.Errorf("OnSelect called with %d, want 0", selected)
	}
}

func TestGridEnterWithoutOnSelectDelegatesToChild(t *testing.T) {
	// Without OnSelect, Enter is forwarded to the focused child.
	// View returns false on Enter (it doesn't handle it), so Grid should too.
	g := loom.NewGrid(1, loom.NewView([]string{"x"}))
	quit := g.HandleKey(loom.KeyEvent{Key: "enter"})
	if quit {
		t.Error("enter delegated to View should not quit")
	}
}

func TestGridFocusHighlightDraws(t *testing.T) {
	g := loom.NewGrid(2, loom.NewView([]string{"a"}), loom.NewView([]string{"b"}))
	c := loom.NewCanvas(20, 4)
	g.Draw(c, c.Bounds()) // focused cell gets FocusBG — must not panic
}

// ── Popup ─────────────────────────────────────────────────────────────────────

func TestPopupDraw(t *testing.T) {
	inner := loom.NewView([]string{"inner content"})
	p := loom.NewPopup("Title", inner)
	c := loom.NewCanvas(40, 12)
	p.Draw(c, c.Bounds()) // must not panic
}

func TestPopupEscCloses(t *testing.T) {
	inner := loom.NewView([]string{"x"})
	p := loom.NewPopup("test", inner)
	quit := p.HandleKey(loom.KeyEvent{Key: "esc"})
	if quit {
		t.Error("esc should close popup, not quit the pane")
	}
	if p.Open {
		t.Error("popup should be closed after esc")
	}
}

// ── Notif ─────────────────────────────────────────────────────────────────────

func TestNotifShowAndDraw(t *testing.T) {
	n := &loom.Notif{}
	n.Show("hello loom")
	c := loom.NewCanvas(20, 3)
	n.Draw(c, c.Bounds()) // must not panic; renders on last row
}

func TestNotifClear(t *testing.T) {
	n := &loom.Notif{Message: "old"}
	n.Clear()
	c := loom.NewCanvas(20, 1)
	n.Draw(c, c.Bounds())
	if got := c.Get(0, 0).Text; got != " " {
		t.Errorf("after Clear, Get(0,0).Text = %q, want blank", got)
	}
}

// ── Settings ──────────────────────────────────────────────────────────────────

func TestSettingsConformance(t *testing.T) {
	widgetCheck(loom.NewSettings(nil))
}

func TestSettingsBoolToggle(t *testing.T) {
	v := true
	s := loom.NewSettings([]loom.Setting{
		{Label: "flag", Kind: loom.KindBool, Bool: &v},
	})
	s.HandleKey(loom.KeyEvent{Key: "enter"})
	if v {
		t.Error("bool should be false after Enter toggle")
	}
	s.HandleKey(loom.KeyEvent{Key: "enter"})
	if !v {
		t.Error("bool should be true after second Enter toggle")
	}
}

func TestSettingsChoiceCycle(t *testing.T) {
	idx := 0
	s := loom.NewSettings([]loom.Setting{
		{Label: "theme", Kind: loom.KindChoice, Options: []string{"dark", "light", "system"}, Index: &idx},
	})
	s.HandleKey(loom.KeyEvent{Key: "right"})
	if idx != 1 {
		t.Errorf("idx = %d after right, want 1", idx)
	}
	s.HandleKey(loom.KeyEvent{Key: "left"})
	if idx != 0 {
		t.Errorf("idx = %d after left, want 0", idx)
	}
}

func TestSettingsNavigation(t *testing.T) {
	s := loom.NewSettings([]loom.Setting{
		{Label: "a", Kind: loom.KindBool, Bool: new(bool)},
		{Label: "b", Kind: loom.KindBool, Bool: new(bool)},
	})
	s.HandleKey(loom.KeyEvent{Key: "down"})
	if s.Sel() != 1 {
		t.Errorf("Sel = %d after down, want 1", s.Sel())
	}
	s.HandleKey(loom.KeyEvent{Key: "up"})
	if s.Sel() != 0 {
		t.Errorf("Sel = %d after up, want 0", s.Sel())
	}
}

func TestSettingsEscQuits(t *testing.T) {
	s := loom.NewSettings(nil)
	if !s.HandleKey(loom.KeyEvent{Key: "esc"}) {
		t.Error("esc should return quit=true")
	}
}

func TestSettingsDraw(t *testing.T) {
	v := false
	idx := 0
	str := "hello"
	s := loom.NewSettings([]loom.Setting{
		{Label: "flag", Kind: loom.KindBool, Bool: &v},
		{Label: "name", Kind: loom.KindString, Str: &str},
		{Label: "theme", Kind: loom.KindChoice, Options: []string{"dark", "light"}, Index: &idx},
	})
	c := loom.NewCanvas(40, 6)
	s.Draw(c, c.Bounds()) // must not panic
}

// ── Choice accessors ──────────────────────────────────────────────────────────

func TestChoiceFilteredAccessors(t *testing.T) {
	items := []loom.Item{{Name: "alpha"}, {Name: "beta"}, {Name: "gamma"}}
	c := loom.NewChoice(items)
	c.HandleKey(loom.KeyEvent{Text: "a"}) // filters to "alpha", "gamma"

	if c.FilteredSel() != 0 {
		t.Errorf("FilteredSel = %d, want 0", c.FilteredSel())
	}
	got := c.FilteredItem(0)
	if got.Name != "alpha" {
		t.Errorf("FilteredItem(0).Name = %q, want alpha", got.Name)
	}
	zero := c.FilteredItem(99)
	if zero.Name != "" {
		t.Errorf("FilteredItem(99).Name = %q, want empty", zero.Name)
	}
}

func TestWideCharactersAndDebugBorders(t *testing.T) {
	// 1. Test Canvas.Write with wide characters
	c := loom.NewCanvas(5, 1)
	c.Write(0, 0, "🔍A", loom.Reset)

	// Cell at 0 should be "🔍"
	if got := c.Get(0, 0).Text; got != "🔍" {
		t.Errorf("expected cell 0 to be 🔍, got %q", got)
	}
	if c.Get(0, 0).Continuation {
		t.Error("expected cell 0 to not be a continuation cell")
	}

	// Cell at 1 should be a continuation cell (empty text)
	if got := c.Get(1, 0).Text; got != "" {
		t.Errorf("expected cell 1 to have empty text, got %q", got)
	}
	if !c.Get(1, 0).Continuation {
		t.Error("expected cell 1 to be a continuation cell")
	}

	// Cell at 2 should be "A"
	if got := c.Get(2, 0).Text; got != "A" {
		t.Errorf("expected cell 2 to be A, got %q", got)
	}

	// 2. Test Canvas.Row rendering output
	row := c.Row(0)
	// The continuation cell at index 1 must be skipped. The remaining canvas cells (width 5) are spaces.
	expected := "🔍A  \x1b[0m"
	if row != expected {
		t.Errorf("Row(0) = %q, want %q", row, expected)
	}

	// 3. Test drawDebugBorder not overpainting continuation cells
	c2 := loom.NewCanvas(3, 3)

	// We create a view with lines containing the emoji " 🔍", which spans columns 1 and 2 of row 0
	view := loom.NewView([]string{" 🔍"})
	stack := loom.NewStack(loom.Vertical, view)

	// Enable Debug mode
	oldDebug := loom.Debug
	loom.Debug = true
	defer func() { loom.Debug = oldDebug }()

	// Draw the stack. The child View will draw itself (writing " 🔍" to row 0)
	// and then the Stack will call drawDebugBorder.
	stack.Draw(c2, c2.Bounds())

	// The debug border for top (y=0) runs from x=0 to x=2.
	// Cell (0,0) is " " (empty space), so it should get overpainted with "⠂".
	// Cell (1,0) is "🔍", so it should NOT get overpainted.
	// Cell (2,0) is the continuation cell, so it should NOT get overpainted.
	if got := c2.Get(0, 0).Text; got != "⠂" {
		t.Errorf("expected cell (0,0) to be overwritten by debug border, got %q", got)
	}
	if got := c2.Get(1, 0).Text; got != "🔍" {
		t.Errorf("expected cell (1,0) (wide char) to remain 🔍, got %q", got)
	}
	if got := c2.Get(2, 0).Text; got != "" {
		t.Errorf("expected cell (2,0) (continuation) to remain empty, got %q", got)
	}
}
