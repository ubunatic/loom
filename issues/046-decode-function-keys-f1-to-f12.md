# 046 — Decode function keys F1 to F12

**Status**: Closed — resolved
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: `event.go`, `event_test.go`,
[issue 045](045-decode-modified-cursor-keys-shift-ctrl-alt-arrows.md), `288230f`

---

## 1. Problem & Motivation

`loom.DecodeKey` previously did not decode terminal function keys (F1–F12). Standard terminal shortcuts like F1 (Help) or F10 (Menu/Quit) were returned as unhandled escape sequences (`KeyEvent{}`).

Standard terminal function key encodings:
- F1–F4: SS3 format `\x1bOP`..`\x1bOS` (or CSI `\x1b[11~`..`\x1b[14~`)
- F5–F12: CSI tilde format `\x1b[15~`, `\x1b[17~`..`\x1b[21~` (F10), `\x1b[23~` (F11), `\x1b[24~` (F12)

---

## 2. Technical Implementation

1. **`event.go`**:
   - Added decoding for SS3 `\x1bOP`..`\x1bOS` as `"f1"`..`"f4"`.
   - Added decoding for CSI tilde numeric sequences `11`..`14` (F1..F4), `15` (F5), `17`..`21` (F6..F10), `23` (F11), `24` (F12).
2. **`event_test.go`**:
   - Added unit test cases for F1–F12 in `TestDecodeKey`.

---

## 3. Verification & Resolution

- Verified with `go test ./...` in `loom` (all tests passing).
- Verified with `cati`'s `imgbrowser` using F10 to exit.
