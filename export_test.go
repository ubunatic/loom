// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// SetHelpRunner overrides the runner used by :help for testing.
// It returns a restore function that resets the runner to its previous value.
func SetHelpRunner(fn func(w Widget, height int) error) func() {
	helpMu.Lock()
	prev := helpRunner
	helpRunner = fn
	helpMu.Unlock()
	return func() {
		helpMu.Lock()
		helpRunner = prev
		helpMu.Unlock()
	}
}
