# 164 — Settings: KindNumber with min/max/step and ←/→ stepping

**Status**: Open
**Priority**: P1 (High)
**Severity**: Moderate
**Category**: Feature
**Related**: settings.go; consumer: ../settings issue 002/006, mockups ../settings/docs/data/03-edit-font-size.ansi, 05-invalid-value.ansi

---

## 1. Problem & Motivation
The `settings` app (foot font size first) needs numeric settings. `Settings` only has
KindBool, KindString, KindChoice. A canary with KindString showed:
- any text is accepted ("a99" commits fine), no range check;
- ←/→ do not step a value; → enters edit mode, ← does nothing;
- no way to show a validation error.

## 2. Technical Specification / Findings
Proposed `KindNumber`:
- fields `Num *float64`, `Min`, `Max`, `Step float64` (Step 0 → 1), optional `Format` (default `%g`);
- ←/→ (and -/+) change the value by Step, clamped to [Min, Max], without entering edit mode;
- Enter opens inline edit that accepts digits and one `.` only; Enter commits if parsed value is in range,
  otherwise stays in edit mode and shows an error line (e.g. "99 is above the maximum 72");
- rendered as `◂ 12 ▸` like the consumer mockup.

## 3. Implementation & Verification Plan
- Implement in settings.go, table tests via HandleKey for step, clamp, typed valid/invalid, cancel.
- Update the settings demo/example if one lists kinds.
