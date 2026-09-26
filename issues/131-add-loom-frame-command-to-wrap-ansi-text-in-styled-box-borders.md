# 131 — Add loom frame command to wrap ANSI text in styled box borders

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Feature / CLI Tooling
**Related**: `cmd/loom/`, `ansibuffer.go`, `canvas.go`

---

## Goal

`/goal`: Implement `loom frame [file.ansi]` (with `--style single|double|rounded|heavy`, `--title "..."`, `--padding N`, and `--color <color>`) in the `loom` CLI, enabling automatic framing and box border wrapping around arbitrary ANSI and text artwork without manual row-by-row border construction.

## 1. Context & Motivation

- When creating new ANSI banners, logo cards, and dialog mockups, developers often have the raw text/glyph body and need to wrap it inside a pixel-aligned frame border with a header title or color scheme.
- Python scripts were previously written to construct top borders `┌───┐`, inject vertical walls `│...│`, and append bottom borders `└───┘`.
- Adding `loom frame` provides a native CLI tool to wrap any text or ANSI buffer inside standard Loom box geometries.

## 2. Command Specification

- **`loom frame <file.ansi>`**:
  - `--style <type>`: `single` (`┌─┐`), `double` (`╔═╗`), `rounded` (`╭─╮`), or `heavy` (`┏━┓`).
  - `--padding <N>`: Adds horizontal and vertical padding around the content before framing.
  - `--color <color>`: Applies ANSI foreground color to border characters (e.g. `208`, `cyan`, `orange`).
  - `--title <text>`: Embeds a centered or left-aligned title badge in the top border.
  - `--write` / `-w`: Modifies file in place or outputs to stdout.
