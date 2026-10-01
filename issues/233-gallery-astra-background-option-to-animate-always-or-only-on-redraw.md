# 233 — Gallery astra background: option to animate always or only on redraw

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 231 (F8 background switch), 067, 068

---

/goal Give the widget gallery a user-selectable astra animation mode — (a) animate only when other
redraws happen, (b) always animate — with tests, or stop and report when blocked on a user decision
(e.g. key binding, default mode) or denied permission.

## 1. Problem & Motivation
In the widget gallery (`loom widgets --show`), the "astra" background (switched on with F8, issue 231)
only animates when another part of the screen redraws (spinner, timer, ...) or when the mouse moves.
The user finds both behaviours useful and wants them as an option:

- **a) on-redraw**: animate only when other redraws are needed — saves CPU/terminal output.
- **b) always**: animate continuously at the background's own cadence — looks nicer.

## 2. Technical Specification / Findings
Likely cause (verify against live code): `pane.go` (~line 812) creates the background ticker only
once, at run start, and only if `p.Background` is already an `AnimatedBackground`. The gallery starts
with "plain" (nil) and switches to astra at runtime via `setBackground`, so no ticker exists and astra
frames are only drawn as a side effect of other redraws. Mode (b) therefore needs the pane to
start/stop/re-arm the background ticker when the background changes — a library fix in `pane.go`, not
a gallery workaround (per AGENTS.md "fix the library, not the caller").

Open choices, decide or ask:
- How the user picks the mode: e.g. extend the F8 cycle (plain → astra → astra (always)), a separate
  key, or a CLI flag. Show the active mode in the status bar.
- Which mode is the default.
- Whether the mode is a general `Pane` option (likely) or gallery-only.

## 3. Implementation & Verification Plan
- Pane: re-evaluate the background ticker on background change; add an option selecting
  on-redraw vs. always; respect `ReduceMotion`.
- Gallery: expose the mode choice and show it in the status bar.
- Tests: switching to astra at runtime in "always" mode produces ticks without other redraws;
  "on-redraw" mode produces none. Run `make test-q1`, `make install`.
