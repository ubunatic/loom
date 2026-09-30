# 218 — TextInput demo: only the first input is usable

**Status**: Closed — fixed with tests, suite green
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: 204, 207, 209

---

## 1. Problem & Motivation
User feedback: in the TextInput demo only the first input works; the others cannot be clicked, and Enter in the first does not advance to the next.

## 2. Technical Specification / Findings
Clicking an input must focus it; Enter in a single-line input must move focus to the next input (form behavior). Fix in the library (focus/mouse routing), not the demo.

## 3. Implementation & Verification Plan
PTY tests: click the 2nd and 3rd inputs and type into them; Enter in the 1st moves focus to the 2nd.

/goal All TextInput demo fields are clickable and Enter advances focus, proven by PTY tests; or stop and report when blocked on a user decision or denied permission.
