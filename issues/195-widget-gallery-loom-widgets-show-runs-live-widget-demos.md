# 195 — Widget gallery: loom widgets --show runs live widget demos

**Status**: Open
**Priority**: P1 (High)
**Severity**: Minor
**Category**: Feature
**Related**: 180, 102

---

## 1. Problem & Motivation

The roadmap (180) adds many widgets. `loom widgets [name]` lists widgets and prints usage,
but nothing lets a user or reviewer try a widget live. The user asked for a widget gallery
callable as `loom widgets --<flag> <widgets...>`, and for `examples/` apps when needed.

## 2. Sprint goal

- New package `gallery/` (importable, no cmd deps): a registry of named demo constructors,
  one per widget, each returning a ready-to-run root Widget with sample data.
- `loom widgets --show <name...>` runs the named demos in one Pane (Tabs when more than
  one); `--show` with no names shows all. `--list` keeps the plain listing.
- Seed demos for the existing widgets that are easy to show (TextInput, TextArea,
  ProgressBar, Table, Tabs, Popup, Pill, Choice).
- A test asserts every gallery name exists in `spec/widgets.yaml` and every demo renders
  a non-empty frame at 80x24 without a terminal.
- Save a render of the gallery as `docs/progress/gallery.ansi`.

Every later roadmap sprint that adds or changes a widget adds or updates its gallery demo;
a full app goes in `examples/` only when the widget needs a realistic context.

/goal Ship the gallery package and `loom widgets --show`, tests green, `make install`
done; stop and report if the existing `widgets` command contract would break.
