# 469 — Form lets a focused TextArea handle Enter, Down and PgDn before form navigation

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: 458, 470, 169, 277

---

## 1. Problem & Motivation
A TextArea inside a Form cannot be used for multi-line text. `Form.ConsumeKey` matches its own
bindings first and hands a key to the focused field only when none matched (`form.go:276-303`;
`spec/defaults.yaml` `form.bindings`: next `tab, down, pgdown, …`, submit `enter`). So in a
focused TextArea, Down moves focus to the next field and Enter submits the whole form. 458
batch 6 canary (`docs/studies/tuios/report-widgets.md`): Down → focus 1, TextArea row 0; Enter
→ `submits=1`, value `"one\ntwo"` unchanged.

## 2. Technical Specification / Findings
- Fix the library, not the caller: a focused field gets the key first; the Form acts only on
  `Ignored()`. TextArea then keeps Enter, Down and PgDn while it can use them (Down on the last
  line may be ignored and move focus, as a desktop form would).
- Tab stays the container's key (Tab policy, `docs/Widgets.md`); `AdvanceOnEnter` for
  single-line TextInput keeps working because TextInput ignores Enter.
- Check other containers with the same "own keys first" order (Stack, tabs, dialogs).
- Event routing: by project rule this starts on terra:med; the owner's Codex quota decides
  (no Codex for 458 from batch 6 on).

## 3. Implementation & Verification Plan
/goal In a Form, a focused TextArea receives Enter, Down and PgDn and edits with them, while
Tab, Esc and keys the field ignores still drive the Form, with tests; stop and report when
blocked on an API decision or denied permission; on a loop or no progress, clear the goal and
write a failure report.

- Re-check live code first (`form.go` ConsumeKey, field widgets' EventResult).
- Port the batch 6 canary (`scratchpad/b6-canary`, `TestFormCapturesDownAndEnterInTextArea`)
  into a loom test; table test per key × field type (TextInput, TextArea, Choice).

## Jules test-draft (GH #123, 2026-11-10)
Session jules-routing-tests-b (Jules ID 10199914946274802065) added `form_field_keys_test.go`
(`TestFormCapturesDownAndEnterInTextArea`). Host run on the PR head: fails as the bug predicts
(Enter in a TextArea submits the form instead of inserting a line; Down/PgDn move form focus instead
of the caret). The Choice subtest's failure message prints the whole widget struct; shorten it when
adopting. Trace references (`form.go:271-296`) hold. Pick the test from `git fetch github
pull/123/head`.
