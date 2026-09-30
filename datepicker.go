// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"time"

	"codeberg.org/ubunatic/loom/measure"
)

// DatePicker is a date-only calendar with optional ISO date entry.
type DatePicker struct {
	Value     *time.Time
	Min       time.Time
	Max       time.Time
	WeekStart time.Weekday
	Style     TableStyle
	Now       func() time.Time
	OnSelect  func(time.Time)
	cursor    time.Time
	month     time.Time
	input     string
	grid      Rect
	focused   bool
}

// NewDatePicker binds a selected date. A nil pointer is initialized to today.
func NewDatePicker(value *time.Time) *DatePicker {
	return NewDatePickerWithClock(value, time.Now)
}

// NewDatePickerWithClock creates a picker using now for the current date and
// for the today marker. Pass a fixed clock to make initialization deterministic.
func NewDatePickerWithClock(value *time.Time, now func() time.Time) *DatePicker {
	if value == nil {
		value = new(time.Time)
	}
	if now == nil {
		now = time.Now
	}
	p := &DatePicker{Value: value, WeekStart: time.Monday, Style: DefaultTableStyle(), Now: now, focused: true}
	today := p.now()
	if value.IsZero() {
		*value = dateOnly(today)
	} else {
		*value = dateOnly(*value)
	}
	p.cursor = dateOnly(*value)
	p.month = firstOfMonth(p.cursor)
	return p
}

func (p *DatePicker) now() time.Time {
	if p.Now == nil {
		return time.Now()
	}
	return p.Now()
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
func sameDate(a, b time.Time) bool {
	ya, ma, da := a.Date()
	yb, mb, db := b.Date()
	return ya == yb && ma == mb && da == db
}
func beforeDate(a, b time.Time) bool {
	y, m, d := a.Date()
	by, bm, bd := b.Date()
	return y < by || y == by && (m < bm || m == bm && d < bd)
}
func firstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}
func (p *DatePicker) Cursor() time.Time            { return p.cursor }
func (p *DatePicker) DisplayMonth() time.Month     { return p.month.Month() }
func (p *DatePicker) Focused() bool                { return p.focused }
func (p *DatePicker) SetFocus(focused bool)        { p.focused = focused }
func (p *DatePicker) ContentHeight() int           { return 10 }
func (p *DatePicker) ApplyTheme(theme ThemeColors) { p.Style = theme.TableStyle() }

func (p *DatePicker) clamp(t time.Time) time.Time {
	t = dateOnly(t)
	if !p.Min.IsZero() && beforeDate(t, p.Min) {
		return dateOnly(p.Min)
	}
	if !p.Max.IsZero() && beforeDate(p.Max, t) {
		return dateOnly(p.Max)
	}
	return t
}

func (p *DatePicker) setCursor(t time.Time) {
	p.cursor = p.clamp(t)
	p.month = firstOfMonth(p.cursor)
}

func (p *DatePicker) navigateMonth(delta int) {
	day := p.cursor.Day()
	month := p.month.AddDate(0, delta, 0)
	last := time.Date(month.Year(), month.Month()+1, 0, 0, 0, 0, 0, month.Location()).Day()
	p.setCursor(time.Date(month.Year(), month.Month(), min(day, last), 0, 0, 0, 0, month.Location()))
}

func (p *DatePicker) selectDate(t time.Time) bool {
	t = dateOnly(t)
	if (!p.Min.IsZero() && beforeDate(t, p.Min)) || (!p.Max.IsZero() && beforeDate(p.Max, t)) {
		return false
	}
	p.cursor, p.month = t, firstOfMonth(t)
	*p.Value = t
	p.input = ""
	if p.OnSelect != nil {
		p.OnSelect(t)
	}
	return true
}

func (p *DatePicker) ConsumeKey(e KeyEvent) EventResult {
	if e.Key == "esc" {
		p.input = ""
		return Consumed()
	}
	if e.Key == "backspace" && p.input != "" {
		p.input = p.input[:len(p.input)-1]
		return Consumed()
	}
	if e.Key == "enter" {
		if p.input != "" {
			if t, err := time.ParseInLocation("2006-01-02", p.input, p.cursor.Location()); err == nil {
				p.selectDate(t)
			}
			p.input = ""
			return Consumed()
		}
		p.selectDate(p.cursor)
		return Consumed()
	}
	if e.Key == "pgup" || e.Key == "pageup" {
		p.navigateMonth(-1)
		return Consumed()
	}
	if e.Key == "pgdown" || e.Key == "pgdn" || e.Key == "pagedown" {
		p.navigateMonth(1)
		return Consumed()
	}
	if e.Key == "ctrl-left" {
		p.navigateMonth(-12)
		return Consumed()
	}
	if e.Key == "ctrl-right" {
		p.navigateMonth(12)
		return Consumed()
	}
	step := 0
	switch e.Key {
	case "left":
		step = -1
	case "right":
		step = 1
	case "up":
		step = -7
	case "down":
		step = 7
	}
	if step != 0 {
		p.setCursor(p.cursor.AddDate(0, 0, step))
		return Consumed()
	}
	if e.Key == "" && len(e.Text) == 1 {
		ch := e.Text[0]
		if ch >= '0' && ch <= '9' || ch == '-' {
			if len(p.input) < 10 {
				p.input += e.Text
			}
			return Consumed()
		}
	}
	return Ignored()
}

func (p *DatePicker) ConsumeMouse(e MouseEvent) EventResult {
	if e.Action != MousePress || e.Button != MouseLeft {
		return Ignored()
	}
	// Mouse coordinates at this widget boundary are 0-based and child-local.
	x, y := e.X, e.Y
	if x < 0 || x >= 21 || y < 3 || y >= 9 {
		return Ignored()
	}
	col, row := x/3, y-3
	startOffset := (int(firstOfMonth(p.month).Weekday()) - int(p.WeekStart) + 7) % 7
	d := row*7 + col - startOffset + 1
	if d < 1 || d > time.Date(p.month.Year(), p.month.Month()+1, 0, 0, 0, 0, 0, p.month.Location()).Day() {
		return Consumed()
	}
	t := time.Date(p.month.Year(), p.month.Month(), d, 0, 0, 0, 0, p.month.Location())
	p.selectDate(t)
	return Consumed()
}

func (p *DatePicker) Draw(c *Canvas, r Rect) {
	p.grid = r
	weekdays := []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	start := int(p.WeekStart) % 7
	var labels []string
	for i := 0; i < 7; i++ {
		labels = append(labels, weekdays[(start+i+6)%7])
	}
	write := func(x, y int, text string, width int, style Style) {
		if y < r.Y || y >= r.Y+r.H || x < r.X || x >= r.X+r.W || width <= 0 {
			return
		}
		width = min(width, r.X+r.W-x)
		c.Write(x, y, measure.Truncate(text, width, ""), style)
	}
	write(r.X, r.Y, p.month.Format("January 2006"), r.W, Style{Bold: true})
	write(r.X, r.Y+1, "←/→ day ↑/↓ week PgUp/PgDn month Ctrl+←/→ year", r.W, Style{Dim: true})
	write(r.X, r.Y+2, strings.Join(labels, " "), r.W, Style{Dim: true})
	first := firstOfMonth(p.month)
	offset := (int(first.Weekday()) - start + 7) % 7
	days := time.Date(p.month.Year(), p.month.Month()+1, 0, 0, 0, 0, 0, p.month.Location()).Day()
	today := dateOnly(p.now())
	for d := 1; d <= days; d++ {
		cell := offset + d - 1
		day := time.Date(p.month.Year(), p.month.Month(), d, 0, 0, 0, 0, p.month.Location())
		label := day.Format("_2")
		style := Style{}
		if sameDate(day, today) {
			label = "*" + day.Format("_2")
			style = Style{Bold: true}
		}
		if !p.Value.IsZero() && sameDate(day, *p.Value) {
			label = ">" + day.Format("_2")
			style = p.Style.Selected
		}
		if sameDate(day, p.cursor) {
			style = Style{Bold: true, Underline: true}
		}
		if (!p.Min.IsZero() && beforeDate(day, p.Min)) || (!p.Max.IsZero() && beforeDate(p.Max, day)) {
			style = Style{Dim: true}
		}
		write(r.X+(cell%7)*3, r.Y+3+cell/7, label, 3, style)
	}
	if p.input != "" {
		write(r.X, r.Y+9, "ISO date: "+p.input, r.W, Style{})
	} else {
		write(r.X, r.Y+9, "Type YYYY-MM-DD then Enter", r.W, Style{Dim: true})
	}
}

var _ Widget = (*DatePicker)(nil)
var _ Focusable = (*DatePicker)(nil)
var _ EventConsumer = (*DatePicker)(nil)
var _ MouseConsumer = (*DatePicker)(nil)
var _ Themeable = (*DatePicker)(nil)
