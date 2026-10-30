// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
	"ubunatic.com/loom"
	"ubunatic.com/loom/terminal"
)

// defaultModsPath is the state file written by voxi-modifierd.
const defaultModsPath = "/run/voxi/modifiers"

// modMask is the one-byte physical modifier state from the state file.
type modMask uint8

// modNames follows the bit order of the voxi-modifierd file contract.
var modNames = [8]string{"lctrl", "rctrl", "lalt", "ralt", "lsuper", "rsuper", "lshift", "rshift"}

var errModsUnavailable = errors.New("modifier state not available")

// Names returns the names of the set bits in bit order.
func (m modMask) Names() []string {
	var names []string
	for bit, name := range modNames {
		if m&(1<<bit) != 0 {
			names = append(names, name)
		}
	}
	return names
}

// readModMask opens path fresh on each call, because the daemon replaces the
// file by rename, and returns its first byte. Extra bytes are ignored.
func readModMask(path string) (modMask, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", errModsUnavailable, err)
	}
	defer f.Close()
	var buf [1]byte
	if _, err := io.ReadFull(f, buf[:]); err != nil {
		if errors.Is(err, io.EOF) {
			return 0, fmt.Errorf("%s: empty modifier state file", path)
		}
		return 0, fmt.Errorf("read %s: %w", path, err)
	}
	return modMask(buf[0]), nil
}

// formatModsLine renders one key with the modifier mask sampled on arrival.
func formatModsLine(chunk []byte, m modMask, err error) string {
	name := loom.DecodeKey(chunk).Name()
	if name == "" {
		name = "(undecoded)"
	}
	line := fmt.Sprintf("  %-24s %-16s ", strconv.Quote(string(chunk)), name)
	if err != nil {
		return line + "mask=n/a"
	}
	names := "-"
	if n := m.Names(); len(n) > 0 {
		names = strings.Join(n, " ")
	}
	return line + fmt.Sprintf("mask=0x%02x %s", uint8(m), names)
}

func runMods(input *os.File, output io.Writer, path string) error {
	inFD := int(input.Fd())
	outFD, ok := output.(*os.File)
	if !ok || !term.IsTerminal(inFD) || !term.IsTerminal(int(outFD.Fd())) {
		return fmt.Errorf("stdin and stdout must be attached to a real terminal")
	}
	oldState, err := term.MakeRaw(inFD)
	if err != nil {
		return fmt.Errorf("enable raw terminal input: %w", err)
	}
	defer func() {
		_ = terminal.Restore(inFD, oldState, output, false)
	}()

	say := func(format string, args ...any) {
		fmt.Fprintf(output, strings.ReplaceAll(format, "\n", "\r\n"), args...)
	}
	say("loom-probe mods: TERM=%q TERM_PROGRAM=%q path=%q\n", os.Getenv("TERM"), os.Getenv("TERM_PROGRAM"), path)
	if _, err := readModMask(path); err != nil {
		say("voxi modifier state not available (%v); showing plain key output\n", err)
	}
	say("%s\n", keysHint)
	for {
		chunk, err := readChunk(inFD, 0)
		if err != nil {
			return fmt.Errorf("read key: %w", err)
		}
		if string(chunk) == "q" {
			return nil
		}
		m, modErr := readModMask(path)
		say("%s\n", formatModsLine(chunk, m, modErr))
	}
}
