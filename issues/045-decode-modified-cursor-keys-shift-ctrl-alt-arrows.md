# 045 — Decode modified cursor keys: shift-, ctrl-, alt- arrows

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: `event.go`, `event_test.go`, `pane.go`

---

## 1. Problem & Motivation

`loom.DecodeKey` previously only recognized unmodified cursor keys (e.g. `\x1b[A` for Up, `\x1b[B` for Down) and single-byte control characters. 

Standard modern terminal emulators (xterm, GNOME Terminal, iTerm2, Alacritty, Kitty) emit CSI modifier sequences when cursor keys are pressed with modifier keys:
- Shift+Up/Down/Right/Left: `\x1b[1;2A`, `\x1b[1;2B`, `\x1b[1;2C`, `\x1b[1;2D` (or rxvt `\x1b[a`..`\x1b[d`)
- Alt+Up/Down/Right/Left: `\x1b[1;3A`, `\x1b[1;3B`, `\x1b[1;3C`, `\x1b[1;3D`
- Ctrl+Up/Down/Right/Left: `\x1b[1;5A`, `\x1b[1;5B`, `\x1b[1;5C`, `\x1b[1;5D`
- Ctrl+Shift+Up/Down: `\x1b[1;6A`, `\x1b[1;6B`

Because `DecodeKey` did not handle the `1;` modifier prefix, all modified arrow key sequences were discarded as unhandled escape sequences (`return KeyEvent{}`), making it impossible for applications built on `loom` to handle shortcuts like Shift-Up/Down or Ctrl-Up/Down.

---

## 2. Technical Implementation

1. **`event.go`**:
   - Added decoding for CSI `\x1b[1;<mod><A|B|C|D>` format:
     - Modifier `2`: `"shift-"`
     - Modifier `3`: `"alt-"`
     - Modifier `4`: `"shift-alt-"`
     - Modifier `5`: `"ctrl-"`
     - Modifier `6`: `"ctrl-shift-"`
   - Added rxvt modifier shorthand sequences (`\x1b[a` $\to$ `"shift-up"`, `\x1b[b` $\to$ `"shift-down"`, etc.).
2. **`event_test.go`**:
   - Added unit test cases for xterm shift/ctrl/alt arrow keys and rxvt shift arrow keys in `TestDecodeKey`.

---

## 3. Verification & Resolution

- Verified with `go test ./...` in `loom` (all packages passing).
- Verified end-to-end integration with `cati`'s `imgbrowser` using `replace codeberg.org/ubunatic/loom => ../loom`.
