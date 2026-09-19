// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	_ "embed"
	"fmt"

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
}

// ResizeModesSpec holds all resize modes declared in spec/resize.yaml.
type ResizeModesSpec struct {
	Modes map[string]ResizeModeSpec `yaml:"modes"`
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
