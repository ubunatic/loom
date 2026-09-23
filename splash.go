// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"context"
	"sync"
	"time"
)

// SplashState describes the lifecycle phase of the splash screen.
type SplashState int

const (
	// SplashInitializing is the initial idle state before background tasks start.
	SplashInitializing SplashState = iota
	// SplashRunning is active while background provider tasks are in flight.
	SplashRunning
	// SplashCompleted indicates all tasks finished normally.
	SplashCompleted
	// SplashDismissed indicates the splash screen was dismissed or skipped by the user.
	SplashDismissed
)

// SplashSnapshot captures an immutable state of the splash screen for rendering.
type SplashSnapshot struct {
	State        SplashState
	SpinnerFrame int
	Progress     float64 // 0.0 to 100.0
	StepText     string
	Pills        []ProviderPill
	Dismissed    bool
	Completed    bool
}

// ProviderTask defines an initialization task for a provider.
type ProviderTask struct {
	Name     string
	Symbol   string
	Duration time.Duration // simulated duration if Action is nil
	Action   func(ctx context.Context) error
}

// SplashConfig configures the SplashController.
type SplashConfig struct {
	Title        string
	Tasks        []ProviderTask
	TickInterval time.Duration // animation ticker rate; defaults to 80ms
	HoldDuration time.Duration // duration to hold completed state before transition; defaults to 400ms
}

// SplashController coordinates the animation ticks, async provider tasks,
// and user skip/dismissal interactions for a startup splash screen.
type SplashController struct {
	cfg        SplashConfig
	mu         sync.RWMutex
	state      SplashState
	frame      int
	progress   float64
	stepText   string
	pills      []ProviderPill
	doneCh     chan struct{}
	cancelFunc context.CancelFunc
	once       sync.Once
}

// NewSplashController creates and initializes a SplashController.
func NewSplashController(cfg SplashConfig) *SplashController {
	if cfg.TickInterval <= 0 {
		cfg.TickInterval = SpeccedDefaults.Splash.TickInterval
	}
	if cfg.HoldDuration <= 0 {
		cfg.HoldDuration = SpeccedDefaults.Splash.HoldDuration
	}
	pills := make([]ProviderPill, len(cfg.Tasks))
	for i, task := range cfg.Tasks {
		pills[i] = ProviderPill{
			Name:   task.Name,
			Symbol: task.Symbol,
			State:  ProviderPending,
		}
	}
	return &SplashController{
		cfg:      cfg,
		state:    SplashInitializing,
		pills:    pills,
		doneCh:   make(chan struct{}),
		stepText: SpeccedDefaults.Splash.StepText,
	}
}

// Start begins the animation tick loop and async provider task execution.
func (sc *SplashController) Start(ctx context.Context) {
	runCtx, cancel := context.WithCancel(ctx)
	sc.cancelFunc = cancel

	sc.mu.Lock()
	sc.state = SplashRunning
	sc.mu.Unlock()

	go sc.runAnimation(runCtx)
	go sc.runTasks(runCtx)
}

func (sc *SplashController) runAnimation(ctx context.Context) {
	ticker := time.NewTicker(sc.cfg.TickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sc.mu.Lock()
			if sc.state == SplashDismissed {
				sc.mu.Unlock()
				return
			}
			sc.frame++
			sc.mu.Unlock()
		}
	}
}

func (sc *SplashController) runTasks(ctx context.Context) {
	total := len(sc.cfg.Tasks)
	if total == 0 {
		sc.finish(false)
		return
	}

	for i, task := range sc.cfg.Tasks {
		select {
		case <-ctx.Done():
			return
		default:
		}

		sc.mu.Lock()
		sc.pills[i].State = ProviderFetching
		sc.stepText = "fetching " + task.Name + "..."
		sc.mu.Unlock()

		var err error
		if task.Action != nil {
			err = task.Action(ctx)
		} else if task.Duration > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(task.Duration):
			}
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		sc.mu.Lock()
		if err != nil {
			sc.pills[i].State = ProviderFailed
		} else {
			sc.pills[i].State = ProviderDone
		}
		sc.stepText = task.Name + " done"
		sc.progress = float64(i+1) / float64(total) * 100.0
		sc.mu.Unlock()
	}

	if sc.cfg.HoldDuration > 0 {
		select {
		case <-ctx.Done():
			return
		case <-time.After(sc.cfg.HoldDuration):
		}
	}

	sc.finish(false)
}

func (sc *SplashController) finish(dismissed bool) {
	sc.once.Do(func() {
		sc.mu.Lock()
		if dismissed {
			sc.state = SplashDismissed
			for i := range sc.pills {
				if sc.pills[i].State == ProviderPending || sc.pills[i].State == ProviderFetching {
					sc.pills[i].State = ProviderSkipped
				}
			}
		} else {
			sc.state = SplashCompleted
		}
		sc.mu.Unlock()

		if sc.cancelFunc != nil {
			sc.cancelFunc()
		}
		close(sc.doneCh)
	})
}

// Dismiss immediately stops tasks and transitions to dismissed state.
func (sc *SplashController) Dismiss() {
	sc.finish(true)
}

// Done returns a channel that is closed when tasks finish or splash is dismissed.
func (sc *SplashController) Done() <-chan struct{} {
	return sc.doneCh
}

// Snapshot returns a copy of the current state snapshot.
func (sc *SplashController) Snapshot() SplashSnapshot {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	pillsCopy := make([]ProviderPill, len(sc.pills))
	copy(pillsCopy, sc.pills)

	return SplashSnapshot{
		State:        sc.state,
		SpinnerFrame: sc.frame,
		Progress:     sc.progress,
		StepText:     sc.stepText,
		Pills:        pillsCopy,
		Dismissed:    sc.state == SplashDismissed,
		Completed:    sc.state == SplashCompleted,
	}
}

// HandleKey intercepts navigation / dismissal keys (Esc, q, Enter, Ctrl-C).
func (sc *SplashController) HandleKey(e KeyEvent) bool {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	switch key {
	case "esc", "q", "enter", "ctrl-c", "ctrl-q":
		sc.Dismiss()
		return true
	}
	return false
}
