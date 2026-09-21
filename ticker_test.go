package loom

import (
	"runtime"
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
