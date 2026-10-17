// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package loomtest provides a deterministic virtual terminal for testing ANSI
// output without a PTY, subprocess, terminal emulator or timing dependency.
// It is a stable public package; see docs/APIStability.md.
//
// Feed output to NewVT's io.Writer implementation, then inspect Screen, Cells,
// Cursor, CursorVisible, Pen and AltScreen. Coordinates are zero-based; Cells
// and FrameCells return independent snapshots. Dimensions are clamped to at
// least one cell. A VT is not safe for concurrent use.
//
// Supported output includes UTF-8 text, CR/LF, backspace, tab, cursor positioning,
// line/display erasure, SGR attributes and indexed/truecolor, cursor save/restore,
// and DEC modes 7 (wrap), 25 (cursor), 1049 (alternate screen) and 2026
// (synchronized frames). OSC and DCS payloads are ignored through BEL/ST.
// Unsupported sequences are ignored; this is not a complete xterm emulator
// (scroll regions, insert/delete operations and terminal queries are unsupported).
// Width follows loom/measure's cell policy; pin LOOM_ZWJ=join when testing joined
// emoji independently of the caller's terminal configuration.
package loomtest
