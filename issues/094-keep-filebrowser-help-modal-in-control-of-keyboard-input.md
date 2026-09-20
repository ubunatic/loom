# 094 — Keep filebrowser help modal in control of keyboard input

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Bug
**Related**: None

---

## Problem & Motivation

In the filebrowser, entering `:help<CR>` opens the help dialog, but subsequent
key presses are routed to the parent application and can exit the app instead of
being handled by the modal. This makes the help screen unsafe to use.

## Scope

Give the help dialog modal focus and exclusive keyboard ownership while it is
visible. Any key should be processed by the dialog's own navigation, dismissal,
or no-op behavior; it must not trigger filebrowser or application-level quit,
close, or other actions. Restore the prior filebrowser focus and input routing
when the dialog closes.

Define behavior for dismissal keys, scrolling/navigation keys, unknown keys,
and repeated help-dialog activation. Preserve normal filebrowser shortcuts once
the modal is gone.

## Goal

/goal: Ensure `:help<CR>` opens a safe modal help dialog where every subsequent
key is handled by the dialog and cannot exit or mutate the parent application,
with correct focus restoration on close.

## Verification

Add a regression test that opens help, sends quit and other application-level
keys, confirms the app remains open and the dialog owns input, then closes the
dialog and verifies normal filebrowser shortcuts work again.
Extend the PTY tests to reproduce `:help<CR>`, send representative keys, and
prove the modal—not the parent app—controls input throughout the session.
