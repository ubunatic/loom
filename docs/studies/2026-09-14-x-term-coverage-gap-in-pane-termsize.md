---
title: x/term Coverage Gap in Pane.termSize
---

# x/term Coverage Gap in Pane.termSize

## Header & Context

Date: 2026-09-14. Prompted by a direct question during treemap example work: "Do we fully leverage x/term for all terminal state management in loom?" This is a focused finding write-up, not a sprint retrospective — it documents one concrete gap found, the fix applied, and what was deliberately left alone, for an independent reviewer to assess.

Status of the fix at time of writing: applied to the working tree (`pane.go`), **not yet committed**. `git diff pane.go` at the end of this document is the literal, current diff.

## Executive Summary

`loom`'s root package already uses `golang.org/x/term` for the two things it's designed for: entering/restoring raw mode (`term.MakeRaw` / `term.Restore` in `pane.go`) and, as of this finding, size queries (`term.GetSize`). Before this fix, `pane.go`'s `termSize()` did **not** use `term.GetSize` — it hand-rolled the same `TIOCGWINSZ` ioctl directly via `syscall.Syscall` and `unsafe.Pointer`, in the same file that already imports `x/term` for other purposes. This was pure duplication of functionality the dependency already provides, done less safely and less portably than necessary.

Two other raw-syscall terminal-state touchpoints were audited and found to be legitimate, not duplicative — `x/term` has no equivalent for either, so they correctly remain hand-rolled. See "What Was Not Changed."

## The Gap

### Before

```go
// termSize returns the terminal dimensions via TIOCGWINSZ.
func termSize(fd int) (cols, rows int) {
	var ws struct{ Row, Col, X, Y uint16 }
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}
```

Call sites (unchanged by the fix, all in `pane.go`): `New()` (initial size at pane open), `Resize()` (explicit height change), and the SIGWINCH resize handler (terminal window resize while a pane is running).

### After

```go
// termSize returns the terminal dimensions via x/term (TIOCGWINSZ under the
// hood on Unix), falling back to a conventional 80x24 on error -- this used
// to hand-roll its own syscall.Syscall(SYS_IOCTL, TIOCGWINSZ, ...) call
// even though term.GetSize (already imported in this file for MakeRaw/
// Restore) does the same thing more portably.
func termSize(fd int) (cols, rows int) {
	cols, rows, err := term.GetSize(fd)
	if err != nil || cols < 1 || rows < 1 {
		return 80, 24
	}
	return cols, rows
}
```

The `"unsafe"` import was dropped entirely from `pane.go` as a consequence — it was used nowhere else in the file.

### Why This Matters

1. **Duplication of a dependency the code already pulls in.** `pane.go` imports `golang.org/x/term` on line 37 specifically for `MakeRaw`/`Restore`. Re-implementing `GetSize` by hand in the same file, using a lower-level and more fragile mechanism, is the kind of drift that accumulates unnoticed — nothing forces a reader to realize `x/term` already solves the problem sitting three lines below their own hand-rolled version.
2. **`unsafe.Pointer` for no remaining benefit.** The hand-rolled version required `unsafe.Pointer(&ws)` to pass a local struct into a raw syscall. `x/term`'s implementation wraps the equivalent platform-specific ioctl internally and exposes a plain `(int, int, error)` signature — no `unsafe` needed at the call site. Removing an `unsafe` usage that buys nothing is a straightforward safety win.
3. **Portability.** `syscall.SYS_IOCTL` and `syscall.TIOCGWINSZ` are Unix-specific constants (and on some GOOS/GOARCH combinations may not even be defined, depending on the `syscall` package's per-platform surface). `x/term.GetSize` has real per-OS implementations (Unix via ioctl, Windows via `GetConsoleScreenBufferInfo`), so routing through it is strictly more portable than the hand-rolled call, even though `loom` is currently Unix/`/dev/tty`-focused throughout (see `docs/TuiInput.md` — `Pane.New` explicitly always opens `/dev/tty`, never `os.Stdin`). This fix doesn't make `loom` cross-platform by itself, but it removes one of the places that would actively block it.
4. **Same struct shape, verified behavior-preserving.** `ws.Col`/`ws.Row` (uint16 fields read from the ioctl result) map directly to `x/term.GetSize`'s `(width, height int)` return — same underlying kernel call, same field order, so this is a behavior-preserving refactor, not a semantic change. The `errno != 0` / `err != nil` fallback to a hardcoded `80, 24` is preserved verbatim.

## What Was Not Changed (and Why)

The audit (`grep -rn "SYS_IOCTL\|TIOCGWINSZ\|TCGETS\|TCSETS\|termios\." --include=*.go`, excluding `_test.go`) found exactly one file with raw terminal-ioctl syscalls: `pane.go`. Within it, two other spots remain hand-rolled deliberately:

```go
// drainInput flushes the terminal input queue (TCIFLUSH) so raw-mode leftovers
// don't bleed into whatever reads the terminal next.
func (p *Pane) drainInput() {
	const tcflsh = 0x540B // Linux TCFLSH; argument 0 selects TCIFLUSH (input queue)
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, uintptr(p.fd), tcflsh, 0)
}
```

`x/term` exposes no flush/drain primitive — its surface is limited to raw-mode enter/restore, size queries, and a couple of state-inspection helpers. There is no library replacement here; this ioctl is already hardcoded to the Linux `TCFLSH` value (`0x540B`) with a comment acknowledging that. If cross-platform support is ever pursued, this is a second, separate gap from the one this document covers — it would need a build-tag-gated implementation per OS, not a swap to an existing `x/term` call.

Signal registration (`signal.Notify(p.interrupts, syscall.SIGINT, syscall.SIGTERM)` and `signal.Notify(p.winch, syscall.SIGWINCH)`) also touches `syscall`, but this is `os/signal` usage, not terminal I/O state — out of scope for "leverage x/term," since `x/term` has nothing to do with signal delivery.

`golang.org/x/sys/unix.Poll` calls (`pane.go` lines ~161, ~384, ~391) are I/O multiplexing on the tty fd, used to make blocking `Read`s cancelable without racing a `Close` from another goroutine (see the in-code comments at each call site). `x/term` has no polling API either; `x/sys/unix` is the correct, and only, layer for this.

## Verification Performed

- `go build ./...` — clean.
- `go vet ./...` — clean.
- `gofmt -l pane.go` — clean after the edit (ran `gofmt -w` once to fix incidental blank-line diffs left by the edit tool; see the diff below).
- `go test ./...` — full suite passes, including `examples/monitor` and `examples/splash`, both of which exercise `Pane.New`/`Resize` (and therefore `termSize`) through their own test suites.

### Coverage Gap Worth Flagging to the Reviewer

There is **no unit test that calls `termSize` directly** — it's only exercised indirectly through `Pane.New()`, `Pane.Resize()`, and the SIGWINCH handler, all of which require a real (or PTY-simulated) tty fd. The existing `scripts/check-watch-pty.py` harness (invoked by `make watch-pty`) is the closest thing to a direct exercise of this path, since it opens a real PTY and resizes it. That script was not re-run as part of this fix (only `go test ./...`, which doesn't include PTY-based checks) — a reviewer verifying this change end-to-end should run `make watch-pty` as well, since it's the only path that would catch a `term.GetSize` behavioral mismatch (e.g. a return-value swap) that `go vet`/unit tests can't see.

## Open Questions for the Reviewer

1. Is the `drainInput`/`TCFLSH` hand-rolled ioctl acceptable to leave as-is given the project's current Unix/`/dev/tty`-only scope (per `docs/TuiInput.md`), or should it be tracked as its own follow-up ticket now that this document has surfaced it as a related gap?
2. Should there be a lightweight guard (e.g. a `go vet`-style check, or just a `grep` step in `make test`) that flags any future direct `syscall.Syscall(syscall.SYS_IOCTL, ...)` usage touching terminal state outside an explicitly allow-listed set, to prevent this kind of duplication from creeping back in as the codebase grows?
3. `TestClampDimensionsNeverExceedsTerminal` and friends (added alongside separate treemap work the same day) test *consumers* of terminal size (`examples/treemap`'s dimension-clamping logic) with an injected size, not `termSize` itself. Is a similar pure-function extraction worth doing for `termSize`'s fallback logic (`err != nil || cols < 1 || rows < 1` → `80, 24`), so that fallback path gets direct unit coverage without needing a real tty?

## Full Diff (`pane.go`, uncommitted at time of writing)

```diff
diff --git a/pane.go b/pane.go
index 5d4acc5..a55d11f 100644
--- a/pane.go
+++ b/pane.go
@@ -31,7 +31,6 @@ import (
 	"strings"
 	"syscall"
 	"time"
-	"unsafe"

 	"golang.org/x/sys/unix"
 	"golang.org/x/term"
@@ -40,7 +39,6 @@ import (
 // DefaultMaxCols is the default maximum canvas width loaded from SpeccedDefaults.
 var DefaultMaxCols = SpeccedDefaults.Pane.MaxCols

-
 // Pane manages an inline terminal region and drives the widget event loop.
 type Pane struct {
 	tty      *os.File
@@ -61,7 +59,6 @@ type Pane struct {
 	// Esc, Ctrl-C, Ctrl-Q, and q keys when the active widget returns false from HandleKey.
 	DisableDefaultQuit bool

-
 	// winch carries SIGWINCH notifications so the Run loop reflows on a
 	// terminal window resize. Buffered (cap 1) to coalesce resize bursts.
 	winch chan os.Signal
@@ -524,7 +521,6 @@ var defaultQuitKeyMap = func() map[string]bool {
 	return m
 }()

-
 // handleKeyFallback reports whether an unhandled key should trigger a default
 // safeguard exit based on embedded spec/defaults.yaml.
 func (p *Pane) handleKeyFallback(ke KeyEvent) bool {
@@ -538,8 +534,6 @@ func (p *Pane) handleKeyFallback(ke KeyEvent) bool {
 	return defaultQuitKeyMap[key]
 }

-
-
 // Close tears down the pane: clears the reserved region, restores terminal
 // state, and closes /dev/tty. Safe to call multiple times.
 func (p *Pane) Close() { p.close() }
@@ -606,13 +600,15 @@ func (p *Pane) installSignalHandler() {
 	signal.Notify(p.winch, syscall.SIGWINCH)
 }

-// termSize returns the terminal dimensions via TIOCGWINSZ.
+// termSize returns the terminal dimensions via x/term (TIOCGWINSZ under the
+// hood on Unix), falling back to a conventional 80x24 on error -- this used
+// to hand-roll its own syscall.Syscall(SYS_IOCTL, TIOCGWINSZ, ...) call
+// even though term.GetSize (already imported in this file for MakeRaw/
+// Restore) does the same thing more portably.
 func termSize(fd int) (cols, rows int) {
-	var ws struct{ Row, Col, X, Y uint16 }
-	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
-		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
-	if errno != 0 {
+	cols, rows, err := term.GetSize(fd)
+	if err != nil || cols < 1 || rows < 1 {
 		return 80, 24
 	}
-	return int(ws.Col), int(ws.Row)
+	return cols, rows
 }
```

Note: the blank-line removals around unrelated functions (`DefaultMaxCols`, `defaultQuitKeyMap`, `Close`) are `gofmt -w`'s doing, not intentional edits — they were pre-existing double-blank-lines that `gofmt` collapsed to single blanks when the file was reformatted after the `termSize` edit. Flagging this explicitly so the reviewer doesn't read them as deliberate, unrelated changes bundled into this diff.
