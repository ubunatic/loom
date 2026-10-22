# 366 — Spec shadows: collector, graph and emoji measure defaults duplicate YAML values

**Status**: Closed — Collector retention, graph background/max width and emoji measure defaults now read from embedded YAML; constants became vars (Upgrading row, API snapshot); regression test changing YAML failed before (init panic) and passes (985e7ee); host q1 green
**Priority**: P2 (Medium) — editing the YAML panics at init
**Severity**: Minor
**Category**: Bug
**Related**: docs/Spec.md (no shadowing)

---

/goal Each value below has one source, the embedded spec; Go code reads it and no Go literal mirrors it, and
changing the YAML value changes behaviour instead of panicking; tests prove both; or stop and report when
blocked on a user decision.

## Findings (spec audit, 2026-10-22)

- `collector/collector.go:31` retention `15m` = `spec/defaults.yaml` `collector.default_retention`.
- `graph/options.go:15` background `"100"` and `graph/bar.go:17` max width `10` = `spec/defaults.yaml`
  graph keys. Init panics when YAML and constant differ.
- `measure/spec.go:81` default width `2` and modes `pad-1`/`default` repeat `spec/emoji.yaml`, also when
  `measure` is imported alone.

## Notes

- Exported constants may become vars or functions; this breaking change is approved by the host. Add a
  row to `docs/Upgrading.md` and refresh the API snapshot.
- Out of scope (later, own ticket if wanted): moving bindings, labels, glyphs, colours and timings that
  are not in the spec yet.
