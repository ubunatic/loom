// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	_ "embed"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed spec/resize.yaml
var resizeSpecYAML []byte

// ResizeModeSpec defines the metadata and default configuration for a single resize mode.
type ResizeModeSpec struct {
	ID             string `yaml:"-"`
	Title          string `yaml:"title"`
	Description    string `yaml:"description"`
	Default        bool   `yaml:"default"`
	DiagnosticOnly bool   `yaml:"diagnostic_only"`
	Key            string `yaml:"key"`
	GuardN         int    `yaml:"guard_n,omitempty"`
	// Adaptive guard policy (see WinchMeter); only set on adaptive_guard.
	WindowMS int `yaml:"window_ms,omitempty"`
	MinN     int `yaml:"min_n,omitempty"`
	MaxN     int `yaml:"max_n,omitempty"`
	RateStep int `yaml:"rate_step,omitempty"`
}

// AutoFullscreenSpec declares quasi-fullscreen detection (see QuasiFullscreen).
type AutoFullscreenSpec struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Default     bool   `yaml:"default"`
	Key         string `yaml:"key"`
	MarginRows  int    `yaml:"margin_rows"`
	MinPercent  int    `yaml:"min_percent"`
	Alt         bool   `yaml:"alt"`
}

// ResizeModesSpec holds all resize modes declared in spec/resize.yaml.
type ResizeModesSpec struct {
	Modes          map[string]ResizeModeSpec `yaml:"modes"`
	AutoFullscreen AutoFullscreenSpec        `yaml:"auto_fullscreen"`
}

// SpeccedResizeModeIDs lists all mode identifiers in fixed presentation order.
var SpeccedResizeModeIDs = []string{
	"coalesce",
	"atomic_flush",
	"row_clear",
	"synchronized_output",
	"auto_wrap",
	"out_of_band_clear",
	"full_screen_buffer",
	"resize_handling",
	"width_guard",
	"adaptive_guard",
	"alt_screen",
}

// SpeccedResizeModes is the loaded immutable spec of resize modes.
var SpeccedResizeModes = func() ResizeModesSpec {
	var spec ResizeModesSpec
	if err := yaml.Unmarshal(resizeSpecYAML, &spec); err != nil {
		panic(fmt.Sprintf("loom: parse spec/resize.yaml: %v", err))
	}
	for id, m := range spec.Modes {
		m.ID = id
		spec.Modes[id] = m
	}
	if a := spec.Modes["adaptive_guard"]; a.MinN > a.MaxN || a.WindowMS <= 0 || a.RateStep <= 0 {
		panic("loom: spec/resize.yaml adaptive_guard needs window_ms > 0, rate_step > 0 and min_n <= max_n")
	}
	return spec
}()

// ResizeConfig holds the active boolean flags controlling resize and flush behaviors.
type ResizeConfig struct {
	Coalesce           bool
	AtomicFlush        bool
	RowClear           bool
	SynchronizedOutput bool
	AutoWrap           bool
	OutOfBandClear     bool
	FullScreenBuffer   bool
	ResizeHandling     bool
	WidthGuard         bool
	WidthGuardN        int
	AdaptiveGuard      bool
	AltScreen          bool

	// Quasi-fullscreen detection, see QuasiFullscreen.
	AutoFullscreen bool // switch to the full-screen layout when the pane is nearly full height
	FullMarginRows int  // full when terminal rows - wanted rows <= this
	FullMinPercent int  // also full at this percent of the terminal height; 0 = off
	FullAlt        bool // use the alternate screen for the automatic full-screen layout
}

// QuasiFullscreen reports whether a pane that wants wantRows rows on a
// termRows-row terminal is close enough to full height to be treated as full
// screen: within FullMarginRows of the terminal height, or at least
// FullMinPercent percent of it. It is pure so the policy can be unit-tested.
func (c ResizeConfig) QuasiFullscreen(wantRows, termRows int) bool {
	if !c.AutoFullscreen || termRows < 1 {
		return false
	}
	if termRows-wantRows <= c.FullMarginRows {
		return true
	}
	return c.FullMinPercent > 0 && wantRows*100 >= termRows*c.FullMinPercent
}

// DefaultResizeConfig returns a ResizeConfig populated with the spec-defined defaults.
func DefaultResizeConfig() ResizeConfig {
	guardN := SpeccedResizeModes.Modes["width_guard"].GuardN
	if guardN <= 0 {
		guardN = 1
	}
	return ResizeConfig{
		Coalesce:           SpeccedResizeModes.Modes["coalesce"].Default,
		AtomicFlush:        SpeccedResizeModes.Modes["atomic_flush"].Default,
		RowClear:           SpeccedResizeModes.Modes["row_clear"].Default,
		SynchronizedOutput: SpeccedResizeModes.Modes["synchronized_output"].Default,
		AutoWrap:           SpeccedResizeModes.Modes["auto_wrap"].Default,
		OutOfBandClear:     SpeccedResizeModes.Modes["out_of_band_clear"].Default,
		FullScreenBuffer:   SpeccedResizeModes.Modes["full_screen_buffer"].Default,
		ResizeHandling:     SpeccedResizeModes.Modes["resize_handling"].Default,
		WidthGuard:         SpeccedResizeModes.Modes["width_guard"].Default,
		WidthGuardN:        guardN,
		AdaptiveGuard:      SpeccedResizeModes.Modes["adaptive_guard"].Default,
		AltScreen:          SpeccedResizeModes.Modes["alt_screen"].Default,
		AutoFullscreen:     SpeccedResizeModes.AutoFullscreen.Default,
		FullMarginRows:     SpeccedResizeModes.AutoFullscreen.MarginRows,
		FullMinPercent:     SpeccedResizeModes.AutoFullscreen.MinPercent,
		FullAlt:            SpeccedResizeModes.AutoFullscreen.Alt,
	}
}

// Get returns the state of the named mode, and true if the mode identifier exists.
func (c *ResizeConfig) Get(id string) (bool, bool) {
	switch id {
	case "coalesce":
		return c.Coalesce, true
	case "atomic_flush":
		return c.AtomicFlush, true
	case "row_clear":
		return c.RowClear, true
	case "synchronized_output":
		return c.SynchronizedOutput, true
	case "auto_wrap":
		return c.AutoWrap, true
	case "out_of_band_clear":
		return c.OutOfBandClear, true
	case "full_screen_buffer":
		return c.FullScreenBuffer, true
	case "resize_handling":
		return c.ResizeHandling, true
	case "width_guard":
		return c.WidthGuard, true
	case "adaptive_guard":
		return c.AdaptiveGuard, true
	case "alt_screen":
		return c.AltScreen, true
	default:
		return false, false
	}
}

// Set updates the state of the named mode, returning true if the mode identifier was valid.
func (c *ResizeConfig) Set(id string, val bool) bool {
	switch id {
	case "coalesce":
		c.Coalesce = val
		return true
	case "atomic_flush":
		c.AtomicFlush = val
		return true
	case "row_clear":
		c.RowClear = val
		return true
	case "synchronized_output":
		c.SynchronizedOutput = val
		return true
	case "auto_wrap":
		c.AutoWrap = val
		return true
	case "out_of_band_clear":
		c.OutOfBandClear = val
		return true
	case "full_screen_buffer":
		c.FullScreenBuffer = val
		return true
	case "resize_handling":
		c.ResizeHandling = val
		return true
	case "width_guard":
		c.WidthGuard = val
		return true
	case "adaptive_guard":
		c.AdaptiveGuard = val
		return true
	case "alt_screen":
		c.AltScreen = val
		return true
	default:
		return false
	}
}

// Toggle inverts the named mode, returning true if the mode identifier was valid.
func (c *ResizeConfig) Toggle(id string) bool {
	if val, ok := c.Get(id); ok {
		c.Set(id, !val)
		return true
	}
	return false
}

// Reset restores all modes to their spec-defined defaults.
func (c *ResizeConfig) Reset() {
	*c = DefaultResizeConfig()
}

// IsDiagnosticOnly reports whether a mode is designated for diagnostic comparison only.
func (c *ResizeConfig) IsDiagnosticOnly(id string) bool {
	if mode, ok := SpeccedResizeModes.Modes[id]; ok {
		return mode.DiagnosticOnly
	}
	return false
}

// SetResizeMode updates a mode on the Pane's ResizeConfig.
func (p *Pane) SetResizeMode(id string, val bool) bool {
	return p.ResizeConfig.Set(id, val)
}

// ToggleResizeMode inverts a mode on the Pane's ResizeConfig.
func (p *Pane) ToggleResizeMode(id string) bool {
	return p.ResizeConfig.Toggle(id)
}

// ResetResizeModes restores the Pane's ResizeConfig to spec defaults.
func (p *Pane) ResetResizeModes() {
	p.ResizeConfig.Reset()
}

// WidthGuardActive reports whether the pane is currently within an active
// SIGWINCH burst and rendering with the width guard deduction applied.
func (p *Pane) WidthGuardActive() bool {
	return p.widthGuardActive
}

// SetWidthGuardN updates the number of columns deducted during an active resize burst.
func (p *Pane) SetWidthGuardN(n int) {
	if n < 1 {
		n = 1
	}
	p.ResizeConfig.WidthGuardN = n
}

// WinchMeter measures SIGWINCH arrival rate over a sliding window. The window
// is also the smoothing: the rate is the event count in the window divided by
// the window length. It takes explicit timestamps so tests need no clock.
type WinchMeter struct {
	stamps []time.Time
}

func winchWindow() time.Duration {
	return time.Duration(SpeccedResizeModes.Modes["adaptive_guard"].WindowMS) * time.Millisecond
}

// Record notes one SIGWINCH received at now.
func (m *WinchMeter) Record(now time.Time) {
	m.prune(now)
	m.stamps = append(m.stamps, now)
}

func (m *WinchMeter) prune(now time.Time) {
	cut := now.Add(-winchWindow())
	i := 0
	for i < len(m.stamps) && m.stamps[i].Before(cut) {
		i++
	}
	m.stamps = m.stamps[i:]
}

// Events returns the number of events inside the window ending at now.
func (m *WinchMeter) Events(now time.Time) int {
	m.prune(now)
	return len(m.stamps)
}

// Rate returns events per second over the span between the first and last
// event in the window ending at now, so a fast burst is measured at once
// instead of ramping up as the window fills. Fewer than two events give 0.
func (m *WinchMeter) Rate(now time.Time) float64 {
	n := m.Events(now)
	if n < 2 {
		return 0
	}
	span := m.stamps[n-1].Sub(m.stamps[0])
	if span < 10*time.Millisecond {
		span = 10 * time.Millisecond
	}
	return float64(n-1) / span.Seconds()
}

// AdaptiveGuardN maps a measured rate (with its event count) to a guard
// width using the spec policy: min_n + floor(rate/rate_step), clamped to
// max_n. With fewer than two events there is no measurable speed, so it
// returns manualN.
func AdaptiveGuardN(rate float64, events, manualN int) int {
	if events < 2 {
		return manualN
	}
	a := SpeccedResizeModes.Modes["adaptive_guard"]
	return min(a.MaxN, a.MinN+int(rate)/a.RateStep)
}

// EffectiveGuardN is the guard width in force: the width latched at the last
// SIGWINCH when adaptive mode is enabled and measurable, otherwise the manual
// WidthGuardN.
func (p *Pane) EffectiveGuardN() int {
	if p.ResizeConfig.AdaptiveGuard && p.adaptiveN > 0 {
		return p.adaptiveN
	}
	return max(1, p.ResizeConfig.WidthGuardN)
}

// WinchRate returns the measured SIGWINCH rate in events per second.
func (p *Pane) WinchRate() float64 { return p.winchMeter.Rate(time.Now()) }

// guardedCols applies the width guard (when active) to the full column count.
func (p *Pane) guardedCols(full int) int {
	if p.MaxCols > 0 && full > p.MaxCols {
		full = p.MaxCols
	}
	if p.ResizeConfig.WidthGuard && p.widthGuardActive {
		full = max(1, full-p.EffectiveGuardN())
	}
	return full
}
