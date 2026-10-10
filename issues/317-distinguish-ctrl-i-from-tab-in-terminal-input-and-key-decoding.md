# 317 — Distinguish Ctrl+I from Tab in terminal input and key decoding

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**:

---

## 1. Problem & Motivation
In legacy ASCII terminal protocol, typing `Ctrl+I` produces the exact byte value `0x09`, which is identical to the ASCII `Tab` key (`\t`).
As a result:
- In terminals running in standard mode, pressing `^I` sends `0x09`, and `DecodeKey` parses it as `tab`.
- Applications cannot reliably distinguish whether the user physically pressed `Tab` (for navigation/indentation) or `Ctrl+I` (often desired for italic formatting, info shortcuts, or custom actions).
- If an application binds `ctrl-i`, legacy terminals trigger that action when `Tab` is pressed, breaking tab navigation unless guarded.

How can Loom safely allow users to press `^I` and distinguish it from `Tab` while preserving standard `Tab` functionality?

## 2. Technical Findings & Solutions

### A. Progressive Keyboard Protocol / Disambiguation Modes (Recommended)
Modern terminal emulators support protocols that unambiguously differentiate keys:
1. **Kitty Keyboard Protocol (CSI-u)**:
   - Enabling progressive enhancement sends `\x1b[105;5u` for `Ctrl+I`, while physical `Tab` sends `\x1b[9u` (or `0x09` when unshifted).
   - Loom's `DecodeKey` in `event.go` already decodes `\x1b[105;5u` as `ctrl-i` and `\x1b[9;...u` as `tab`.
2. **XTerm `modifyOtherKeys` (format 2)**:
   - When enabled via `\x1b[>4;2m`, `Ctrl+I` sends `\x1b[27;5;105~`, while `Tab` remains `0x09` (or `\x1b[27;2;9~` for Shift-Tab).
   - Loom already decodes `\x1b[27;5;105~` via `decodeModifyOtherKeys`.

**Solution 1: Opt-in Terminal Disambiguation Sequence**
- In `Pane.New` / raw mode setup, optionally emit CSI-u / `modifyOtherKeys` push mode when enabled by spec (`keyboard_enhancement: true` or `disambiguate_keys: true`), and restore it on `Pane.Close`.
- Under this mode, `Ctrl+I` arrives as `ctrl-i` and `Tab` arrives as `tab` without collision.

### B. Fallback / Legacy Terminal Mapping Strategy
For terminals that lack CSI-u / `modifyOtherKeys` (e.g. VTE-based terminals such as GNOME Terminal / Tilix on primary screen):
- Physical `Ctrl+I` and `Tab` generate the identical byte `0x09`.
- **Policy 1: Tab Priority (Default Safe Policy)**
  - `0x09` always decodes to `tab`. `ctrl-i` is only produced when the terminal transmits an unambiguous CSI-u or `modifyOtherKeys` sequence.
  - Apps desiring italic formatting use alternative shortcuts (`ctrl-i` works where supported; `F`-keys or popover shortcuts provide universal access).
- **Policy 2: Contextual Disambiguation**
  - When focus is in a context where Tab is not a text-entry key (e.g. within an active text field without tab-stop support), or when configured via application keymaps.
- **Policy 3: Configurable Key Remapping (`spec/keys.yaml`)**
  - Allow users/apps to configure whether legacy `0x09` maps to `tab` (default) or `ctrl-i`.

## 3. Implementation & Verification Plan
- Investigate terminal query for progressive keyboard protocol support (`kitty_keyboard` flags or terminfo).
- Add configurable terminal key disambiguation in `spec/` (e.g. `spec/defaults.yaml` under `terminal`).
- Verify with unit tests in `event_test.go` and PTY interaction tests in `pane_test.go` or `cmd/loom/edit_pty_test.go`.

/goal Design and implement terminal input disambiguation so Ctrl+I and Tab can be distinguished when supported, while safely preserving standard Tab key behavior, or stop and report if blocked on a user decision or denied permission.
