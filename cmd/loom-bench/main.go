// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command loom-bench is a non-interactive smoke harness over loom's
// examples/* programs (see internal/examplesreg). Examples are interactive
// TUIs that read from a live TTY, so loom-bench does not attempt to drive
// full interactive behavior. For v1 the smoke pass is intentionally
// minimal: for each example it runs `go build ./examples/<name>` and, for
// examples that support a fast, non-blocking `--help` exit (see
// Example.SupportsHelp), also runs the built binary with `--help` under a
// bounded timeout. Examples without a fast-exit flag (currently: split,
// which parses no flags at all) are build-checked only, and that is called
// out in the per-example report line.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"codeberg.org/ubunatic/loom/internal/examplesreg"
)

// perExampleTimeout bounds each example's build+run smoke pass. The whole
// harness is meant to run in a few seconds total, not open-ended.
const perExampleTimeout = 15 * time.Second

type result struct {
	name    string
	ok      bool
	elapsed time.Duration
	detail  string
}

func main() {
	os.Exit(run())
}

func run() int {
	tmp, err := os.MkdirTemp("", "loom-bench-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "loom-bench:", err)
		return 1
	}
	defer os.RemoveAll(tmp)

	results := make([]result, 0, len(examplesreg.Registry))
	failed := false
	for _, e := range examplesreg.Registry {
		r := smokeTest(tmp, e)
		results = append(results, r)
		if !r.ok {
			failed = true
		}
	}

	for _, r := range results {
		status := "PASS"
		if !r.ok {
			status = "FAIL"
		}
		fmt.Printf("%-4s %-12s %8s  %s\n", status, r.name, r.elapsed.Round(time.Millisecond), r.detail)
	}

	if failed {
		fmt.Println("loom-bench: one or more examples failed")
		return 1
	}
	fmt.Println("loom-bench: all examples passed")
	return 0
}

func smokeTest(tmpDir string, e examplesreg.Example) result {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), perExampleTimeout)
	defer cancel()

	binPath := filepath.Join(tmpDir, e.Name)
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, e.Package)
	buildCmd.Stdout = nil
	if out, err := buildCmd.CombinedOutput(); err != nil {
		return result{name: e.Name, ok: false, elapsed: time.Since(start), detail: fmt.Sprintf("build failed: %v: %s", err, firstLine(out))}
	}

	if !e.SupportsHelp {
		return result{name: e.Name, ok: true, elapsed: time.Since(start), detail: "build-only (no fast-exit flag; interactive TTY required to run)"}
	}

	runCmd := exec.CommandContext(ctx, binPath, "--help")
	out, err := runCmd.CombinedOutput()
	elapsed := time.Since(start)
	if ctx.Err() == context.DeadlineExceeded {
		return result{name: e.Name, ok: false, elapsed: elapsed, detail: fmt.Sprintf("timed out after %s running --help", perExampleTimeout)}
	}
	if err != nil {
		return result{name: e.Name, ok: false, elapsed: elapsed, detail: fmt.Sprintf("--help failed: %v: %s", err, firstLine(out))}
	}
	return result{name: e.Name, ok: true, elapsed: elapsed, detail: "build + --help OK"}
}

func firstLine(out []byte) string {
	for i, b := range out {
		if b == '\n' {
			return string(out[:i])
		}
	}
	return string(out)
}
