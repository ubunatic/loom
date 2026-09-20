// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Transition is called once when a startup splash hands control to its
// destination widget. The callback runs on the pane event-loop goroutine, so
// implementations must return promptly.
type Transition interface {
	Enter(ctx context.Context, from, to Widget) error
}

// TransitionFunc adapts a function to a Transition.
type TransitionFunc func(context.Context, Widget, Widget) error

// Enter implements Transition.
func (f TransitionFunc) Enter(ctx context.Context, from, to Widget) error {
	return f(ctx, from, to)
}

// StartupConfig configures Pane.RunStartup.
//
// Splash is required and owns task execution and the final completed-frame
// hold. View is the widget rendered while Splash is active; when it is nil, a
// view is created from Splash's title and provider tasks. Next is the widget
// revealed after the final splash frame. Cadence controls snapshot collection
// and redraw; zero values use the splash animation interval. Transition is
// optional and defaults to an immediate handover.
type StartupConfig struct {
	Splash    *SplashController
	View      *SplashView
	Next      Widget
	Cadence   Cadence
	Transition Transition
}

// RunStartup starts a splash controller, renders it until its completion hold
// has elapsed, and then hands the pane to the destination widget. It returns
// when the destination quits, the context is cancelled, or a transition or
// pane error occurs. The completed splash frame is always drawn once before
// handover; no caller-side sleep is needed.
func (p *Pane) RunStartup(ctx context.Context, cfg StartupConfig) error {
	if p == nil {
		return fmt.Errorf("loom: nil pane")
	}
	if cfg.Splash == nil {
		return fmt.Errorf("loom: startup splash controller required")
	}
	if cfg.Next == nil {
		return fmt.Errorf("loom: startup destination widget required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	view := cfg.View
	if view == nil {
		view = NewSplashView(cfg.Splash.cfg.Title)
		view.Controller = cfg.Splash
	}
	cadence := cfg.Cadence
	if cadence.Collect <= 0 {
		cadence.Collect = cfg.Splash.cfg.TickInterval
	}
	if cadence.Redraw <= 0 {
		cadence.Redraw = cfg.Splash.cfg.TickInterval
	}
	if err := cadence.Validate(); err != nil {
		return err
	}

	runner := &startupRoot{
		view:       view,
		next:       cfg.Next,
		controller: cfg.Splash,
		transition: cfg.Transition,
	}
	if runner.transition == nil {
		runner.transition = immediateTransition{}
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	cfg.Splash.Start(runCtx)
	defer func() {
		// Dismiss is idempotent and ensures actions that observe the controller
		// context are released if the destination quits or the pane errors.
		cfg.Splash.Dismiss()
	}()

	collect := func(time.Time) error {
		runner.collect()
		if err := runner.transitionError(); err != nil {
			return err
		}
		return nil
	}
	if err := p.RunWatch(runCtx, runner, cadence, collect); err != nil {
		if err == context.Canceled && ctx.Err() == nil && runner.transitionError() == nil {
			return nil
		}
		return err
	}
	return runner.transitionError()
}

type immediateTransition struct{}

func (immediateTransition) Enter(context.Context, Widget, Widget) error { return nil }

// startupRoot keeps the handover decision in the render lifecycle. A
// completed snapshot is collected first, then drawn, and only the following
// collection performs the transition. This ordering makes the final frame
// observable and deterministic even when collection and redraw ticks coincide.
type startupRoot struct {
	view       *SplashView
	next       Widget
	controller *SplashController
	transition Transition

	mu               sync.Mutex
	completed        bool
	completedRendered bool
	active           bool
	transitionErr    error
}

func (r *startupRoot) collect() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active {
		return
	}
	snap := r.controller.Snapshot()
	r.view.ApplySnapshot(snap)
	if snap.Completed || snap.Dismissed {
		if !r.completed {
			r.completed = true
			return
		}
		if r.completedRendered {
			if err := r.transition.Enter(context.Background(), r.view, r.next); err != nil {
				r.transitionErr = err
				return
			}
			r.active = true
		}
	}
}

func (r *startupRoot) transitionError() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.transitionErr
}

func (r *startupRoot) Draw(c *Canvas, rect Rect) {
	r.mu.Lock()
	active := r.active
	if !active {
		r.view.Draw(c, rect)
		if r.completed {
			r.completedRendered = true
		}
		r.mu.Unlock()
		return
	}
	next := r.next
	r.mu.Unlock()
	next.Draw(c, rect)
}

func (r *startupRoot) HandleKey(e KeyEvent) bool {
	r.mu.Lock()
	active := r.active
	transitionErr := r.transitionErr
	r.mu.Unlock()
	if transitionErr != nil {
		return true
	}
	if active {
		return r.next.HandleKey(e)
	}
	return r.view.HandleKey(e)
}

func (r *startupRoot) HandleMouse(e MouseEvent) bool {
	r.mu.Lock()
	active := r.active
	transitionErr := r.transitionErr
	r.mu.Unlock()
	if transitionErr != nil {
		return true
	}
	if active {
		return r.next.HandleMouse(e)
	}
	return r.view.HandleMouse(e)
}
