# Constitución SDD (Hermes)

Este repo sigue un flujo tipo **Spec Kit (SDD)** ejecutado por Hermes.

## Artefactos

- `.specify/features/<feature>/spec.md`
- `.specify/features/<feature>/plan.md`
- `.specify/features/<feature>/tasks.md`

## Reglas

1) No implementar nada sin `spec.md` aprobado.
2) No implementar nada sin `plan.md` aprobado.
3) `tasks.md` debe ser trazable a los escenarios/criterios del spec.
4) En implementación: preferir cambios pequeños y verificables (tests / checks / lint).
5) No exponer secretos (redactar como `[REDACTED]`).
