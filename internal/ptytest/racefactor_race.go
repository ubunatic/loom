// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build race

package ptytest

// RaceFactor scales process deadlines for race-instrumented PTY tests.
const RaceFactor = 4
