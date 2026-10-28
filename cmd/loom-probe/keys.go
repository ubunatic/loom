// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
	"ubunatic.com/loom"
	"ubunatic.com/loom/terminal"
)

// kittyQuery asks for the current kitty keyboard flags (CSI ? u) followed by
// primary device attributes (CSI c). Every terminal answers DA1, so the probe
// never waits for a kitty reply that a legacy terminal will not send.
const kittyQuery = "\x1b[?u\x1b[c"

const keysHint = "Try Ctrl-1..Ctrl-9, Ctrl-3, Ctrl-8, Esc, Alt-1, Ctrl-A, Enter, Shift-Enter, Tab, Backspace. Press q to end this phase."

// modifyOtherKeys2 enables xterm modifyOtherKeys level 2, which reports
// modified keys as CSI 27 ; <mod> ; <code> ~. terminal.Restore resets it.
const modifyOtherKeys2 = "\x1b[>4;2m"

type kittySupport struct {
	supported bool
	flags     int
	answered  bool // DA1 reply seen
}

// parseKittyQueryReply scans the reply to kittyQuery for "\x1b[?<flags>u" and
// the DA1 terminator "\x1b[?...c".
func parseKittyQueryReply(reply string) kittySupport {
	var s kittySupport
	for rest := reply; ; {
		i := strings.Index(rest, "\x1b[?")
		if i < 0 {
			return s
		}
		rest = rest[i+3:]
		end := strings.IndexAny(rest, "uc")
		if end < 0 {
			return s
		}
		body, final := rest[:end], rest[end]
		rest = rest[end+1:]
		if final == 'c' {
			s.answered = true
			continue
		}
		if flags, err := strconv.Atoi(body); err == nil {
			s.supported, s.flags = true, flags
		}
	}
}

func runKeys(input *os.File, output io.Writer) error {
	inFD := int(input.Fd())
	outFD, ok := output.(*os.File)
	if !ok || !term.IsTerminal(inFD) || !term.IsTerminal(int(outFD.Fd())) {
		return fmt.Errorf("stdin and stdout must be attached to a real terminal")
	}
	oldState, err := term.MakeRaw(inFD)
	if err != nil {
		return fmt.Errorf("enable raw terminal input: %w", err)
	}
	kittyPushed := false
	defer func() {
		_ = terminal.Restore(inFD, oldState, output, false, kittyPushed)
	}()

	say := func(format string, args ...any) {
		fmt.Fprintf(output, strings.ReplaceAll(format, "\n", "\r\n"), args...)
	}
	say("loom-probe keys: TERM=%q TERM_PROGRAM=%q TMUX=%t\n", os.Getenv("TERM"), os.Getenv("TERM_PROGRAM"), os.Getenv("TMUX") != "")

	if _, err := io.WriteString(output, kittyQuery); err != nil {
		return fmt.Errorf("write kitty query: %w", err)
	}
	var reply strings.Builder
	deadline := time.Now().Add(time.Second)
	support := kittySupport{}
	for !support.answered && time.Now().Before(deadline) {
		chunk, err := readChunk(inFD, time.Until(deadline))
		if err != nil {
			return fmt.Errorf("read kitty query reply: %w", err)
		}
		reply.Write(chunk)
		support = parseKittyQueryReply(reply.String())
	}
	switch {
	case !support.answered:
		say("Kitty query: no DA1 reply within 1s (reply %s)\n", strconv.Quote(reply.String()))
	case support.supported:
		say("Kitty query: supported, current flags %d\n", support.flags)
	default:
		say("Kitty query: not supported (DA1 only)\n")
	}

	say("\nPhase 1: legacy keys. %s\n", keysHint)
	if err := echoKeys(inFD, say); err != nil {
		return err
	}
	if !support.supported {
		say("\nPhase 2 skipped: terminal did not report the kitty keyboard protocol.\n")
	} else {
		if _, err := io.WriteString(output, "\x1b[>1u"); err != nil {
			return fmt.Errorf("push kitty flags: %w", err)
		}
		kittyPushed = true
		say("\nPhase 2: kitty flag 1 (disambiguate). %s\n", keysHint)
		if err := echoKeys(inFD, say); err != nil {
			return err
		}
		if _, err := io.WriteString(output, terminal.ResetKittyKeyboard); err != nil {
			return fmt.Errorf("pop kitty flags: %w", err)
		}
		kittyPushed = false
	}

	if _, err := io.WriteString(output, modifyOtherKeys2); err != nil {
		return fmt.Errorf("enable modifyOtherKeys: %w", err)
	}
	say("\nPhase 3: xterm modifyOtherKeys=2 (no query exists; unchanged output means unsupported). %s\n", keysHint)
	return echoKeys(inFD, say)
}

// echoKeys prints the raw bytes and loom's decoded name of each key until q.
func echoKeys(fd int, say func(string, ...any)) error {
	for {
		chunk, err := readChunk(fd, 0)
		if err != nil {
			return fmt.Errorf("read key: %w", err)
		}
		if string(chunk) == "q" {
			return nil
		}
		name := loom.DecodeKey(chunk).Name()
		if name == "" {
			name = "(undecoded)"
		}
		say("  %-24s %s\n", strconv.Quote(string(chunk)), name)
	}
}
