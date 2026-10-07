// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"ubunatic.com/loom/layout"
)

// Frame places ordered boxes between title/status lines. A positive Breakpoint
// switches the row to a vertical stack at narrower widths; zero keeps a row.
type Frame struct {
	Title            string        `yaml:"title"`
	Status           string        `yaml:"status"`
	Gap              int           `yaml:"gap"`
	Breakpoint       int           `yaml:"breakpoint"`
	Boxes            []Box         `yaml:"boxes"`
	Actions          []FrameAction `yaml:"actions"`
	Style            FrameStyle    `yaml:"-"`
	ControlSeparator string        `yaml:"control_separator"`
	FocusNextKey     string        `yaml:"focus_next_key"` // default: tab
	FocusPrevKey     string        `yaml:"focus_prev_key"` // default: shift-tab
	focused          int
	focusSet         bool
	focusManaged     bool
	hasFocus         bool
	lastRect         Rect
	mouseCapture     int
	mouseCaptureRect Rect
	mouseCaptured    bool
	modalMouse       ModalMouseCapture
}

func (f *Frame) TickInterval() (shortest time.Duration) {
	for _, box := range f.Boxes {
		if t, ok := box.Child.(Ticker); ok && t.TickInterval() > 0 && (shortest == 0 || t.TickInterval() < shortest) {
			shortest = t.TickInterval()
		}
	}
	return shortest
}

func (f *Frame) Tick(now time.Time) {
	for _, box := range f.Boxes {
		if t, ok := box.Child.(Ticker); ok && t.TickInterval() > 0 {
			t.Tick(now)
		}
	}
}

// PaneRequest merges the terminal requirements of all boxes.
func (f *Frame) PaneRequest() (request PaneRequest) {
	first := true
	for _, box := range f.Boxes {
		if requester, ok := box.Child.(PaneRequester); ok {
			childRequest := requester.PaneRequest()
			if first {
				request, first = childRequest, false
				continue
			}
			mergePaneRequest(&request, childRequest)
		}
	}
	return request
}

// FrameStyle controls the frame background, title, and status row.
type FrameStyle struct {
	Background Style
	Title      Style
	Status     Style
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
	f.syncFocus()
	f.lastRect = r
	paintClipped(c, r, func(local *Canvas) {
		w, h := local.Cols(), local.Rows()
		local.PaintSurface(local.Bounds(), f.Style.Background)
		writeBoundedStyled(local, 0, 0, w, f.Title, f.Style.Title)
		if h < 2 {
			return
		}
		local.PaintSurface(Rect{Y: h - 1, W: w, H: 1}, f.Style.Status)
		writeBoundedStyled(local, 0, h-1, w, f.StatusText(), f.Style.Status)
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
	visible := make([]int, 0, len(f.Boxes))
	for i, b := range f.Boxes {
		if !b.Hidden {
			visible = append(visible, i)
		}
	}
	if len(visible) == 0 {
		return result
	}
	if stacked {
		// Check if any boxes have FillHeight
		anyFilling := false
		for _, index := range visible {
			if f.Boxes[index].FillHeight {
				anyFilling = true
				break
			}
		}

		var allocations []layout.Allocation
		if anyFilling {
			// Use special FillHeight allocation for stacked layout
			allocations = f.stackedFillHeightAllocations(height-2, visible, f.Gap)
		} else {
			// Use standard layout.Plan for non-filling boxes
			items := make([]layout.Item, len(visible))
			allDynamic := true
			for i, index := range visible {
				items[i] = layout.Item{Visible: true, Constraint: f.boxConstraint(f.Boxes[index], false)}
				allDynamic = allDynamic && f.Boxes[index].Dynamic
			}
			var err error
			allocations, err = layout.Plan(height-2, f.Gap, items)
			if err != nil {
				allocations = f.stackedFallback(height-2, items)
			}
			if allocations != nil && allDynamic && stackedDynamicOverflow(items, height-2, f.Gap) {
				allocations = balancedStackedAllocations(height-2, f.Gap, items)
			}
		}

		if allocations != nil {
			for i, index := range visible {
				var allocation layout.Allocation
				if anyFilling {
					// stackedFillHeightAllocations returns a slice indexed by box index
					allocation = allocations[index]
				} else {
					// layout.Plan returns a slice indexed by position in visible
					allocation = allocations[i]
				}
				box := f.Boxes[index]
				// Stacking changes flow, not the available inline space. Dynamic
				// panes therefore stretch across the frame instead of retaining
				// their horizontal preferred size.
				w := min(width, max(0, box.Width))
				if box.Dynamic {
					w = width
				}
				if w >= 2 && allocation.Size >= 2 {
					result[index] = Rect{X: 0, Y: 1 + allocation.Offset, W: w, H: allocation.Size}
				}
			}
			return result
		}
	}
	if !stacked {
		items := make([]layout.Item, len(visible))
		allDynamic := false
		for i, index := range visible {
			items[i] = layout.Item{Visible: true, Constraint: f.boxConstraint(f.Boxes[index], true)}
			allDynamic = allDynamic || f.Boxes[index].Dynamic
		}
		if allDynamic {
			if allocations, err := layout.Plan(width, f.Gap, items); err == nil {
				for i, index := range visible {
					allocation := allocations[i]
					box := f.Boxes[index]
					h := min(height-2, max(0, box.Height))
					if box.Dynamic {
						h = clampBox(box.Height, box.MinHeight, box.MaxHeight, height-2)
					}
					if box.FillHeight {
						h = clampBox(height-2, box.MinHeight, box.MaxHeight, height-2)
					}
					if allocation.Size >= 2 && h >= 2 {
						result[index] = Rect{X: allocation.Offset, Y: 1, W: allocation.Size, H: h}
					}
				}
				return result
			}
		}
	}
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

func stackedDynamicOverflow(items []layout.Item, total, gap int) bool {
	if len(items) == 0 || total < gap*(len(items)-1) {
		return false
	}
	for _, item := range items {
		if item.Constraint.Min != 2 {
			return false
		}
	}
	preferred := 0
	for _, item := range items {
		preferred += item.Constraint.Preferred
	}
	return preferred > total-gap*(len(items)-1)
}

func balancedStackedAllocations(total, gap int, items []layout.Item) []layout.Allocation {
	result := make([]layout.Allocation, len(items))
	available := total - gap*(len(items)-1)
	if available < 0 {
		return result
	}
	base, extra := available/len(items), available%len(items)
	offset := 0
	for i := range items {
		size := base
		if i < extra {
			size++
		}
		result[i] = layout.Allocation{Offset: offset, Size: size}
		offset += size + gap
	}
	return result
}

// stackedFillHeightAllocations distributes height among stacked boxes when FillHeight is present.
// Non-filling boxes get their preferred height; filling boxes share remaining space equally,
// with remainder distributed to earlier boxes.
func (f *Frame) stackedFillHeightAllocations(total int, visible []int, gap int) []layout.Allocation {
	result := make([]layout.Allocation, len(f.Boxes))
	if total <= 0 || len(visible) == 0 {
		return result
	}

	// Separate filling and non-filling boxes
	var fillingIndices, fixedIndices []int
	for _, idx := range visible {
		if f.Boxes[idx].FillHeight {
			fillingIndices = append(fillingIndices, idx)
		} else {
			fixedIndices = append(fixedIndices, idx)
		}
	}

	// If no filling boxes, fall back to normal allocation
	if len(fillingIndices) == 0 {
		return result
	}

	// Calculate space needed for fixed boxes
	fixedSpace := 0
	for _, idx := range fixedIndices {
		h := max(0, f.Boxes[idx].Height)
		fixedSpace += h
	}

	// Calculate gaps between all visible boxes
	totalGaps := gap * max(0, len(visible)-1)

	// Calculate remaining space for filling boxes
	remaining := total - fixedSpace - totalGaps
	if remaining < 0 {
		remaining = 0
	}

	// Distribute remaining space among filling boxes
	baseHeight := remaining / len(fillingIndices)
	extraHeight := remaining % len(fillingIndices)

	// Build allocations
	offset := 0
	visibleIdx := 0

	for _, idx := range visible {
		isFixedBox := false
		for _, fixedIdx := range fixedIndices {
			if idx == fixedIdx {
				isFixedBox = true
				break
			}
		}

		var boxHeight int
		if isFixedBox {
			boxHeight = max(0, f.Boxes[idx].Height)
		} else {
			// This is a filling box
			boxHeight = baseHeight
			if visibleIdx < extraHeight {
				boxHeight++
			}
			// Apply constraints
			boxHeight = clampBox(boxHeight, f.Boxes[idx].MinHeight, f.Boxes[idx].MaxHeight, boxHeight)
		}

		result[idx] = layout.Allocation{Offset: offset, Size: boxHeight}
		offset += boxHeight
		if visibleIdx < len(visible)-1 {
			offset += gap
		}
		visibleIdx++
	}

	return result
}

// stackedFallback preserves the legacy first-fit behavior for rigid boxes,
// while clipping every allocation to the available stack. A stack must never
// fall back to unconstrained preferred heights: doing so can place a box below
// the frame when a short terminal cannot satisfy every minimum.
func (f *Frame) stackedFallback(total int, items []layout.Item) []layout.Allocation {
	if total <= 0 {
		return make([]layout.Allocation, len(items))
	}
	result := make([]layout.Allocation, len(items))
	offset := 0
	for i, item := range items {
		if offset >= total {
			break
		}
		preferred := item.Constraint.Preferred
		if preferred == 0 {
			preferred = item.Constraint.Min
		}
		size := min(preferred, total-offset)
		if size < 2 {
			break
		}
		result[i] = layout.Allocation{Offset: offset, Size: size}
		offset += size
		if i+1 < len(items) {
			offset += min(f.Gap, total-offset)
		}
	}
	return result
}

func (f *Frame) boxConstraint(b Box, horizontal bool) layout.Constraint {
	if !b.Dynamic {
		value := b.Width
		if !horizontal {
			value = b.Height
		}
		return layout.Constraint{Min: value, Preferred: value, Max: value, HasMax: true}
	}
	minValue, preferred, maxValue := b.MinWidth, b.Width, b.MaxWidth
	if !horizontal {
		minValue, preferred, maxValue = b.MinHeight, b.Height, b.MaxHeight
	}
	if minValue < 2 {
		minValue = 2
	}
	if preferred < minValue {
		preferred = minValue
	}
	return layout.Constraint{Min: minValue, Preferred: preferred, Max: maxValue, HasMax: maxValue > 0}
}

func clampBox(preferred, minimum, maximum, available int) int {
	if minimum < 2 {
		minimum = 2
	}
	if preferred < minimum {
		preferred = minimum
	}
	if maximum > 0 && preferred > maximum {
		preferred = maximum
	}
	return min(available, preferred)
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

// FocusedBox returns the visible box receiving keyboard input, or nil.
func (f *Frame) FocusedBox() *Box {
	f.syncFocus()
	if !f.focusSet {
		return nil
	}
	return &f.Boxes[f.focused]
}

func (f *Frame) syncFocus() {
	if f.focusSet && f.focused >= 0 && f.focused < len(f.Boxes) && !f.Boxes[f.focused].Hidden {
		f.applyFocus()
		return
	}
	start := -1
	if f.focusSet {
		start = f.focused
	}
	f.focusSet = false
	for step := 1; step <= len(f.Boxes); step++ {
		i := (start + step) % len(f.Boxes)
		if i >= 0 && !f.Boxes[i].Hidden {
			f.focused, f.focusSet = i, true
			break
		}
	}
	f.applyFocus()
}

func (f *Frame) applyFocus() {
	for i := range f.Boxes {
		if child, ok := f.Boxes[i].Child.(Focusable); ok {
			child.SetFocus(f.frameFocused() && f.focusSet && i == f.focused && !f.Boxes[i].Hidden)
		}
	}
}

func (f *Frame) frameFocused() bool { return !f.focusManaged || f.hasFocus }

// Focused reports whether the frame is active. A root frame is active by
// default; SetFocus controls this state when the frame is nested.
func (f *Frame) Focused() bool { return f.frameFocused() }

// SetFocus activates or deactivates this frame and its focused descendant.
func (f *Frame) SetFocus(focused bool) {
	f.focusManaged = true
	f.hasFocus = focused
	f.syncFocus()
}

// FocusNext advances within the focused child, then to the next visible box.
// It returns false at the final descendant so a parent can continue traversal.
func (f *Frame) FocusNext() bool {
	f.syncFocus()
	if !f.focusSet {
		return false
	}
	if child, ok := f.Boxes[f.focused].Child.(FocusContainer); ok && child.FocusNext() {
		return true
	}
	for i := f.focused + 1; i < len(f.Boxes); i++ {
		if f.Boxes[i].Hidden {
			continue
		}
		f.focused = i
		f.applyFocus()
		focusFirstWidget(f.Boxes[i].Child)
		return true
	}
	return false
}

// FocusPrevious retreats within the focused child, then to the previous
// visible box. It returns false at the first descendant.
func (f *Frame) FocusPrevious() bool {
	f.syncFocus()
	if !f.focusSet {
		return false
	}
	if child, ok := f.Boxes[f.focused].Child.(FocusContainer); ok && child.FocusPrevious() {
		return true
	}
	for i := f.focused - 1; i >= 0; i-- {
		if f.Boxes[i].Hidden {
			continue
		}
		f.focused = i
		f.applyFocus()
		focusLastWidget(f.Boxes[i].Child)
		return true
	}
	return false
}

func (f *Frame) cycleFocus(direction int) {
	f.syncFocus()
	if !f.focusSet {
		return
	}
	for step := 1; step <= len(f.Boxes); step++ {
		i := (f.focused + direction*step + len(f.Boxes)*step) % len(f.Boxes)
		if !f.Boxes[i].Hidden {
			f.focused = i
			f.applyFocus()
			return
		}
	}
}

// ConsumeKey offers the event to the focused child before frame bindings.
func (f *Frame) ConsumeKey(k KeyEvent) EventResult {
	if box := f.FocusedBox(); box != nil && box.Child != nil {
		if res := box.Child.ConsumeKey(k); res.Consumed {
			return res
		}
	}
	if res := f.handleKey(k); res.Consumed {
		return res
	}
	return Ignored()
}

func (f *Frame) handleKey(k KeyEvent) EventResult {
	f.syncFocus()
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
			return QuitResult()
		case "toggle":
			for i := range f.Boxes {
				if f.Boxes[i].ID == a.Target {
					f.Boxes[i].Hidden = !f.Boxes[i].Hidden
					f.syncFocus()
					return Handled()
				}
			}
		}
	}
	next, prev := f.FocusNextKey, f.FocusPrevKey
	if next == "" {
		next = "tab"
	}
	if prev == "" {
		prev = "shift-tab"
	}
	switch key {
	case next:
		if !f.FocusNext() {
			f.focusFirst()
		}
		return Handled()
	case prev:
		if !f.FocusPrevious() {
			f.focusLast()
		}
		return Handled()
	}
	return Ignored()
}

func (f *Frame) ConsumePaste(e PasteEvent) EventResult {
	if box := f.FocusedBox(); box != nil && box.Child != nil {
		return DispatchPasteEvent(box.Child, e)
	}
	return Ignored()
}

func (f *Frame) focusFirst() {
	for i := range f.Boxes {
		if f.Boxes[i].Hidden {
			continue
		}
		f.focused, f.focusSet = i, true
		f.applyFocus()
		focusFirstWidget(f.Boxes[i].Child)
		return
	}
}

func (f *Frame) focusLast() {
	for i := len(f.Boxes) - 1; i >= 0; i-- {
		if f.Boxes[i].Hidden {
			continue
		}
		f.focused, f.focusSet = i, true
		f.applyFocus()
		focusLastWidget(f.Boxes[i].Child)
		return
	}
}

// ConsumeMouse focuses clicked boxes and forwards events inside a child's bounds.
func (f *Frame) ConsumeMouse(e MouseEvent) EventResult {
	if result, captured := f.modalMouse.Dispatch(f, e); captured {
		f.mouseCaptured = false
		return result
	}
	x, y := e.X, e.Y
	if f.mouseCaptured && (e.Action == MouseDrag || e.Action == MouseRelease) {
		index := f.mouseCapture
		if index >= 0 && index < len(f.Boxes) && f.Boxes[index].Child != nil {
			captured := f.Boxes[index].Child
			e.X, e.Y = x-f.mouseCaptureRect.X, y-f.mouseCaptureRect.Y
			if e.Action == MouseRelease {
				f.mouseCaptured = false
			}
			return captured.ConsumeMouse(e)
		}
		f.mouseCaptured = false
	}
	for i, rect := range f.Layout(f.lastRect.W, f.lastRect.H) {
		if rect.W < 2 || rect.H < 2 || x < rect.X || x >= rect.X+rect.W || y < rect.Y || y >= rect.Y+rect.H {
			continue
		}
		if e.Action == MousePress {
			f.focused, f.focusSet = i, true
			f.applyFocus()
		}
		box := &f.Boxes[i]
		padding := max(0, box.Padding)
		inner := Rect{X: rect.X + 1 + padding, Y: rect.Y + 1 + padding,
			W: rect.W - 2 - 2*padding, H: rect.H - 2 - 2*padding}
		if box.Child == nil || inner.W <= 0 || inner.H <= 0 || x < inner.X || x >= inner.X+inner.W || y < inner.Y || y >= inner.Y+inner.H {
			return Handled()
		}
		e.X, e.Y = x-inner.X, y-inner.Y
		if e.Action == MousePress && e.Button == MouseLeft {
			f.mouseCapture, f.mouseCaptureRect, f.mouseCaptured = i, inner, true
		}
		return box.Child.ConsumeMouse(e)
	}
	return Ignored()
}

// Box returns a pointer to the box with matching ID, or nil if not found.
func (f *Frame) Box(id string) *Box {
	for i := range f.Boxes {
		if f.Boxes[i].ID == id {
			return &f.Boxes[i]
		}
	}
	return nil
}

// ApplyTheme restyles the frame and box chrome and forwards the theme to all
// box children that implement Themeable.
func (f *Frame) ApplyTheme(theme ThemeColors) {
	f.Style = theme.FrameStyle()
	boxStyle := theme.BoxStyle()
	for i := range f.Boxes {
		f.Boxes[i].Style = boxStyle
		if themeable, ok := f.Boxes[i].Child.(Themeable); ok {
			themeable.ApplyTheme(theme)
		}
	}
}

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
	decoder.KnownFields(false) // Allow extra fields like "styles" added in issue 037
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
		b.Border = border
		if b.Rows != nil {
			if err := b.Rows.validate(); err != nil {
				return fmt.Errorf("frame.boxes[%d]: %w", i, err)
			}
			b.Child = b.Rows
		}
		if b.Dynamic {
			preferred := b.Measure(0)
			if b.Width == 0 {
				b.Width = preferred.Width
			}
			if b.Height == 0 {
				b.Height = preferred.Height
			}
		}
		if b.Width < 2 || b.Height < 2 || b.MinWidth < 0 || b.MaxWidth < 0 || b.MinHeight < 0 || b.MaxHeight < 0 || b.MinWidth > b.Width || b.MinHeight > b.Height || b.MaxWidth > 0 && b.MaxWidth < b.Width || b.MaxHeight > 0 && b.MaxHeight < b.Height || b.Padding < 0 || b.Padding > (b.Width-2)/2 || b.Padding > (b.Height-2)/2 {
			return fmt.Errorf("frame.boxes[%d]: width/height must fit border and nonnegative padding", i)
		}
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
	if local.CursorX >= 0 && local.CursorY >= 0 {
		c.CursorX, c.CursorY = x+local.CursorX, y+local.CursorY
		c.CursorShape, c.CursorShapeSet = local.CursorShape, local.CursorShapeSet
	}
}

func writeBounded(c *Canvas, x, y, width int, text string) {
	writeBoundedStyled(c, x, y, width, text, Style{})
}

func writeBoundedStyled(c *Canvas, x, y, width int, text string, style Style) {
	end := min(c.Cols(), x+max(0, width))
	for _, cluster := range textClusters(text) {
		w := StringWidth(cluster)
		if x+w > end {
			break
		}
		c.Write(x, y, cluster, style)
		x += w
	}
}
