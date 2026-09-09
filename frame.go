// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"bytes"
	"embed"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed spec/box.yaml
var frameSpecs embed.FS

// BoxBorder specifies the one-cell border glyphs and title decoration.
type BoxBorder struct {
	TopLeft     string `yaml:"top_left"`
	TopRight    string `yaml:"top_right"`
	BottomLeft  string `yaml:"bottom_left"`
	BottomRight string `yaml:"bottom_right"`
	Horizontal  string `yaml:"horizontal"`
	Vertical    string `yaml:"vertical"`
	TitlePrefix string `yaml:"title_prefix"`
	TitleSuffix string `yaml:"title_suffix"`
}

// Box is a titled border with padding and an optional isolated child.
// Width and Height are preferred outer dimensions used by Frame.
// YAML frames load Border from Loom's embedded spec; Go callers supply it.
type Box struct {
	ID      string    `yaml:"id"`
	Title   string    `yaml:"title"`
	Width   int       `yaml:"width"`
	Height  int       `yaml:"height"`
	Padding int       `yaml:"padding"`
	Border  BoxBorder `yaml:"-"`
	Child   Widget    `yaml:"-"`
	Hidden  bool      `yaml:"hidden"`
}

// Draw paints a box. Bounds smaller than a complete border are left blank.
func (b *Box) Draw(c *Canvas, r Rect) {
	paintClipped(c, r, func(local *Canvas) {
		w, h := local.Cols(), local.Rows()
		if w < 2 || h < 2 {
			return
		}
		for x := 1; x < w-1; x++ {
			local.Set(x, 0, Cell{Text: b.Border.Horizontal})
			local.Set(x, h-1, Cell{Text: b.Border.Horizontal})
		}
		for y := 1; y < h-1; y++ {
			local.Set(0, y, Cell{Text: b.Border.Vertical})
			local.Set(w-1, y, Cell{Text: b.Border.Vertical})
		}
		local.Set(0, 0, Cell{Text: b.Border.TopLeft})
		local.Set(w-1, 0, Cell{Text: b.Border.TopRight})
		local.Set(0, h-1, Cell{Text: b.Border.BottomLeft})
		local.Set(w-1, h-1, Cell{Text: b.Border.BottomRight})
		if b.Title != "" {
			writeBounded(local, 1, 0, w-2, b.Border.TitlePrefix+b.Title+b.Border.TitleSuffix)
		}
		padding := max(0, b.Padding)
		// Compare before doubling to avoid overflow for programmatic inputs.
		if b.Child != nil && padding < (w-1)/2 && padding < (h-1)/2 {
			inner := Rect{X: 1 + padding, Y: 1 + padding, W: w - 2 - 2*padding, H: h - 2 - 2*padding}
			paintClipped(local, inner, func(child *Canvas) { b.Child.Draw(child, child.Bounds()) })
		}
	})
}

// HandleKey leaves static boxes inert.
func (b *Box) HandleKey(KeyEvent) bool { return false }

// HandleMouse leaves static boxes inert.
func (b *Box) HandleMouse(MouseEvent) bool { return false }

// Frame places ordered boxes between title/status lines. A positive Breakpoint
// switches the row to a vertical stack at narrower widths; zero keeps a row.
type Frame struct {
	Title            string        `yaml:"title"`
	Status           string        `yaml:"status"`
	Gap              int           `yaml:"gap"`
	Breakpoint       int           `yaml:"breakpoint"`
	Boxes            []Box         `yaml:"boxes"`
	Actions          []FrameAction `yaml:"actions"`
	ControlSeparator string        `yaml:"control_separator"`
}

// FrameAction declares a bounded toggle or quit binding and its displayed hints.
// Toggle targets are box IDs; hints remain reachable even when all boxes hide.
type FrameAction struct {
	ID         string `yaml:"id"`
	Action     string `yaml:"action"`
	Key        string `yaml:"key"`
	Target     string `yaml:"target"`
	Hint       string `yaml:"hint"`
	HiddenHint string `yaml:"hidden_hint"`
	TitleHint  string `yaml:"title_hint"`
}

// StatusText derives current hints from visibility without mutating the spec.
func (f *Frame) StatusText() string {
	parts := []string{}
	if f.Status != "" {
		parts = append(parts, f.Status)
	}
	for _, a := range f.Actions {
		hint := a.Hint
		for _, b := range f.Boxes {
			if b.ID == a.Target && b.Hidden {
				hint = a.HiddenHint
			}
		}
		if hint != "" {
			parts = append(parts, hint)
		}
	}
	return strings.Join(parts, f.ControlSeparator)
}

// Draw reserves the first and last rows for chrome. At one row only title fits.
func (f *Frame) Draw(c *Canvas, r Rect) {
	paintClipped(c, r, func(local *Canvas) {
		w, h := local.Cols(), local.Rows()
		writeBounded(local, 0, 0, w, f.Title)
		if h < 2 {
			return
		}
		writeBounded(local, 0, h-1, w, f.StatusText())
		for i, rect := range f.Layout(w, h) {
			box := f.Boxes[i]
			for _, a := range f.Actions {
				if a.Target == box.ID {
					box.Title = a.TitleHint + box.Title
				}
			}
			box.Draw(local, rect)
		}
	})
}

// Layout returns one bounded outer rectangle per declared box. Chrome takes
// precedence, then boxes in declaration order. Unavailable boxes have zero
// rectangles; areas too small for a full border are omitted rather than broken.
func (f *Frame) Layout(width, height int) []Rect {
	result := make([]Rect, len(f.Boxes))
	if width < 2 || height < 4 {
		return result
	}
	x, y, bottom := 0, 1, height-1
	stacked := f.Breakpoint > 0 && width < f.Breakpoint
	for i, b := range f.Boxes {
		if b.Hidden {
			continue
		}
		w, h := min(max(0, b.Width), width-x), min(max(0, b.Height), bottom-y)
		if w < 2 || h < 2 {
			break
		}
		result[i] = Rect{X: x, Y: y, W: w, H: h}
		if stacked {
			y += h
			y += min(max(0, f.Gap), bottom-y)
		} else {
			x += w
			x += min(max(0, f.Gap), width-x)
		}
	}
	return result
}

// HeightForWidth reports the preferred height, including chrome and stack gaps.
// It is independent of Draw, so resize never changes application data.
func (f *Frame) HeightForWidth(width int) int {
	if f.Breakpoint <= 0 || width >= f.Breakpoint {
		return f.ContentHeight()
	}
	h := 2
	count := 0
	for _, box := range f.Boxes {
		if box.Hidden {
			continue
		}
		if count > 0 {
			h += max(0, f.Gap)
		}
		h += max(0, box.Height)
		count++
	}
	return h
}

// ContentHeight includes two chrome rows and the tallest box.
func (f *Frame) ContentHeight() int {
	h := 0
	for _, box := range f.Boxes {
		if box.Hidden {
			continue
		}
		h = max(h, box.Height)
	}
	return h + 2
}

// HandleKey dispatches declared actions. Hidden children never receive input.
func (f *Frame) HandleKey(k KeyEvent) bool {
	key := k.Key
	if key == "" {
		key = k.Text
	}
	for _, a := range f.Actions {
		if a.Key != key {
			continue
		}
		switch a.Action {
		case "quit":
			return true
		case "toggle":
			for i := range f.Boxes {
				if f.Boxes[i].ID == a.Target {
					f.Boxes[i].Hidden = !f.Boxes[i].Hidden
					return false
				}
			}
		}
	}
	return false
}

// HandleMouse leaves show-once frames inert.
func (f *Frame) HandleMouse(MouseEvent) bool { return false }

func (f *Frame) validate() error {
	if f.Breakpoint < 0 {
		return fmt.Errorf("frame.breakpoint: cannot be negative")
	}
	if f.Gap < 0 {
		return fmt.Errorf("frame.gap: cannot be negative")
	}
	if !shellText(f.Title) || !shellText(f.Status) {
		return fmt.Errorf("frame.title/status: expected printable ASCII for the static shell")
	}
	if len(f.Boxes) == 0 {
		return fmt.Errorf("frame.boxes: at least one box required")
	}
	data, err := frameSpecs.ReadFile("spec/box.yaml")
	if err != nil {
		return fmt.Errorf("frame border spec: %w", err)
	}
	var border BoxBorder
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&border); err != nil {
		return fmt.Errorf("frame border spec: %w", err)
	}
	seen := make(map[string]bool)
	for i := range f.Boxes {
		b := &f.Boxes[i]
		if b.ID == "" || !shellText(b.ID) {
			return fmt.Errorf("frame.boxes[%d].id: nonempty printable ASCII required", i)
		}
		if seen[b.ID] {
			return fmt.Errorf("frame.boxes[%d].id: duplicate %q", i, b.ID)
		}
		seen[b.ID] = true
		if !shellText(b.Title) {
			return fmt.Errorf("frame.boxes[%d].title: expected printable ASCII", i)
		}
		if b.Width < 2 || b.Height < 2 || b.Padding < 0 || b.Padding > (b.Width-2)/2 || b.Padding > (b.Height-2)/2 {
			return fmt.Errorf("frame.boxes[%d]: width/height must fit border and nonnegative padding", i)
		}
		b.Border = border
	}
	keys, ids, targets := map[string]bool{}, map[string]bool{}, map[string]bool{}
	if !shellText(f.ControlSeparator) {
		return fmt.Errorf("frame.control_separator: expected printable ASCII")
	}
	for _, a := range f.Actions {
		if a.ID == "" || ids[a.ID] || keys[a.Key] {
			return fmt.Errorf("frame.actions: empty/duplicate id or key")
		}
		if !((len(a.Key) == 1 && a.Key[0] >= '!' && a.Key[0] <= '~') || a.Key == "ctrl-c") {
			return fmt.Errorf("frame.actions: unsupported key %q", a.Key)
		}
		if !shellText(a.Hint) || !shellText(a.HiddenHint) || !shellText(a.TitleHint) {
			return fmt.Errorf("frame.actions: hints must be printable ASCII")
		}
		switch a.Action {
		case "toggle":
			if !seen[a.Target] || targets[a.Target] || a.Hint == "" || a.HiddenHint == "" {
				return fmt.Errorf("frame.actions: toggle needs a unique existing target and both hints")
			}
			targets[a.Target] = true
		case "quit":
			if a.Target != "" || a.HiddenHint != "" || a.TitleHint != "" {
				return fmt.Errorf("frame.actions: quit cannot target boxes or carry visibility hints")
			}
		default:
			return fmt.Errorf("frame.actions: unknown action %q", a.Action)
		}
		ids[a.ID], keys[a.Key] = true, true
	}
	return nil
}

func shellText(s string) bool {
	for _, r := range s {
		if r < ' ' || r > '~' {
			return false
		}
	}
	return true
}

// paintClipped gives a child its own canvas so even Fill/Clear cannot escape.
// Empty regions never pass through NewCanvas's minimum-size fallback.
func paintClipped(c *Canvas, r Rect, paint func(*Canvas)) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	x, y := max(0, r.X), max(0, r.Y)
	w, h := r.W, r.H
	if r.X < 0 {
		w += r.X
	}
	if r.Y < 0 {
		h += r.Y
	}
	w, h = min(w, c.Cols()-x), min(h, c.Rows()-y)
	if w <= 0 || h <= 0 {
		return
	}
	local := NewCanvas(w, h)
	paint(local)
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			c.Set(x+col, y+row, local.Get(col, row))
		}
	}
}

func writeBounded(c *Canvas, x, y, width int, text string) {
	end := min(c.Cols(), x+max(0, width))
	for _, cluster := range textClusters(text) {
		w := StringWidth(cluster)
		if x+w > end {
			break
		}
		c.Write(x, y, cluster, Style{})
		x += w
	}
}
