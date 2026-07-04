// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// Confirm is a yes/no prompt for preview→confirm flows (e.g. the final step of
// :git commit) and destructive-action guards. ←/→ (or h/l) move between Yes and
// No, 'y'/'n' answer directly, Enter confirms the highlighted choice, and Esc
// answers No. After the widget quits, Answered reports that a choice was made
// and Confirmed reports whether it was Yes.
type Confirm struct {
	Prompt    string
	yes       bool // currently highlighted choice
	answered  bool
	confirmed bool
}

// NewConfirm creates a Confirm with the given prompt. The Yes option is
// highlighted by default; set DefaultNo to start on No for destructive actions.
func NewConfirm(prompt string) *Confirm {
	return &Confirm{Prompt: prompt, yes: true}
}

// DefaultNo starts the prompt highlighting No (safer for destructive actions).
func (c *Confirm) DefaultNo() *Confirm {
	c.yes = false
	return c
}

// Answered reports whether the user made a choice (vs. the widget never ran).
func (c *Confirm) Answered() bool { return c.answered }

// Confirmed reports whether the user chose Yes.
func (c *Confirm) Confirmed() bool { return c.confirmed }

func (c *Confirm) answer(yes bool) {
	c.yes = yes
	c.answered = true
	c.confirmed = yes
}

// Draw renders the prompt on the first row and the Yes/No options below, with
// the highlighted option marked and bold.
func (c *Confirm) Draw(cv *Canvas, r Rect) {
	for i := 0; i < r.H; i++ {
		cv.Fill(Rect{r.X, r.Y + i, r.W, 1}, Cell{Text: " "})
	}
	if c.Prompt != "" {
		cv.Write(r.X, r.Y, c.Prompt, Style{})
	}
	optY := r.Y + 1
	if r.H < 2 {
		optY = r.Y
	}
	yesMark, noMark := "  ", "  "
	yesStyle, noStyle := Style{Dim: true}, Style{Dim: true}
	if c.yes {
		yesMark, yesStyle = "▶ ", Style{Bold: true}
	} else {
		noMark, noStyle = "▶ ", Style{Bold: true}
	}
	x := r.X
	x += cv.Write(x, optY, yesMark+"Yes", yesStyle)
	x += cv.Write(x, optY, "    ", Style{})
	cv.Write(x, optY, noMark+"No", noStyle)
}

// HandleKey drives the selection and answers the prompt.
func (c *Confirm) HandleKey(e KeyEvent) (quit bool) {
	switch e.Key {
	case "left", "right":
		c.yes = !c.yes
		return false
	case "enter":
		c.answer(c.yes)
		return true
	case "esc", "ctrl-c":
		c.answer(false)
		return true
	}
	switch e.Text {
	case "y", "Y":
		c.answer(true)
		return true
	case "n", "N":
		c.answer(false)
		return true
	case "h", "l":
		c.yes = !c.yes
	}
	return false
}

// HandleMouse is a no-op; Confirm is keyboard-driven.
func (c *Confirm) HandleMouse(MouseEvent) (quit bool) { return false }

// ContentHeight reports the rows needed: prompt + options.
func (c *Confirm) ContentHeight() int { return 2 }
