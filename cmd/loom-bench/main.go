// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command loom-bench is a self-contained, non-interactive smoke harness over
// loom's examples/* programs (see internal/examplesreg). It does not shell
// out to `go build` or exec subprocesses: it imports every example's Run
// function directly (the same way cmd/loom-demo does) and calls it in
// process, so loom-bench works as a plain installed binary with no source
// tree or go toolchain nearby.
//
// Examples are interactive TUIs that read from a live TTY, so loom-bench
// does not attempt to drive full interactive behavior. For v1 the smoke
// pass is intentionally minimal: for each example that supports a fast,
// non-blocking `--help` exit (see Example.SupportsHelp; all examples use
// flag.ContinueOnError or Cobra with SilenceErrors, so `--help` returns an
// error instead of calling os.Exit, making it safe to call in process), it
// calls Run(["--help"]) under a bounded timeout and reports PASS/FAIL and
// elapsed time. Examples without a fast-exit flag (currently: split) are
// skipped with that reason reported, since running them would block on a
// live TTY.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"codeberg.org/ubunatic/loom/internal/examplesreg"
)

// perExampleTimeout bounds each example's smoke pass. The whole harness is
// meant to run in well under a second per example, not open-ended.
const perExampleTimeout = 5 * time.Second

type result struct {
	name    string
	status  string // "PASS", "FAIL", "SKIP"
	elapsed time.Duration
	detail  string
}

func main() {
	os.Exit(run())
}

func run() int {
	results := make([]result, 0, len(examplesreg.Registry))
	failed := false
	for _, e := range examplesreg.Registry {
		r := smokeTest(e)
		results = append(results, r)
		if r.status == "FAIL" {
			failed = true
		}
	}

	for _, r := range results {
		fmt.Printf("%-4s %-12s %8s  %s\n", r.status, r.name, r.elapsed.Round(time.Millisecond), r.detail)
	}

	if failed {
		fmt.Println("loom-bench: one or more examples failed")
		return 1
	}
	fmt.Println("loom-bench: all examples passed")
	return 0
}

// runResult carries a smoke-run outcome across the timeout select below.
type runResult struct {
	err error
}

func smokeTest(e examplesreg.Example) result {
	if !e.SupportsHelp {
		return result{name: e.Name, status: "SKIP", detail: "no fast-exit flag; interactive TTY required to run"}
	}

	restore := silenceStdIO()
	start := time.Now()
	done := make(chan runResult, 1)
	go func() {
		done <- runResult{err: e.Run([]string{"--help"})}
	}()
	defer restore()

	select {
	case r := <-done:
		elapsed := time.Since(start)
		// examples using the standard flag package (rather than Cobra, which
		// silences help into a nil error) return flag.ErrHelp on -h/--help;
		// that is the documented clean exit, not a failure.
		if r.err != nil && !errors.Is(r.err, flag.ErrHelp) {
			return result{name: e.Name, status: "FAIL", elapsed: elapsed, detail: fmt.Sprintf("--help failed: %v", r.err)}
		}
		return result{name: e.Name, status: "PASS", elapsed: elapsed, detail: "--help OK"}
	case <-time.After(perExampleTimeout):
		return result{name: e.Name, status: "FAIL", elapsed: time.Since(start), detail: fmt.Sprintf("timed out after %s running --help", perExampleTimeout)}
	}
}

// silenceStdIO redirects os.Stdout/os.Stderr to a drained pipe for the
// duration of one example's --help call, so flag/Cobra usage text doesn't
// clutter loom-bench's own report. It returns a restore func.
func silenceStdIO() func() {
	origOut, origErr := os.Stdout, os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		return func() {}
	}
	os.Stdout, os.Stderr = w, w
	go io.Copy(io.Discard, r) //nolint:errcheck // best-effort drain
	return func() {
		os.Stdout, os.Stderr = origOut, origErr
		w.Close()
		r.Close()
	}
}
