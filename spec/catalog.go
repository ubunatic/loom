// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package spec embeds Loom's application specifications.
package spec

import "embed"

//go:embed widgets.yaml
var widgetFiles embed.FS

// WidgetsYAML returns the authoritative widget catalog document.
func WidgetsYAML() ([]byte, error) { return widgetFiles.ReadFile("widgets.yaml") }
