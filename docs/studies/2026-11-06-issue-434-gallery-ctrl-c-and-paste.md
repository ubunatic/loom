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
