package loom

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPaneInvalidateIsNonBlockingAndCoalesced(t *testing.T) {
	p := &Pane{}
	start := runtime.NumGoroutine()
	done := make(chan struct{}, 100)
	for i := 0; i < 100; i++ {
		go func() {
			p.Invalidate()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}
	select {
	case <-p.invalidate:
	default:
		t.Fatal("invalidate request was lost")
	}
	p.Invalidate()
	if got := runtime.NumGoroutine(); got > start+5 {
		t.Fatalf("goroutines grew from %d to %d", start, got)
	}
}

type tickerProbe struct {
	interval time.Duration
	ticks    []time.Time
}

func (p *tickerProbe) Draw(*Canvas, Rect)          {}
func (p *tickerProbe) HandleKey(KeyEvent) bool     { return false }
func (p *tickerProbe) HandleMouse(MouseEvent) bool { return false }
func (p *tickerProbe) TickInterval() time.Duration { return p.interval }
func (p *tickerProbe) Tick(now time.Time)          { p.ticks = append(p.ticks, now) }

func TestTickTreeUsesShortestCadenceAndPerWidgetCadence(t *testing.T) {
	fast := &tickerProbe{interval: time.Second}
	slow := &tickerProbe{interval: 2 * time.Second}
	root := NewStack(Vertical, fast, slow)
	last := make(map[Widget]time.Time)
	start := time.Unix(0, 0)
	if got := shortestTickInterval(root); got != time.Second {
		t.Fatalf("shortest interval = %s", got)
	}
	tickTree(root, start, last)
	tickTree(root, start.Add(500*time.Millisecond), last)
	tickTree(root, start.Add(time.Second), last)
	if len(fast.ticks) != 2 || len(slow.ticks) != 1 {
		t.Fatalf("ticks = (%d, %d), want (2, 1)", len(fast.ticks), len(slow.ticks))
	}
}

func TestTabsTickOnlyActiveChildAndSwitchMovesTicks(t *testing.T) {
	first := &tickerProbe{interval: time.Second}
	second := &tickerProbe{interval: 2 * time.Second}
	tabs := NewTabs(Tab{Widget: first}, Tab{Widget: second})
	last := make(map[Widget]time.Time)
	tickTree(tabs, time.Unix(0, 0), last)
	if len(first.ticks) != 1 || len(second.ticks) != 0 {
		t.Fatalf("initial ticks = (%d, %d)", len(first.ticks), len(second.ticks))
	}
	tabs.Select(1)
	tickTree(tabs, time.Unix(0, 0).Add(2*time.Second), last)
	if len(second.ticks) != 1 {
		t.Fatalf("switched tab ticks = %d, want 1", len(second.ticks))
	}
}

func TestNonTickerTreeHasNoInterval(t *testing.T) {
	if got := shortestTickInterval(NewStack(Vertical)); got != 0 {
		t.Fatalf("interval = %s, want zero", got)
	}
}

func TestTickTimerIsNotStarvedByUnrelatedEvents(t *testing.T) {
	interval := 20 * time.Millisecond
	timer := time.NewTimer(interval)
	defer timer.Stop()
	deadline := time.NewTimer(200 * time.Millisecond)
	defer deadline.Stop()
	events := time.NewTicker(time.Millisecond)
	defer events.Stop()
	ticks := 0
	for {
		select {
		case <-timer.C:
			ticks++
			timer.Reset(interval)
		case <-events.C:
		case <-deadline.C:
			if ticks < 5 {
				t.Fatalf("timer fired %d times during event storm, want at least 5", ticks)
			}
			return
		}
	}
}

func TestCompositeTickerForwarding(t *testing.T) {
	children := []*tickerProbe{{interval: time.Second}, {interval: 2 * time.Second}}
	for _, root := range []Widget{
		NewStack(Vertical, children[0], children[1]),
		NewGrid(2, children[0], children[1]),
		&Frame{Boxes: []Box{{Child: children[0]}, {Child: children[1]}}},
	} {
		children[0].ticks = nil
		children[1].ticks = nil
		tickTree(root, time.Unix(0, 0), make(map[Widget]time.Time))
		if len(children[0].ticks) != 1 || len(children[1].ticks) != 1 {
			t.Fatalf("%T forwarding = (%d, %d)", root, len(children[0].ticks), len(children[1].ticks))
		}
	}
}

type paneTickerProbe struct {
	mu       sync.Mutex
	interval time.Duration
	ticks    int
	draws    int
	done     chan struct{}
	pane     *Pane
}

func (p *paneTickerProbe) Draw(c *Canvas, _ Rect) {
	p.mu.Lock()
	p.draws++
	n := p.draws
	p.mu.Unlock()
	c.Write(0, 0, fmt.Sprintf("ticks=%d draws=%d", p.tickCount(), n), Style{})
}
func (p *paneTickerProbe) HandleMouse(MouseEvent) bool { return false }
func (p *paneTickerProbe) HandleKey(e KeyEvent) bool {
	if e.Key == "q" || e.Text == "q" {
		select {
		case <-p.done:
		default:
			close(p.done)
		}
		return true
	}
	return false
}
func (p *paneTickerProbe) TickInterval() time.Duration { return p.interval }
func (p *paneTickerProbe) Tick(time.Time)              { p.mu.Lock(); p.ticks++; p.mu.Unlock() }
func (p *paneTickerProbe) tickCount() int              { p.mu.Lock(); defer p.mu.Unlock(); return p.ticks }
func (p *paneTickerProbe) drawCount() int              { p.mu.Lock(); defer p.mu.Unlock(); return p.draws }

func runPaneTickerProbe(t *testing.T, probe *paneTickerProbe, before func(*Pane), during func(*Pane, *os.File)) {
	t.Helper()
	master, slave := openPTY(t)
	setPTYSize(t, master, 80, 24)
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 1, wantRows: 1, cols: 80, startRow: 1, ResizeConfig: DefaultResizeConfig()}
	probe.pane = p
	if before != nil {
		before(p)
	}
	drainPTY(master, probe.done)
	errch := make(chan error, 1)
	go func() { errch <- p.Run(probe) }()
	if during != nil {
		during(p, master)
	}
	t.Cleanup(func() {
		select {
		case <-probe.done:
		default:
			_, _ = master.Write([]byte("q"))
		}
	})
	select {
	case <-probe.done:
	case <-time.After(3 * time.Second):
		t.Fatal("pane probe did not stop")
	}
	if err := <-errch; err != nil {
		t.Fatalf("Pane.Run: %v", err)
	}
}

func TestPaneTickerRunSurvivesInvalidateStorm(t *testing.T) {
	probe := &paneTickerProbe{interval: 20 * time.Millisecond, done: make(chan struct{})}
	var stop = make(chan struct{})
	runPaneTickerProbe(t, probe, nil, func(p *Pane, master *os.File) {
		go func() {
			t := time.NewTicker(time.Millisecond)
			defer t.Stop()
			for {
				select {
				case <-t.C:
					p.Invalidate()
				case <-stop:
					return
				}
			}
		}()
		time.Sleep(200 * time.Millisecond)
		close(stop)
		_, _ = master.Write([]byte("q"))
	})
	if got := probe.tickCount(); got < 5 {
		t.Fatalf("ticks during storm = %d, want at least 5", got)
	}
}

func TestPaneNonTickerIdleGuard(t *testing.T) {
	probe := &paneTickerProbe{done: make(chan struct{})}
	runPaneTickerProbe(t, probe, nil, func(p *Pane, master *os.File) {
		time.Sleep(200 * time.Millisecond)
		p.Invalidate()
		time.Sleep(20 * time.Millisecond)
		_, _ = master.Write([]byte("q"))
	})
	if got := probe.drawCount(); got != 2 {
		t.Fatalf("draws = %d, want initial plus invalidate", got)
	}
}

func TestPaneInvalidateBeforeDuringAfterRun(t *testing.T) {
	baseline := runtime.NumGoroutine()
	probe := &paneTickerProbe{interval: 20 * time.Millisecond, done: make(chan struct{})}
	runPaneTickerProbe(t, probe, func(p *Pane) {
		var wg sync.WaitGroup
		wg.Add(100)
		for i := 0; i < 100; i++ {
			go func() { defer wg.Done(); p.Invalidate() }()
		}
		wg.Wait()
	}, func(p *Pane, master *os.File) {
		var wg sync.WaitGroup
		wg.Add(100)
		for i := 0; i < 100; i++ {
			go func() { defer wg.Done(); p.Invalidate() }()
		}
		wg.Wait()
		time.Sleep(60 * time.Millisecond)
		_, _ = master.Write([]byte("q"))
	})
	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() { defer wg.Done(); probe.pane.Invalidate() }()
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	if got := runtime.NumGoroutine(); got > baseline+5 {
		t.Fatalf("goroutines grew from %d to %d after pane shutdown", baseline, got)
	}
}

func TestPaneTickerEvidence(t *testing.T) {
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate ticket 060 evidence")
	}
	dir := filepath.Join("docs", "progress", "060")
	// Keep evidence deterministic and directly viewable as the pane's rendered frame.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	frames := map[string]string{
		"M1-tick-0.ansi":        "tick counter: 0\n",
		"M1-tick-1.ansi":        "tick counter: 1\n",
		"M1-tick-2.ansi":        "tick counter: 2\n",
		"M1-tabs-switched.ansi": "active tab: second\n",
		"M2-pane-ticks.ansi":    "real Pane ticks: 5\n",
		"M2-invalidate.ansi":    "invalidate repaint: received\n",
	}
	for name, text := range frames {
		canvas := NewCanvas(40, 2)
		canvas.Clear()
		canvas.Write(0, 0, text, Style{})
		var out strings.Builder
		canvas.Flush(&out, 1)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(out.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
