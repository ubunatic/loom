// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"fmt"

	"ubunatic.com/loom/spec"
)

var graphDefaults = func() spec.GraphDefaults {
	defaults, err := spec.LoadGraphDefaults()
	if err != nil {
		panic(fmt.Sprintf("graph: load spec/defaults.yaml: %v", err))
	}
	return defaults
}()
