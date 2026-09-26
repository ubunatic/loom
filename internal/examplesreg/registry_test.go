// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package examplesreg

import (
	"reflect"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom"
)

func TestInteractiveExamplesHaveDemoArgs(t *testing.T) {
	want := map[string][]string{
		"monitor": []string{"--watch"},
		"splash":  []string{"--watch"},
		"treemap": []string{"--watch", "--ansi"},
	}
	for _, name := range []string{"monitor", "splash", "treemap"} {
		example, ok := Find(name)
		if !ok {
			t.Fatalf("Find(%q) returned no example", name)
		}
		if !reflect.DeepEqual(example.DemoArgs, want[name]) {
			t.Errorf("Find(%q).DemoArgs = %v, want %v", name, example.DemoArgs, want[name])
		}
	}
}

func TestANSIViewerIsRegistered(t *testing.T) {
	e, ok := Find("ansiviewer")
	if !ok {
		t.Fatal("ansiviewer is not registered")
	}
	if e.Package != "codeberg.org/ubunatic/loom/examples/ansiviewer" || !e.SupportsHelp {
		t.Fatalf("registration = %#v", e)
	}
}

func TestNewWidgetConverted(t *testing.T) {
	examples := []string{"split", "tabs", "filebrowser", "splash", "monitor", "treemap"}
	for _, name := range examples {
		example, ok := Find(name)
		if !ok {
			t.Fatalf("Find(%q) returned no example", name)
		}
		if example.NewWidget == nil {
			t.Errorf("Find(%q).NewWidget = nil, expected non-nil", name)
		}
	}
}

type countedTicker struct {
	loom.Widget
	ticks int
}

func (c *countedTicker) TickInterval() time.Duration {
	if ticker, ok := c.Widget.(loom.Ticker); ok {
		return ticker.TickInterval()
	}
	return 0
}
func (c *countedTicker) Tick(now time.Time) {
	c.ticks++
	if ticker, ok := c.Widget.(loom.Ticker); ok {
		ticker.Tick(now)
	}
}
func (c *countedTicker) Focused() bool {
	if f, ok := c.Widget.(loom.Focusable); ok {
		return f.Focused()
	}
	return false
}
func (c *countedTicker) SetFocus(focused bool) {
	if f, ok := c.Widget.(loom.Focusable); ok {
		f.SetFocus(focused)
	}
}

func TestHostedSplashAndMonitorTabsTickOnlyActiveTab(t *testing.T) {
	makeWidget := func(name string) loom.Widget {
		t.Helper()
		example, ok := Find(name)
		if !ok || example.NewWidget == nil {
			t.Fatalf("%s widget factory missing", name)
		}
		value, err := example.NewWidget([]string{"--watch"})
		if err != nil {
			t.Fatalf("%s factory: %v", name, err)
		}
		widget, ok := value.(loom.Widget)
		if !ok {
			t.Fatalf("%s factory returned %T, not loom.Widget", name, value)
		}
		return widget
	}
	splashWidget := makeWidget("splash")
	monitorWidget := makeWidget("monitor")
	defer func() {
		if c, ok := splashWidget.(interface{ Close() }); ok {
			c.Close()
		}
		if c, ok := monitorWidget.(interface{ Close() }); ok {
			c.Close()
		}
	}()
	splash := &countedTicker{Widget: splashWidget}
	monitor := &countedTicker{Widget: monitorWidget}
	tabs := loom.NewTabs(loom.Tab{Title: "Splash", Widget: splash}, loom.Tab{Title: "Monitor", Widget: monitor})
	tabs.Tick(time.Now())
	if splash.ticks != 1 || monitor.ticks != 0 {
		t.Fatalf("initial tab ticks = splash:%d monitor:%d, want 1:0", splash.ticks, monitor.ticks)
	}
	tabs.Select(1)
	tabs.Tick(time.Now())
	if splash.ticks != 1 || monitor.ticks != 1 {
		t.Fatalf("after selecting monitor ticks = splash:%d monitor:%d, want 1:1", splash.ticks, monitor.ticks)
	}
}
