# 176 — Add loom widgets command listing all widgets with examples and a completeness test

**Status**: Open
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
