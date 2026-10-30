# 389 — Detect modifier keys via voxi-modifierd on non-kitty terminals

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 384 (ctrl-1..9 won't fix without kitty), 321 (kitty CSI-u decoding), voxi (codeberg.org/ubunatic/voxi)

---

/goal When `voxi-modifierd` is installed and its state file is readable, loom can tell which
physical modifiers are held for a key event on a legacy (non-kitty) terminal, so bindings such as
`ctrl-1`..`ctrl-9` can be recovered; without the daemon, behaviour is unchanged. Stop and report when
blocked on a user decision, on the state-file format, or when the timing proves unreliable; on a loop
or no progress, clear the goal and write a failure report saying why.

Model: start on `codex:sol:med` (touches key decoding / event routing, see AGENTS.md).

## 1. Problem & Motivation

Legacy terminals (for example the Gram integrated terminal, see 384) do not report kitty keyboard
events. Ctrl-digits are then either dropped or collide with other keys (`^1` arrives as `1`, `^3`
arrives as Esc). We keep `alt-<num>` as the binding, but a host-side source of modifier state would
let loom disambiguate these keys without terminal support.

## 2. Technical Specification / Findings

- `voxi-modifierd` (system unit `voxi-modifierd.service`, from the voxi project) reads evdev and
  writes only the current physical modifier state, with no other keys logged, to
  `/run/voxi/modifiers` every 10 ms (`--state-file /run/voxi/modifiers --interval 10ms`). The file
  is world-readable (`0644`) and contained a single byte, `\0`, when no modifiers were held
  (2026-10-30).
- Unknown: the bitmask layout (ctrl/alt/shift/super bits, left/right). Read it from the voxi
  source instead of guessing.
- Detection: the feature is available only when the state file exists and is readable. Do not
  shell out to systemctl. It is opt-in or automatic only on legacy terminals; kitty mode always
  takes precedence.
- Limits: the daemon sees only the local physical keyboard. It is wrong over SSH and for remote or
  container sessions, and it races with input arriving 10 ms later, so samples must be taken on byte
  arrival. Ambiguous bytes (`1` with ctrl held → `ctrl-1`) are re-interpreted only when the sample
  agrees.
- Where: the decoder layer (`loom.DecodeKey` / session input path); fix the library, not callers.

## 3. Implementation & Verification Plan

- Canary first (docs/Canary.md): a `loom-probe` subcommand that prints raw bytes plus the sampled
  modifier state, run in Gram, to prove that `^1`..`^9` can be recovered reliably.
- If the canary passes, add the modifier source behind a spec flag with an injectable reader, so
  tests use a fake state file. Add unit tests for the re-interpretation table, then run
  `make test-q1`.
- If timing is unreliable, close as won't fix and record the canary results here.
