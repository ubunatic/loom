# 074 — Use measured WINCH speed for adaptive resize width guard

**Status**: In Progress — M1-M4 done; manual terminal check (M5) pending
**Priority**: P2
**Severity**: Moderate
**Category**: Feature
**Related**: [Issue 073](073-add-configurable-winch-resize-diagnostics-app-and-spec-backed-rendering-modes.md)

---

## 1. Problem & Motivation

Issue 073 added a fixed resize width guard, rendering at `W-n` during an active
resize burst and restoring full width after the burst settles. A fixed `n` is
useful for comparison, but different terminals and resize gestures produce
different `SIGWINCH` rates and may need different protection margins.

Add WINCH-rate measurement to the framework and `examples/winch`. Expose a
switch that uses the measured WINCH speed to select the width guard instead of
the manually configured `n`.

## 2. Technical Specification / Findings

The measurement should be based on recent `SIGWINCH` arrival timing, not on
render-frame rate. The app should display the current rate and enough state to
understand when the adaptive guard is active. The manual `n` remains available
when adaptive mode is disabled.

Define and document the mapping from measured WINCH speed to guard width,
including minimum and maximum bounds, smoothing/window duration, startup
behavior, and the behavior when no recent WINCH events exist. Avoid making the
mapping terminal-specific or duplicating constants outside `spec/`.

## 3. Implementation & Verification Plan

### M1 — Specify WINCH measurement and adaptive policy

- Add spec-backed settings for measurement window, smoothing, minimum/maximum
  guard width, and speed-to-width mapping.
- Define the `Use WINCH Speed` switch and its default.
- Add positive and negative spec validation.

Verification: spec validator accepts valid settings and rejects invalid bounds
or mappings.

### M2 — Measure WINCH speed in the framework

- Track timestamps when `SIGWINCH` notifications are received.
- Expose a read-only rate/state API without introducing a tty dependency.
- Compute the adaptive guard width from the spec-defined policy.

Verification: deterministic unit tests cover no events, one event, steady
bursts, changing rates, and clamping.

### M3 — Integrate adaptive width guarding

- When `Use WINCH Speed` is enabled, use the measured adaptive width instead
  of manual `n` during the active guard interval.
- Preserve manual `n` behavior when disabled.
- Keep the one-second settle/restore behavior and full-width rendering after
  the burst.

Verification: PTY tests compare manual and adaptive modes across resize bursts.

### M4 — Add Winch diagnostics

- Add a visible WINCH speed meter (events/sec and/or interval) to
  `examples/winch`.
- Add keyboard control for `Use WINCH Speed` and display whether the current
  guard width is manual or adaptive.
- Show effective guard width, configured `n`, and measured speed together.

Verification: headless app tests cover switch toggling and display state;
manual testing exercises slow and rapid terminal drags.

### M5 — Verify and document behavior

- Run `make test-q1` and PTY resize coverage.
- Test across at least one slow resize drag and one rapid resize burst.
- Document the policy, controls, and known terminal limitations.

---

## 4. Outcome & Findings

M1–M4 implemented; M5 manual terminal checks still pending.

- **Spec (M1)**: `adaptive_guard` mode in `spec/resize.yaml` (key `a`, default off) with `window_ms`, `min_n`, `max_n`, `rate_step`; schema fields plus a negative control (`invalid_rate_step`); the Go loader rejects `min_n > max_n`.
- **Policy**: rate = events in the last `window_ms` / window (the window is the smoothing); `n = min_n + floor(rate / rate_step)`, clamped to `max_n`. Fewer than two events in the window (startup, single resize, settled) fall back to the manual `n`. The width is latched at each SIGWINCH so it does not jitter while the rate decays.
- **Framework (M2/M3)**: `WinchMeter` takes explicit timestamps (deterministic tests: none, one, steady, saturated, aged-out). `Pane.EffectiveGuardN`, `Pane.WinchRate`, and one `guardedCols` helper replace three duplicated width computations.
- **Winch (M4)**: key `a`, effective n / manual n / adaptive-or-manual / `WINCH: x/s` on the status line; headless and PTY tests.
- **Limitation**: the kernel coalesces pending SIGWINCH, so very fast drags under-report the raw rate.
- **Pending (M5)**: manual slow-drag and rapid-drag checks in a real terminal; tune `rate_step`/`max_n` from those.
