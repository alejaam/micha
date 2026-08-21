# Domain Docs

How the engineering skills should consume this repo's domain documentation.

## Before exploring, read these

- `CONTEXT.md` at the repo root.
- Relevant ADRs under `docs/adr/`.

If either is missing, proceed silently.

## File structure

Single-context repo:

```text
/
├── CONTEXT.md
└── docs/adr/
```

## Use the glossary's vocabulary

Use terms as defined in `CONTEXT.md`. If a required concept is missing,
reconsider the terminology or note the gap for `/domain-modeling`.

## Flag ADR conflicts

Surface contradictions with existing ADRs instead of silently overriding them.
