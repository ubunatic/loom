# 384 — loom session: opt-in always-on kitty keyboard mode with legacy re-encoding and ctrl-1..9 tab switching

**Status**: Blocked — needs a kitty-capable terminal; the user's Gram integrated terminal reports neither kitty nor modifyOtherKeys and sends nothing for ctrl-digits
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 321 (kitty flag 1 / CSI-u decoding), 372 (terminal restore), 383 (attach/detach), `cmd/loom-probe` (`loom-probe keys` canary)

---

/goal With `session.kitty_keyboard: auto` on a terminal that reports the kitty keyboard protocol,
`ctrl-1`..`ctrl-9` switch session tabs, while every tab whose child did not request kitty mode keeps
receiving byte-identical legacy input; with `off` (the default) or on a legacy terminal, behaviour is
unchanged; or stop and report when blocked on a user decision.

Model: start on `codex:sol:med` (touches event routing, see AGENTS.md).

## 1. Problem & Motivation

The user wants `^1`..`^9` to switch session tabs (next to `alt-1`..`alt-9` and prefix `1`..`9`).
`loom.DecodeKey` already decodes `\x1b[49;5u` to `ctrl-1` (`event.go` `decodeCSIu`), but the session
only enables kitty mode on the outer terminal while the active child requested it
(`internal/session/session.go` `switchActiveTab`, `oldKitty != newTab.vt.KittyKeyboard`). In a plain
shell tab the outer terminal is legacy, where ctrl-digits are ambiguous or lost:

| Key | Legacy bytes (xterm-like) | Collides with |
| --- | --- | --- |
| `^1`, `^9` | `1`, `9` | plain digits |
| `^2` | `\x00` | ctrl-space |
| `^3` | `\x1b` | Esc |
| `^4`..`^7` | `\x1c`..`\x1f` | ctrl-\ ] ^ _ |
| `^8` | `\x7f` | Backspace |

Confirmed live with `loom-probe keys` in tmux: `C-3` → `"\x1b"` (esc), `C-8` → `"\x7f"` (backspace).
Adding `ctrl-N` bindings alone would therefore only work while a kitty-aware app is in front.

## 2. Technical Specification / Findings

### Probe result on the user's terminal (2026-10-28)

`loom-probe keys` in the Gram editor's integrated terminal (`TERM=xterm-256color`, `TERM_PROGRAM=gram`, no tmux, no loom session):

| Key | Phase 1 (legacy) | Phase 3 (modifyOtherKeys=2) |
| --- | --- | --- |
| Ctrl-1..Ctrl-9 | nothing | nothing |
| Alt-1 | nothing; `"\x1b1"` (alt-1) once the Gram `Terminal` keymap sends it | same as phase 1 |
| Shift-Enter | `"\n"` | `"\n"` |
| Enter | `"\r"` | `"\r"` |
| Esc | `"\x1b"` | `"\x1b"` |

Kitty query: not supported (DA1 only), so phase 2 was skipped. Phase 3 was identical to phase 1, so
modifyOtherKeys is not supported either. Shift-Enter is distinct only because it sends LF, which
legacy Ctrl-J also sends. Alt-1 produced nothing at first because Gram binds Alt-digits itself. The
working setup keeps Gram's own Alt-digit bindings and adds a `"Terminal"` context block to
`~/.config/gram/keymap.jsonc` that maps `alt-N` to `["terminal::SendText", "\u001bN"]` (and unbinds
`ctrl-n` in `Workspace` so it reaches the terminal). Alt-1 then arrives as `"\x1b1"`, so the
session's existing `alt-1`..`alt-9` work there and cover the use case. For new Alt or Ctrl bindings,
fix conflicts in that `Terminal` context, not by disabling Gram bindings globally. Conclusion: ctrl-digit tab switching needs a kitty-capable terminal (kitty, Ghostty, WezTerm,
foot, Alacritty); the user's terminal cannot support it.

- **Detection** per client attach: write `CSI ? u` + `CSI c`; a `CSI ? <flags> u` reply before the DA1
  reply means supported. Never wait past DA1. `cmd/loom-probe/keys.go` `parseKittyQueryReply` is the
  reference parser. Re-detect on every attach (383): clients on different terminals may differ.
- **Fallback: xterm modifyOtherKeys=2** (`CSI > 4 ; 2 m`). The user's terminal does not report kitty
  but sends Shift-Enter distinctly from Enter, which hints at modifyOtherKeys. There is no query for
  it, so `auto` cannot detect it; it would need an explicit `session.kitty_keyboard: modify_other_keys`
  value. `decodeModifyOtherKeys` in `event.go` only decodes `a`..`z` today and must cover digits,
  Enter, Tab and punctuation first. `loom-probe keys` phase 3 shows what the terminal sends.
- **Flags**: push only flag 1 (disambiguate). Plain text, Enter, Tab and Backspace stay legacy; Esc,
  ctrl-/alt-combos arrive as CSI-u.
- **Re-encoding** for tabs whose child has not pushed kitty flags: map decoded CSI-u back to the legacy
  bytes (`ctrl-a`..`ctrl-z` → 0x01..0x1a, `alt-x` → `ESC x`, `esc` → 0x1b, `ctrl-[` → 0x1b, modified
  arrows/F-keys keep their legacy CSI forms). Keys without a legacy form (`ctrl-1`, `ctrl-enter`, …)
  are consumed by session bindings or dropped. Kitty-aware tabs receive raw bytes as today.
- **Per-tab flags**: track the flags each child pushed (`CSI > n u`, `CSI = n ; m u`, `CSI < u`)
  instead of the boolean `vt.KittyKeyboard`; today `session.go` always pushes `>1u`.
- **Restore**: pop the pushed flags on every exit path, detach included (`terminal.Restorer`
  `SetKittyKeyboard`). After `kill -9` the outer shell sees CSI-u until `reset`; document it.
- **Spec**: `session.kitty_keyboard: off|auto` (default `off`) and `ctrl-1`..`ctrl-9` appended to
  `session.bindings.select_tab` in `spec/defaults.yaml`; no Go literals for either.

## 3. Implementation & Verification Plan

1. Run `loom-probe keys` on the user's real terminal(s) (and inside tmux/ssh if used); record what
   `^1`..`^9`, `^3`, Esc, Alt-1 and Ctrl-A send (no `^[` on a DE layout; Esc covers it) in both phases in this ticket. Stop if the terminal
   does not report kitty support or ctrl-digits do not arrive as CSI-u.
2. Re-encoder as a pure function with an exhaustive table test (every ctrl/alt letter, digits,
   punctuation, Esc, Enter, Tab, Backspace, modified arrows and F-keys) asserting byte-identical legacy
   output.
3. Session: detection on attach, opt-in push, per-tab flag tracking, re-encoding on the input path,
   `ctrl-N` select bindings, restore on all exit and detach paths.
4. PTY tests (full tier): fake terminal answering the kitty query; a legacy shell tab receives
   `\x01` for CSI-u ctrl-a and `\x1b` for CSI-u esc; a kitty child receives raw CSI-u; `\x1b[49;5u`
   switches to tab 1; detach/attach from a non-kitty client restores legacy mode.
5. Live check in tmux and the user's terminal; `docs/Session.md` documents the setting and the
   `kill -9` caveat. If `auto` later becomes the default, add a row to `docs/Upgrading.md`.
