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
- Format (voxi source, `internal/modifiers/modifiers.go`, verified 2026-10-30): one byte;
  bit 0 left Ctrl, 1 right Ctrl, 2 left Alt, 3 right Alt, 4 left Super, 5 right Super,
  6 left Shift, 7 right Shift. There is no AltGr bit. The daemon writes a temp file and renames it
  over the state file (atomic), so readers read byte 0 of a fresh open.
- voxi's reader is in `internal/` (not importable) and voxi is AGPL-3.0: loom implements its own
  reader against this documented file contract and copies no voxi code.
- Detection: the feature is available only when the state file exists and is readable. Do not
  shell out to systemctl. It is opt-in or automatic only on legacy terminals; kitty mode always
  takes precedence.
- Limits: the daemon sees only the local physical keyboard. It is wrong over SSH and for remote or
  container sessions, and it races with input arriving 10 ms later, so samples must be taken on byte
  arrival. Ambiguous bytes (`1` with ctrl held → `ctrl-1`) are re-interpreted only when the sample
  agrees.
- Where (searcher, 2026-10-30): `DecodeKey` `event.go:120` (legacy control map `:129`, CSI-u
  `:342`); pane input reads and splits at `pane_run.go:519-537`; session has its own input path at
  `internal/session/session.go:941`. There is no single choke point, so both paths need the
  library hook. The spec flag follows `pane.kitty_keyboard` (`spec/defaults.yaml:62`,
  `specdefaults/defaults.go:891`, `spec/schemas/defaults.schema.json:533`, consumer `pane.go:95`).
  The probe goes next to `loom-probe keys` (`cmd/loom-probe/main.go:33`, `keys.go:62`, `echoKeys`
  `:133`).

## 3. Implementation & Verification Plan

### M1 — Canary probe (`loom-probe mods`)
- A new `loom-probe mods` subcommand next to `keys`, in raw mode. After each read it samples
  `/run/voxi/modifiers` and prints one line: quoted raw bytes, `DecodeKey` result, the mask in hex,
  and decoded names (`lctrl rctrl lalt ralt lsuper rsuper lshift rshift`). With `--path` it reads
  another file. A missing or unreadable file gives a clear message and the plain `keys`-style
  output.
- A small mask parser in its own file, unit-tested with a fake state file (bit table, missing file,
  empty file). No PTY test.
- Files: `cmd/loom-probe/main.go`, a new `cmd/loom-probe/mods.go`, and `cmd/loom-probe/main_test.go`
  or a new `mods_test.go`. Do not touch the decoder or the spec in M1.
- Gate: the user runs `loom-probe mods` in the Gram terminal and presses `^1`..`^9`, `^2`/`^3`
  and plain digits. M2 starts only if every ctrl-digit shows the ctrl bit set and every plain digit
  shows it clear. Otherwise, close as won't fix with the results.
- M1 delivered (2026-10-30): `loom-probe mods [--path]`; mask parser in `cmd/loom-probe/mods.go`
  with tests; awaiting the Gram gate.

### M2 — Library feature (after the M1 gate)
- Spec `pane.modifier_state` (off by default) in YAML, schema and `specdefaults`, plus a session
  consumer. An injectable reader interface, so tests use a fake file. Re-interpret only on
  non-kitty terminals and only the ambiguous legacy bytes when the sampled mask agrees.
- Unit tests for the re-interpretation table, including near misses (ctrl bit clear, shift held,
  kitty active). `make test-q1`, then `make install`.

## Decision

No build tag for `loom-probe mods`: a missing state file falls back to plain key output on any OS.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| audit | jules-audit-389 | 7007712254790743791 | GH #84 | kept: partly done, 4/4 refs hold (lines within 10 on 24e2214); M1 probe in `cmd/loom-probe/mods.go`, M2 (`modifier_state` spec and consumer) missing: host search empty; report `reports/jules/389.md` |
