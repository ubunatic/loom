// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package measure

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

const (
	zwjSplit int32 = iota
	zwjJoin
)

var detectedZWJMode atomic.Int32

func init() {
	detectedZWJMode.Store(zwjSplit)
	if mode, ok := parseZWJMode(os.Getenv("LOOM_ZWJ")); ok {
		detectedZWJMode.Store(mode)
	}
}

// DetectZWJMode probes an interactive terminal after its caller has entered
// raw mode and selected its screen. It is safe to call with non-terminal files.
func DetectZWJMode(input, output *os.File, row int) {
	if mode, ok := parseZWJMode(os.Getenv("LOOM_ZWJ")); ok {
		detectedZWJMode.Store(mode)
		return
	}
	if mode, err := probeTerminalZWJMode(input, output, row, 200*time.Millisecond); err == nil {
		detectedZWJMode.Store(mode)
	}
}

// ZWJClusterWidth reports the visible width of a ZWJ sequence under the
// startup terminal probe, with LOOM_ZWJ=join|split taking precedence.
func ZWJClusterWidth(cluster string) int {
	mode := detectedZWJMode.Load()
	if override, ok := parseZWJMode(os.Getenv("LOOM_ZWJ")); ok {
		mode = override
	}
	if mode == zwjJoin {
		return ActiveEmojiSpec().ZWJDefaultWidth
	}
	width := 0
	for _, part := range strings.Split(cluster, "\u200D") {
		width += ClusterWidth(part)
	}
	return width
}

func parseZWJMode(value string) (int32, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "join":
		return zwjJoin, true
	case "split":
		return zwjSplit, true
	default:
		return zwjSplit, false
	}
}

func zwjModeFromAdvance(advance int) int32 {
	if advance == ActiveEmojiSpec().ZWJDefaultWidth {
		return zwjJoin
	}
	return zwjSplit
}

func probeTerminalZWJMode(input, output *os.File, row int, timeout time.Duration) (int32, error) {
	if input == nil || output == nil || !term.IsTerminal(int(input.Fd())) || !term.IsTerminal(int(output.Fd())) {
		return zwjSplit, fmt.Errorf("terminal probe requires tty stdin and stdout")
	}
	if row < 1 {
		row = 1
	}
	defer output.WriteString("\r\x1b[2K\x1b8") //nolint:errcheck
	if _, err := fmt.Fprintf(output, "\x1b7\x1b[%d;1H👨‍👩‍👧‍👦\x1b[6n", row); err != nil {
		return zwjSplit, err
	}
	_, col, err := readProbeCursor(int(input.Fd()), timeout)
	if err != nil {
		return zwjSplit, err
	}
	return zwjModeFromAdvance(col - 1), nil
}

func readProbeCursor(fd int, timeout time.Duration) (int, int, error) {
	deadline := time.Now().Add(timeout)
	var reply strings.Builder
	for reply.Len() < 64 {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return 0, 0, fmt.Errorf("terminal probe timed out")
		}
		fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		ms := int(remaining.Milliseconds())
		if ms < 1 {
			ms = 1
		}
		n, err := unix.Poll(fds, ms)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return 0, 0, err
		}
		if n == 0 {
			return 0, 0, fmt.Errorf("terminal probe timed out")
		}
		var one [1]byte
		count, err := unix.Read(fd, one[:])
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return 0, 0, err
		}
		if count == 0 {
			continue
		}
		reply.WriteByte(one[0])
		if one[0] == 'R' {
			var row, col int
			if _, err := fmt.Sscanf(reply.String(), "\x1b[%d;%dR", &row, &col); err != nil || row < 1 || col < 1 {
				return 0, 0, fmt.Errorf("invalid cursor report %q", reply.String())
			}
			return row, col, nil
		}
	}
	return 0, 0, fmt.Errorf("invalid cursor report %q", reply.String())
}
