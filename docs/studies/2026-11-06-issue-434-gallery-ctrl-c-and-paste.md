---
kind: sprint
date: 2026-11-06
ticket: 434
---

# Issue 434 M1: gallery TextEntry Ctrl+C and paste

- **Baseline**: Loom already had modeled chat-canvas and `Pane.run` PTY paste preservation coverage from `49e5bad`. Neither that history nor this milestone reproduced the reported Taia terminal clearing.
- **Decision**: keep library TextEntry semantics unchanged. The gallery parent offers Ctrl+C to TextEntry first, preserving selection-copy; otherwise it clears nonempty input and consumes the key. Empty input remains ignored so normal parent routing applies.
- **Coverage**: gallery unit test covers selected copy precedence, clear, ignored empty Ctrl+C, and subsequent typing/paste. The command PTY test exercises `loom widgets --show TextEntry`, verifies the top-level process remains active after empty Ctrl+C, and checks paste and gallery controls.
- **Clipboard evidence**: existing library tests assert exact OSC 52 bytes and clipboard-hook delivery. The gallery wrapper does not expose the private binding path, so its PTY test makes no claim about clipboard delivery to a desktop terminal.
- **Regression-first**: before the gallery handler was added, `go test ./gallery -run '^TestTextEntryDemoCtrlCPolicy$' -count=1` failed at unselected nonempty Ctrl+C. Afterward, targeted routing tests passed; `go test ./gallery ./cmd/loom` passed; `go vet ./...` passed.
- **Quota verification**: `make test-q1` passed once (exit 0; full output at `/tmp/loom-issue434-test-q1.log`; no `--- FAIL` output).
- **Remaining**: original Taia report remains unreproduced and unattributed; issue 434 remains open. Manual check: `loom widgets --show TextEntry`; Ctrl+A, Ctrl+C (draft stays); Right, Ctrl+C (input clears, gallery controls stay); Ctrl+C (gallery remains); paste `after\nclear` (text appears and controls remain). Verify desktop clipboard receipt separately.
- **Install**: `make install` passed.

## M2: app-level Ctrl+C confirmation

- **Delivery**: `loom widgets --show TextEntry` alone opts into the gallery confirmation state. It offers Ctrl+C to TextEntry first, so selection copy retains precedence; otherwise a nonempty entry clears without arming, and empty input requires a visible confirmation followed by a consecutive Ctrl+C to quit. Other gallery demos, library behavior, F10 and Ctrl+Q are unchanged.
- **Cancellation/layout**: typing, navigation, paste, mouse/focus changes and F8/F9 reset pending confirmation. The row is repainted blank after cancellation. Tests cover 0/1/2 row heights and a narrow one-row canvas. No timer is used.
- **Evidence**: regression-first `cmd/loom` tests and the actual command PTY test pass; the PTY waits on the process after the final Ctrl+C. `go vet ./...` passes. The one `make test-q1` run failed at `internal/session/TestSessionDetachPrefixDAndReattach` (`session_pty_test.go:167`) on retained screen/tabs repaint; full output is `/tmp/loom-issue434-m2-test-q1.log`. No repeat run. `make install` passed. Selected text preservation is not proof of desktop clipboard delivery.
- **Manual check**: `loom widgets --show TextEntry`; Ctrl+A/Ctrl+C leaves selection; Right/Ctrl+C clears while controls remain; Ctrl+C shows the hint; type `x` to cancel and verify input remains usable; Ctrl+C clears `x` without arming; paste `after\nclear`, verify both lines and F2/F9/F10; Ctrl+C clears pasted text, Ctrl+C shows hint, next Ctrl+C exits. Start another fresh empty demo and confirm two presses exit. F10 and Ctrl+Q remain immediate. Verify desktop clipboard receipt separately.
- **Limit**: this only demonstrates Loom’s in-repository gallery policy. Taia’s reported terminal clearing remains unreproduced/unattributed, so issue 434 stays open for host review and original-report attribution.

## M3: preserve control-payload and child geometry coverage

- **Hardening**: the real-command Ctrl+C/paste PTY flow again sends `\x1b[2J` inside bracketed paste and continues to assert the pasted lines and gallery controls. This checks surrounding display retention for a control-containing payload as well as plain text.
- **Geometry**: `themedGallery.Draw` clamps the child height only after reserving all rows. A draw-recording child test exercises heights 0–3 both with and without hotkey hints. Before the clamp, it failed at height 2 with a negative child height (`H:-1`) in both cases.
- **Verification**: the focused post-fix package run `go test ./cmd/loom -count=1`, `go vet ./...`, and the single `make test-q1` passed. The full quota log is `/tmp/loom-issue434-m3-test-q1.log`. `make install` passed. The prior M2 quota failure in `TestSessionDetachPrefixDAndReattach` did not recur; that test’s assertions were not changed.
- **Limit**: no Taia reproduction or desktop clipboard receipt was established. Issue 434 remains open pending host review and original-report attribution.
