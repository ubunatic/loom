# 169 — Add Form and FieldGroup composite layout widget with tab navigation and validation summary

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: settings.go, textinput.go, confirm.go, issues/155-built-in-modal-and-dialog-overlay-primitive.md

---

## 1. Problem & Motivation
Loom provides individual input widgets (`TextInput`, `Settings`, `Confirm`). However, complex interactive TUIs often need structured multi-field forms (e.g. login/auth prompts, database connection configuration, wizard dialogs, git commit metadata):
- `Settings` is vertically oriented with key/value rows and its own inline editing state machine, which doesn't fit freeform multi-input forms with submit/cancel buttons or custom field layouts.
- Applications currently have to manually wire `Tab`/`Shift-Tab` focus cycling between multiple `TextInput` widgets, manage button bars, and collect validation errors before submit.

## 2. Technical Specification / Findings
Introduce `loom.Form` and `loom.FormField` (implementing `loom.Widget`, `loom.EventConsumer`, `loom.MouseConsumer`, `loom.FocusContainer`):
- **Structure**:
  ```go
  type FormField struct {
      Label     string
      Widget    loom.Widget // TextInput, Choice, Settings item, Checkbox
      Help      string
      Validate  func() error
      Required  bool
  }

  type Form struct {
      Fields     []FormField
      Actions    []FormAction // e.g. Submit, Reset, Cancel buttons
      OnSubmit   func(values map[string]any)
      OnCancel   func()
      Validation map[int]string
  }
  ```
- **Navigation & Focus**:
  - `Tab` / `Down`: Move focus to next form field or action button.
  - `Shift-Tab` / `Up`: Move focus to previous field.
  - `Enter`: Submit form if all validators pass; otherwise focus first invalid field and render field-level error messages.
  - `Esc`: Trigger `OnCancel`.
- **Rendering**:
  - Aligned field labels and inputs.
  - Required marker indicators (`*`).
  - Validation error text beneath invalid fields or collected in a top/bottom summary callout.
  - Submit / Cancel action buttons.

## 3. Implementation & Verification Plan
- Create `form.go` and `form_test.go`.
- Unit tests:
  - Tab and Shift-Tab focus traversal across text inputs, dropdowns, and buttons.
  - Validation execution on Enter and error banner presentation.
  - Successful submission triggering `OnSubmit` callback with collected values.
- Document in `docs/Widgets.md`.
