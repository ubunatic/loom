// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"context"
	"testing"
	"time"

	"golang.org/x/term"
)

func TestRunPaneCompletionExitsAndCancellationAborts(t *testing.T) {
	choice := NewChoice([]Item{{Name: "one"}})
	runner := paneableRunner{Paneable: choice}
	if result := runner.ConsumeKey(KeyEvent{Key: "enter"}); !result.Quit || result.Done {
		t.Fatalf("standalone Choice confirmation = %+v, want runner exit", result)
	}
	if item, ok := choice.Selected(); !ok || item.Name != "one" {
		t.Fatalf("confirmed selection = %+v, %t; want one", item, ok)
	}

	choice = NewChoice([]Item{{Name: "one"}})
	runner = paneableRunner{Paneable: choice}
	if result := runner.ConsumeKey(KeyEvent{Key: "esc"}); !result.Quit || result.Done {
		t.Fatalf("standalone Choice cancellation = %+v, want abort exit", result)
	}
	if _, ok := choice.Selected(); ok || !choice.Aborted() {
		t.Fatal("cancelled Choice selected an item or did not abort")
	}
}

func TestRunPaneMouseCompletionExits(t *testing.T) {
	choice := NewChoice([]Item{{Name: "one"}, {Name: "two"}})
	runner := paneableRunner{Paneable: choice}
	if result := runner.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: 1}); result != QuitResult() {
		t.Fatalf("standalone Choice click = %+v, want runner exit", result)
	}
	if item, ok := choice.Selected(); !ok || item.Name != "two" {
		t.Fatalf("clicked selection = %+v, %t; want two", item, ok)
	}
}

type paneableHooksProbe struct {
	*Choice
	ticks      []time.Time
	invalidate func()
	resetTick  func()
	ready      chan Rect
}

func (*paneableHooksProbe) TickInterval() time.Duration { return time.Second }
func (w *paneableHooksProbe) Tick(now time.Time)        { w.ticks = append(w.ticks, now) }
func (w *paneableHooksProbe) SetInvalidate(f func())    { w.invalidate = f }
func (w *paneableHooksProbe) SetResetTick(f func())     { w.resetTick = f }
func (*paneableHooksProbe) PaneRequest() PaneRequest    { return PaneRequest{Mouse: 1000} }
func (w *paneableHooksProbe) Draw(c *Canvas, r Rect) {
	w.Choice.Draw(c, r)
	select {
	case w.ready <- r:
	default:
	}
}

func TestPaneableRunnerPreservesOptionalHooks(t *testing.T) {
	child := &paneableHooksProbe{Choice: NewChoice([]Item{{Name: "one"}})}
	runner := paneableRunner{Paneable: child}
	if got := UnwrapWidget(runner); got != child {
		t.Fatalf("unwrapped runner = %T, want wrapped Paneable", got)
	}
	if got := shortestTickInterval(runner); got != time.Second {
		t.Fatalf("tick interval = %v, want one second", got)
	}
	now := time.Unix(1, 0)
	tickTree(runner, now, make(map[Widget]time.Time))
	if len(child.ticks) != 1 || child.ticks[0] != now {
		t.Fatalf("ticks = %v, want [%v]", child.ticks, now)
	}
	invalidated, reset := false, false
	bindInvalidationTree(runner, func() { invalidated = true })
	bindTickerControlTree(runner, func() { reset = true })
	if child.invalidate == nil || child.resetTick == nil {
		t.Fatal("runner hid invalidation or ticker control hook")
	}
	child.invalidate()
	child.resetTick()
	if !invalidated || !reset {
		t.Fatal("callbacks did not reach the host")
	}
	themeable, ok := UnwrapWidget(runner).(Themeable)
	if !ok {
		t.Fatal("runner hid theme hook")
	}
	theme := Theme("mc")
	themeable.ApplyTheme(theme)
	if child.Style != theme.ChoiceStyle() {
		t.Fatal("theme did not reach wrapped Choice")
	}
}

func TestStandaloneChoiceClickConfirmPTY(t *testing.T) {
	master, slave := openPTY(t)
	setPTYSize(t, master, 40, 6)
	pane := &Pane{tty: slave, fd: int(slave.Fd()), rows: 6, cols: 40, startRow: 1}
	state, err := term.GetState(pane.fd)
	if err != nil {
		t.Fatal(err)
	}
	pane.oldState = state
	child := &paneableHooksProbe{
		Choice: NewChoice([]Item{{Name: "one"}, {Name: "two"}}),
		ready:  make(chan Rect, 1),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	drainDone := make(chan struct{})
	drainPTY(master, drainDone)
	defer close(drainDone)
	defer pane.Close()
	runDone := make(chan error, 1)
	go func() { runDone <- pane.run(ctx, paneableRunner{Paneable: child}, nil, nil, nil) }()
	joined := false
	defer func() {
		cancel()
		// Join the event loop even when an assertion fails before confirmation.
		if !joined {
			<-runDone
		}
	}()
	select {
	case <-child.ready:
	case <-ctx.Done():
		t.Fatal("standalone Choice did not draw")
	}
	// SGR coordinates are 1-based; Pane converts this to child-local (0, 1).
	if _, err := master.Write([]byte("\x1b[<0;1;2M")); err != nil {
		t.Fatal(err)
	}
	err = <-runDone
	joined = true
	if err != nil {
		t.Fatalf("click-confirm did not exit standalone prompt: %v", err)
	}
	if item, ok := child.Selected(); !ok || item.Name != "two" {
		t.Fatalf("standalone clicked selection = %+v, %t; want two", item, ok)
	}
}
