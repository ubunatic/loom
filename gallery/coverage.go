// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package gallery

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"ubunatic.com/loom"
	"ubunatic.com/loom/graph"
	"ubunatic.com/loom/measure"
	"ubunatic.com/loom/syntax"
)

func init() {
	demos["AlignBox"] = func() loom.Widget {
		a := loom.NewCenter(loom.NewView([]string{"Centered content"}))
		a.Width, a.Height = 24, 3
		return a
	}
	demos["AnsiEditor"] = func() loom.Widget {
		b, _ := loom.ParseAnsiBuffer("Edit this ANSI canvas\nUse arrows and type", 40, 8)
		return loom.NewAnsiEditor(b)
	}
	demos["Confirm"] = func() loom.Widget { return loom.NewConfirm("Continue?") }
	demos["Frame"] = func() loom.Widget {
		return &loom.Frame{Title: "Frame", Status: "Two boxed views", Boxes: []loom.Box{{Title: "First", Child: loom.NewView([]string{"First pane"})}, {Title: "Second", Child: loom.NewView([]string{"Second pane"})}}}
	}
	demos["Notif"] = func() loom.Widget { return &loom.Notif{Message: "Changes saved"} }
	demos["Router"] = func() loom.Widget {
		w, _, err := loom.BuildWidget(strings.NewReader("views:\n  - name: main\n    elements: {a: {type: view, static: Router main view}}\n  - name: details\n    elements: {a: {type: view, static: Router details view}}\n"))
		if err != nil {
			return loom.NewView([]string{err.Error()})
		}
		return &routerDemo{router: w.(*loom.Router)}
	}
	demos["Rows"] = func() loom.Widget {
		return &loom.Rows{Columns: []loom.RowColumn{{Width: 16, Bold: true}, {Width: 8, Align: "right"}}, Gap: 2, Values: [][]string{{"Requests", "1200"}, {"Errors", "4"}}}
	}
	demos["SplashView"] = func() loom.Widget {
		s := loom.NewSplashView("Starting Loom", loom.ProviderPill{Name: "API", Symbol: "✓", State: loom.ProviderDone})
		s.Progress = 65
		s.StepText = "Loading workspace"
		return s
	}
	demos["Split"] = func() loom.Widget {
		return loom.NewSplit(loom.NewView([]string{"Left pane"}), loom.NewView([]string{"Right pane"}))
	}
	demos["Stack"] = func() loom.Widget {
		return loom.NewStack(loom.Vertical, loom.NewView([]string{"Top pane"}), loom.NewView([]string{"Bottom pane"}))
	}
	demos["StyledRows"] = func() loom.Widget {
		return loom.NewStyledRows("Plain text", "\x1b[31mRed\x1b[0m and \x1b[1mbold\x1b[0m")
	}
	demos["Gauge"] = graphDemo
}

type routerDemo struct{ router *loom.Router }

func (d *routerDemo) Unwrap() loom.Widget { return d.router }

func (d *routerDemo) Draw(c *loom.Canvas, r loom.Rect) {
	d.router.Draw(c, r)
	if r.H > 1 {
		c.Write(r.X, r.Y+r.H-1, "Enter: switch view  Esc: back", loom.Style{})
	}
}
func (d *routerDemo) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if e.Key == "enter" {
		if d.router.Current() == "main" {
			d.router.RouteTo("details")
		} else {
			d.router.RouteTo("main")
		}
		return loom.Handled()
	}
	return d.router.ConsumeKey(e)
}
func (d *routerDemo) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	return d.router.ConsumeMouse(e)
}

func graphDemo() loom.Widget {
	lines := []string{"Gauge", graph.RenderBar(65, graph.BarOptions{Width: 24}), "Stacked bar", graph.RenderStackedBar([]float64{5, 3, 2}, graph.StackedBarOptions{Width: 24}), "Treemap"}
	lines = append(lines, graph.RenderTreemap([]graph.TreemapSegment{{Name: "API", Value: 5}, {Name: "Worker", Value: 3}, {Name: "Cache", Value: 2}}, graph.TreemapOptions{Width: 32, Height: 8})...)
	return loom.NewStack(loom.Vertical, &loom.Gauge{Value: 65, Min: 0, Max: 100, Width: 24}, loom.NewStyledRows(lines...))
}

type textEntryControls struct {
	entry              *loom.TextEntry
	alternate          bool
	lastSubmittedPlain string
	lastSubmittedMD    string
}

func (d *textEntryControls) Unwrap() loom.Widget { return d.entry }

func newTextAreaDemo() loom.Widget {
	a := loom.NewTextArea("A multi-line editor\npackage main\nfunc main() { println(\"hello\") }\nUse the arrow keys to move.")
	a.SetHighlighter(syntax.NewLexicalEngine("go"))
	return &textAreaWidget{area: a}
}

func newTextEntryControls() loom.Widget {
	d := &textEntryControls{}
	d.entry = loom.NewTextEntry(loom.TextEntryOptions{Placeholder: "Message the team…", Prompt: "> ", MaxHeight: 4, Highlight: leadingCommandSpans, OnSubmit: func(value string) loom.EventResult {
		d.lastSubmittedPlain = value
		d.lastSubmittedMD = d.entry.MarkdownValue()
		d.entry.SetValue("")
		return loom.Handled()
	}})
	d.entry.EnableBold()
	d.entry.SetValue("/topic Draft: a longer message that wraps across several rows of the chat input.")
	d.entry.SetFocus(true)
	return d
}

// textEntryDemoRows is the entry height in the demo. It is small on purpose, so
// a few lines of text already make the entry scroll.
const textEntryDemoRows = 4

// textEntrySamples are inserted at the caret by the F5-F8 keys.
var textEntrySamples = map[string]string{
	"f5": "A short sample.",
	"f6": "line one\nline two\nline three\nline four\nline five\nline six",
	"f7": "A long sample without any line breaks that has to wrap across several rows of the entry, so the soft wrap and the scroll position can be watched. ",
	"f8": "wide: 日本語のテキスト 🙂 e\u0301 emoji and combining marks",
}

func (d *textEntryControls) Draw(c *loom.Canvas, r loom.Rect) {
	controlsHeight := min(1, max(0, r.H))
	avail := r.H - controlsHeight
	reserve := 0 // a submitted preview keeps two rows, so its overflow mark stays readable
	if d.lastSubmittedMD != "" {
		reserve = 2
	}
	entryHeight := min(textEntryDemoRows, max(1, avail-reserve))
	entryHeight = min(entryHeight, avail)
	if entryHeight < 1 {
		if r.H > 0 {
			d.entry.Draw(c, loom.Rect{X: r.X, Y: r.Y, W: r.W, H: 1})
		}
		return
	}
	rest := avail - entryHeight
	debug := d.debugLines()
	if len(debug) > rest-reserve {
		debug = debug[:max(0, rest-reserve)]
	}
	previewBudget := rest - len(debug)
	preview := []string(nil)
	if d.lastSubmittedMD != "" && r.W > 0 && previewBudget > 0 {
		preview = wrapGalleryPreview("Submitted Markdown: "+strconv.Quote(d.lastSubmittedMD), r.W)
		if len(preview) > previewBudget {
			preview = appendGalleryOverflow(preview[:previewBudget], r.W)
		}
	}
	d.entry.Draw(c, loom.Rect{X: r.X, Y: r.Y, W: r.W, H: entryHeight})
	y := r.Y + entryHeight
	for _, line := range debug {
		writeGalleryClipped(c, r.X, y, r.W, line)
		y++
	}
	for _, line := range preview {
		c.Write(r.X, y, line, loom.Style{})
		y++
	}
	if controlsHeight > 0 {
		writeGalleryClipped(c, r.X, r.Y+r.H-1, r.W, "Ctrl+B: bold  F2: placeholder  F3 accept  F4 caret  Enter: submit Markdown")
	}
}

// debugLines describes the entry state shown below it.
func (d *textEntryControls) debugLines() []string {
	v := d.entry.Value()
	row, col := d.entry.Cursor()
	return []string{
		fmt.Sprintf("debug: %d runes, %d lines, offset %d, caret row %d col %d", utf8.RuneCountInString(v), strings.Count(v, "\n")+1, d.entry.CursorOffset(), row, col),
		fmt.Sprintf("debug: selection %v, focused %v, entry rows %d", d.entry.HasSelection(), d.entry.Focused(), textEntryDemoRows),
		"insert at caret: F5 short  F6 lines  F7 long wrapping  F8 wide characters",
	}
}

func writeGalleryClipped(c *loom.Canvas, x, y, width int, text string) {
	if width < 1 {
		return
	}
	var clipped strings.Builder
	used := 0
	for _, cluster := range measure.Clusters(text) {
		w := measure.ClusterWidth(cluster)
		if used+w > width {
			break
		}
		clipped.WriteString(cluster)
		used += w
	}
	c.Write(x, y, clipped.String(), loom.Style{})
}

func wrapGalleryPreview(text string, width int) []string {
	if width < 1 {
		return nil
	}
	var lines []string
	var line strings.Builder
	lineWidth := 0
	for _, cluster := range measure.Clusters(text) {
		clusterWidth := measure.ClusterWidth(cluster)
		if lineWidth > 0 && lineWidth+clusterWidth > width {
			lines = append(lines, line.String())
			line.Reset()
			lineWidth = 0
		}
		if clusterWidth > width {
			continue
		}
		line.WriteString(cluster)
		lineWidth += clusterWidth
	}
	if line.Len() > 0 {
		lines = append(lines, line.String())
	}
	return lines
}

func appendGalleryOverflow(lines []string, width int) []string {
	if len(lines) == 0 || width < 1 {
		return lines
	}
	marker := "… [overflow]"
	markerWidth := loom.StringWidth(marker)
	if markerWidth > width {
		marker = "!"
		markerWidth = 1
	}
	last := lines[len(lines)-1]
	for loom.StringWidth(last)+markerWidth > width {
		clusters := measure.Clusters(last)
		if len(clusters) == 0 {
			break
		}
		last = strings.Join(clusters[:len(clusters)-1], "")
	}
	lines[len(lines)-1] = last + marker
	return lines
}
func (d *textEntryControls) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	switch e.Key {
	case "f2":
		d.alternate = !d.alternate
		if d.alternate {
			d.entry.SetPlaceholder("Write an update…")
		} else {
			d.entry.SetPlaceholder("Message the team…")
		}
		d.entry.SetValue("")
		return loom.Handled()
	case "f3":
		d.entry.AcceptPlaceholder()
		return loom.Handled()
	case "f4":
		d.entry.SetCursorOffset(0)
		return loom.Handled()
	case "f5", "f6", "f7", "f8":
		return d.entry.ConsumePaste(loom.PasteEvent{Text: textEntrySamples[e.Key]})
	}
	result := d.entry.ConsumeKey(e)
	if result.Consumed {
		return result
	}
	if e.Key == "ctrl-c" && d.entry.Value() != "" {
		d.entry.SetValue("")
		return loom.Handled()
	}
	return result
}
func (d *textEntryControls) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	return d.entry.ConsumeMouse(e)
}
func (d *textEntryControls) ConsumePaste(e loom.PasteEvent) loom.EventResult {
	return d.entry.ConsumePaste(e)
}
func (d *textEntryControls) SetFocus(f bool) { d.entry.SetFocus(f) }
func (d *textEntryControls) Focused() bool   { return d.entry.Focused() }
