// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package specdefaults

import "testing"

// TestMovedKeyValidatorsRejectMissingValues checks that each key moved out of Go
// fails validation when it is zero or empty, so a broken spec panics at startup
// instead of reaching a consumer.
func TestMovedKeyValidatorsRejectMissingValues(t *testing.T) {
	cases := []struct {
		name     string
		clear    func(*LibDefaults)
		validate func(LibDefaults) error
	}{
		{
			name:     "clock.tick_interval",
			clear:    func(d *LibDefaults) { d.Clock.TickInterval = 0 },
			validate: func(d LibDefaults) error { return d.Clock.Validate() },
		},
		{
			name:     "progress_bar.fallback_interval",
			clear:    func(d *LibDefaults) { d.ProgressBar.FallbackInterval = 0 },
			validate: func(d LibDefaults) error { return d.ProgressBar.Validate() },
		},
		{
			name:     "spinner.tick_interval",
			clear:    func(d *LibDefaults) { d.Spinner.TickInterval = 0 },
			validate: func(d LibDefaults) error { return d.Spinner.Validate() },
		},
		{
			name:     "key_help.separator",
			clear:    func(d *LibDefaults) { d.KeyHelp.Separator = "" },
			validate: func(d LibDefaults) error { return d.KeyHelp.Validate() },
		},
		{
			name:     "menu.checked_marker",
			clear:    func(d *LibDefaults) { d.Menu.CheckedMarker = "" },
			validate: func(d LibDefaults) error { return d.Menu.Validate() },
		},
		{
			name:     "menu.unchecked_marker",
			clear:    func(d *LibDefaults) { d.Menu.UncheckedMarker = "" },
			validate: func(d LibDefaults) error { return d.Menu.Validate() },
		},
		{
			name:     "media.play_label",
			clear:    func(d *LibDefaults) { d.Media.PlayLabel = "" },
			validate: func(d LibDefaults) error { return d.Media.Validate() },
		},
		{
			name:     "media.pause_label",
			clear:    func(d *LibDefaults) { d.Media.PauseLabel = "" },
			validate: func(d LibDefaults) error { return d.Media.Validate() },
		},
		{
			name:     "media.default_fps",
			clear:    func(d *LibDefaults) { d.Media.DefaultFPS = 0 },
			validate: func(d LibDefaults) error { return d.Media.Validate() },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.validate(SpeccedDefaults); err != nil {
				t.Fatalf("embedded spec rejected: %v", err)
			}
			d := SpeccedDefaults
			tc.clear(&d)
			if err := tc.validate(d); err == nil {
				t.Fatalf("%s accepted a missing value", tc.name)
			}
		})
	}
}
