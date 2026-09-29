# 176 — Add loom widgets command listing all widgets with examples and a completeness test

**Status**: Closed — resolved
**Priority**: P1 (High)
**Severity**: Moderate
**Category**: Feature
**Related**: cmd/loom/main.go, docs/Widgets.md §13 (input inventory), issues/172-*.md, issues/174-*.md

---

## 1. Problem & Motivation
On 2026-09-29 an agent answered "which input widgets do we have?" from memory and missed `Settings` kinds (`KindNumber`, `KindBool`, `KindChoice`), `examples/filebrowser` `NavigationPane`, and invented a nonexistent `FileOpener`, which a consumer (settings wizard) was then told to use. Widgets are spread over the root package, `graph/`, `media/` and example packages, with no single authoritative list.

## 2. Technical Specification / Findings
- `loom widgets` prints every widget: name, one-line purpose, category (input, display, layout, infra), a minimal code example, and ref pointers (source file, docs/Widgets.md section, ticket).
- `loom widgets <name>` shows one entry in full; output should be greppable/plain by default.
- Single source: a spec file (per docs/Spec.md, e.g. `spec/widgets.yaml`) that the command and docs render from; Go must not duplicate it.
- Completeness test: enumerate exported types that implement `loom.Widget` (via `go/types` or `go/packages` over the root, `graph`, `media` and reusable example packages) and fail when one is missing from the catalog, or when a catalog entry names a type that no longer exists. Record deliberate exclusions (internal helpers) explicitly in the spec.
- Include sub-widget capabilities that are easy to miss (e.g. `Settings` row kinds).
- Scope (decided 2026-09-29): library packages only. Widgets in example packages (e.g. `NavigationPane`) are excluded; they are to be moved into the library later (see 172), and join the catalog then.

## 3. Implementation & Verification Plan
/goal Never lose track of our own widgets again: `loom widgets` lists all widgets with minimal examples and refs from one spec, and `make test` fails when a Widget type is added or removed without updating the catalog. Stop and report when blocked on a user decision or denied permission.

## Outcome (2026-09-29)

- `loom widgets` lists every library widget from `spec/widgets.yaml` (embedded, schema-validated); `loom widgets <name>` shows one entry and rejects unknown or ambiguous names.
- `widgets_catalog_test.go` fails when an exported type with `Draw(*Canvas, Rect[, bool])` + `HandleKey(KeyEvent) bool` in a library package is missing from the spec or vice versa; this includes embedded controls (`TextInput`, `TextArea`). Named exclusions must still exist.
- `docs/Widgets.md` §13 now points to the catalog instead of a hand-kept table.
- Built by rs176 (codex luna, stalled before commit), finished by the host; reviewed by agy:flash38 (no blockers).
- Known limit: detection is AST-based, so widgets that only get their methods through struct embedding, or import loom under an alias, are not detected. Switch to `go/types` if that case appears.

## M2 — Catalog output polish (follow-up, 2026-09-29)

Acceptance criteria:
1. `loom widgets` (no args) prints one line per widget: name, category, purpose; grouped by category (input, display, layout, infra), sorted by name within each group. `loom widgets <name>` keeps the full entry (example, source, docs, ticket).
2. Entries in `spec/widgets.yaml` are sorted by name, and a test enforces that order.
3. No entry points `docs:` at `docs/Widgets.md §13` (the catalog section itself); make `docs` optional in the schema and drop those values; only print `Docs:` when set.
4. `ticket:` values use the short form `issues/NNN` everywhere.
5. Update `cmd/loom/main_test.go` for the new list format; run `make test-q1` once to a file and grep `--- FAIL`; `make install`; commit `feat(cmd): ... (issue 176 M2)`.

M2 delivered (fc8e58c, agy:flash37 dev): compact grouped list, name-sorted spec with order test, §13 self-pointers removed, short ticket refs.
