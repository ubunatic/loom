// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "sort"

// KeyMap maps key names to logical action names.
//
// Key names are matched without regard to ASCII letter case. Non-ASCII key
// names are compared exactly.
type KeyMap struct {
	actions map[string][]string
	labels  map[string]string
}

// NewKeyMap creates a key map from action names to their key aliases.
func NewKeyMap(bindings map[string][]string) *KeyMap {
	return NewKeyMapWithLabels(bindings, nil)
}

// NewKeyMapWithLabels creates a key map and assigns optional short help labels
// to actions. Bindings and labels are copied; later caller changes do not
// affect the map.
func NewKeyMapWithLabels(bindings map[string][]string, labels map[string]string) *KeyMap {
	km := &KeyMap{
		actions: make(map[string][]string, len(bindings)),
		labels:  make(map[string]string, len(labels)),
	}
	for action, keys := range bindings {
		km.actions[action] = append([]string(nil), keys...)
	}
	for action, label := range labels {
		km.labels[action] = label
	}
	return km
}

// Action returns the action bound to e, or an empty string if none matches.
// If a key is bound to multiple actions, the lexicographically first action
// name is returned.
func (km *KeyMap) Action(e KeyEvent) string {
	if km == nil {
		return ""
	}
	name := e.Name()
	if name == "" {
		return ""
	}
	actions := make([]string, 0, len(km.actions))
	for action := range km.actions {
		actions = append(actions, action)
	}
	sort.Strings(actions)
	for _, action := range actions {
		for _, key := range km.actions[action] {
			if equalKeyName(name, key) {
				return action
			}
		}
	}
	return ""
}

// Matches reports whether e is bound to action.
func (km *KeyMap) Matches(e KeyEvent, action string) bool {
	return action != "" && km.Action(e) == action
}

// Label returns the short help label for action, or an empty string when none
// was configured.
func (km *KeyMap) Label(action string) string {
	if km == nil {
		return ""
	}
	return km.labels[action]
}

func equalKeyName(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ac, bc := a[i], b[i]
		if ac >= 'A' && ac <= 'Z' {
			ac += 'a' - 'A'
		}
		if bc >= 'A' && bc <= 'Z' {
			bc += 'a' - 'A'
		}
		if ac != bc {
			return false
		}
	}
	return true
}
