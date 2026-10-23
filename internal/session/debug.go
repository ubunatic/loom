// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later
//go:build !windows

package session

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
	"unicode/utf8"
)

const debugRingSize = 64 * 1024
const debugControlLimit = 1024
const debugStallTimeout = time.Second
const debugRepairCooldown = 10 * time.Second
const debugBundleLimit = 5
const debugCaptureDelay = 150 * time.Millisecond

// debugChunk records a raw PTY read in the rolling stream.
type debugChunk struct {
	Offset int64
	Data   []byte
}

// debugRing retains the newest debugRingSize bytes and their read boundaries.
type debugRing struct {
	data  []byte
	start int
	size  int
	base  int64
	end   int64
	reads []debugChunk
}

func newDebugRing() *debugRing { return &debugRing{data: make([]byte, debugRingSize)} }

func (r *debugRing) Write(p []byte) {
	total := len(p)
	offset := r.end
	if len(p) > len(r.data) {
		p = p[len(p)-len(r.data):]
		offset = r.end + int64(total-len(p))
	}
	for _, b := range p {
		if r.size == len(r.data) {
			r.start = (r.start + 1) % len(r.data)
			r.base++
		} else {
			r.size++
		}
		r.data[(r.start+r.size-1)%len(r.data)] = b
	}
	chunk := debugChunk{Offset: offset, Data: append([]byte(nil), p...)}
	r.reads = append(r.reads, chunk)
	r.end += int64(total)
	if r.end-int64(r.size) > r.base {
		r.base = r.end - int64(r.size)
	}
	if r.size == len(r.data) && r.base > 0 {
		kept := r.reads[:0]
		for _, read := range r.reads {
			if read.Offset+int64(len(read.Data)) <= r.base {
				continue
			}
			if read.Offset < r.base {
				drop := int(r.base - read.Offset)
				read.Data = read.Data[drop:]
				read.Offset = r.base
			}
			kept = append(kept, read)
		}
		r.reads = kept
	}
}

func (r *debugRing) Bytes() []byte {
	out := make([]byte, r.size)
	for i := range out {
		out[i] = r.data[(r.start+i)%len(r.data)]
	}
	return out
}

func (r *debugRing) Chunks() []debugChunk {
	out := make([]debugChunk, len(r.reads))
	for i, read := range r.reads {
		out[i] = debugChunk{Offset: read.Offset, Data: append([]byte(nil), read.Data...)}
	}
	return out
}

func (r *debugRing) Snapshot() *debugRing {
	copyRing := &debugRing{data: r.Bytes(), size: r.size, base: r.base, end: r.end, reads: r.Chunks()}
	return copyRing
}

type debugAnomaly string

const (
	debugIncomplete debugAnomaly = "incomplete control sequence"
	debugOverlong   debugAnomaly = "overlong control sequence"
	debugLeak       debugAnomaly = "suspicious rendered control fragment"
)

// debugDetector observes PTY bytes without altering the stream passed to VT.
type debugDetector struct {
	pending   []byte
	pendingAt time.Time
}

func (d *debugDetector) Observe(p []byte, now time.Time) debugAnomaly {
	data := append(append([]byte(nil), d.pending...), p...)
	d.pending = nil
	for i := 0; i < len(data); {
		if data[i] == 0x18 || data[i] == 0x1a {
			d.pending = nil
			i++
			continue
		}
		if data[i] == 0x1b {
			if i+1 == len(data) {
				d.pending = append(d.pending, data[i:]...)
				break
			}
			if data[i+1] == '[' || data[i+1] == ']' || data[i+1] == 'P' || data[i+1] == '^' || data[i+1] == '_' || data[i+1] == 'X' {
				start := i
				terminator := byte(0)
				if data[i+1] == '[' {
					i += 2
					canceled := false
					for i < len(data) && !(data[i] >= 0x40 && data[i] <= 0x7e) {
						if data[i] == 0x18 || data[i] == 0x1a {
							i++
							canceled = true
							break
						}
						i++
					}
					if canceled {
						continue
					}
					if i == len(data) {
						d.pending = append(d.pending, data[start:]...)
						break
					}
					i++
				} else {
					terminator = 7
					i += 2
					for i < len(data) && data[i] != terminator && data[i] != 0x18 && data[i] != 0x1a {
						if data[i] == 0x1b && i+1 < len(data) && data[i+1] == '\\' {
							i += 2
							break
						}
						i++
					}
					if i == len(data) && data[len(data)-1] != terminator {
						d.pending = append(d.pending, data[start:]...)
						break
					}
					if i < len(data) && data[i] == terminator {
						i++
					}
				}
				if len(data[start:i]) > debugControlLimit {
					d.pendingAt = time.Time{}
					return debugOverlong
				}
				continue
			}
			i += 2
			continue
		}
		if data[i] < 0x20 || data[i] == 0x7f {
			i++
			continue
		}
		if !utf8.FullRune(data[i:]) {
			d.pending = append(d.pending, data[i:]...)
			break
		}
		_, n := utf8.DecodeRune(data[i:])
		i += n
	}
	if len(d.pending) > debugControlLimit {
		d.pending = nil
		d.pendingAt = time.Time{}
		return debugOverlong
	}
	if len(d.pending) > 0 {
		if d.pendingAt.IsZero() {
			d.pendingAt = now
		}
	} else {
		d.pendingAt = time.Time{}
	}
	return ""
}

func (d *debugDetector) Stalled(now time.Time) debugAnomaly {
	if len(d.pending) > 0 && !d.pendingAt.IsZero() && now.Sub(d.pendingAt) > debugStallTimeout {
		return debugIncomplete
	}
	return ""
}

var leakedANSIFragment = regexp.MustCompile(`^(?:[0-9]{1,3};){2,}[0-9]{1,3}[mHf]$|^38;2;[0-9;]{3,}$`)

func suspiciousDebugText(text string) bool {
	if bytes.ContainsRune([]byte(text), 0x1b) {
		return true
	}
	for _, field := range bytes.Fields([]byte(text)) {
		if leakedANSIFragment.Match(field) {
			return true
		}
	}
	return false
}

// writeDebugBundle writes the stream log and before/after VT captures, then
// removes oldest complete bundles until at most five remain.
func writeDebugBundle(dir string, tabID int, reason string, ring *debugRing, before, after string, at time.Time) error {
	if dir == "" {
		return fmt.Errorf("empty debug directory")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	prefix := filepath.Join(dir, fmt.Sprintf("loom-session-%s-tab%d", at.Format("20060102-150405.000000000"), tabID))
	logPath := prefix + ".log"
	files := []struct{ path, content string }{
		{logPath, fmt.Sprintf("time: %s\ntab: %d\nanomaly: %s\nraw stream chunks (offset: quoted bytes):\n", at.Format(time.RFC3339Nano), tabID, reason)},
		{prefix + "-before.ansi", before},
		{prefix + "-after.ansi", after},
	}
	if ring != nil {
		var b bytes.Buffer
		for _, chunk := range ring.Chunks() {
			fmt.Fprintf(&b, "%d: %q\n", chunk.Offset, chunk.Data)
		}
		files[0].content += b.String()
	}
	created := make([]string, 0, len(files))
	for _, file := range files {
		f, err := os.OpenFile(file.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			for _, path := range created {
				_ = os.Remove(path)
			}
			return err
		}
		created = append(created, file.path)
		_, writeErr := f.WriteString(file.content)
		closeErr := f.Close()
		if writeErr != nil || closeErr != nil {
			for _, path := range created {
				_ = os.Remove(path)
			}
			if writeErr != nil {
				return writeErr
			}
			return closeErr
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var logs []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".log" && len(entry.Name()) >= len("loom-session-") && entry.Name()[:len("loom-session-")] == "loom-session-" {
			logs = append(logs, entry)
		}
	}
	sort.Slice(logs, func(i, j int) bool { return logs[i].Name() < logs[j].Name() })
	for len(logs) > debugBundleLimit {
		name := logs[0].Name()
		base := filepath.Join(dir, name[:len(name)-len(".log")])
		for _, suffix := range []string{".log", "-before.ansi", "-after.ansi"} {
			_ = os.Remove(base + suffix)
		}
		logs = logs[1:]
	}
	return nil
}

func sessionDebugDirectory(configured string) (string, error) {
	if configured != "" {
		return configured, nil
	}
	if stateHome := os.Getenv("XDG_STATE_HOME"); stateHome != "" {
		return filepath.Join(stateHome, "loom", "debug"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "state", "loom", "debug"), nil
}
