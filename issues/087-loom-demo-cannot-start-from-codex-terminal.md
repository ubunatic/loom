# 087 — loom-demo cannot start from Codex terminal

**Status**: Closed — resolved: launch loom-demo through foot for a controlling TTY
**Priority**: P1
**Severity**: Major
**Category**: Bug
**Related**: None

---

## 1. Problem & Motivation

`loom-demo` cannot be launched from the terminal environment used by Codex, even
when an interactive terminal is opened with `foot`. This prevents exercising the
demo locally in the same environment where other terminal applications work.

## 2. Technical Specification / Findings

From the Codex environment:

```text
$ foot -- loom-demo
loom: open /dev/tty: open /dev/tty: no such device or address
```

The terminal emulator itself works: `foot -- mc` launches successfully. Determine
whether Loom's startup/TTY detection needs to support this launch context or
whether the documented invocation needs adjustment.

## 3. Implementation & Verification Plan

/goal: Launch `loom-demo` successfully from the Codex terminal environment, with
the interactive Loom UI attached to a usable TTY and no regression for ordinary
terminal launches.

- Reproduce the failure and inspect the startup path that opens `/dev/tty`.
- Implement the smallest compatible fix or document the required invocation if
  the environment is intrinsically unsupported.
- Verify with the relevant startup tests and by launching `loom-demo` through
  `foot`.
