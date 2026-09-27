# 149 — Convert the media example to a Cobra CLI app

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature / CLI Tooling
**Related**: [143](143-make-the-media-example-full-width-and-play-video.md)

---

## 1. Problem & Motivation

`examples/media` uses a hand-written argument parser, so it lacks the standard help, flag parsing, and command behavior used by Cobra-based Loom examples.

## 2. Technical Specification / Findings

Convert the media example to a Cobra CLI app. Preserve its existing media-path and render-mode behavior while exposing the command's usage and options through Cobra help.

## 3. Implementation & Verification Plan

- **/goal**: Make `examples/media` a Cobra-based CLI with documented usage and preserved media display behavior, or stop and report when blocked on a user decision or denied permission.
