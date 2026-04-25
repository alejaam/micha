# Proposal: UX Shared Finance Ritual v1

## Intent

Mejorar la UX para parejas/roomies, alineando la app al ritual real de finanzas compartidas: registrar, conciliar y cerrar periodo con claridad. Hoy hay fricción por lenguaje inconsistente, reglas difusas de captura y señales provisionales en histórico.

## Scope

### In Scope
- Unificar microcopy crítico en español (dashboard, gastos, balances, reglas, onboarding).
- Alinear captura de gastos a reglas de negocio: `fixed` solo desde setup dedicado y mejor claridad de `paid_by_member`.
- Reforzar estado de periodo (open/review/closed) y bloqueos de acciones.
- Mejorar historial/cierre con estado provisional y siguiente acción visible.

### Out of Scope
- Reescritura completa del design system.
- Backend completo de consenso/aprobaciones de cierre.
- Integraciones bancarias.

## Capabilities

### New Capabilities
- `shared-finance-ux-ritual`: UX orientada al ciclo colaborativo de hogar (captura → conciliación → cierre).
- `period-governance-ux`: comunicación de estado de periodo y reglas de bloqueo accionables.

### Modified Capabilities
- None

## Approach

Aplicar mejoras incrementales en frontend con enfoque “behavior-first”: primero reglas y claridad de flujo, luego refinamiento visual. Reusar componentes actuales (`ExpenseModal`, `PeriodStatusRibbon`, `HistorySection`) y ajustar copy, estados y validaciones según `MICHA_VISION.md`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `frontend/src/components/ExpenseModal.jsx` | Modified | Restricciones para tipo de gasto y “paid by”. |
| `frontend/src/hooks/useHouseholdData.jsx` | Modified | Reglas y mensajes de bloqueo por estado de periodo. |
| `frontend/src/components/PeriodStatusRibbon.jsx` | Modified | Mensajes de estado orientados a acción del usuario. |
| `frontend/src/hooks/useHistoricalPeriods.js` | Modified | Señalización explícita de datos provisionales e historial. |
| `frontend/src/pages/{DashboardPage,ExpensesPage,BalancesPage,RulesPage}.jsx` | Modified | CTA y copy para flujo colaborativo. |
| `frontend/src/pages/Onboarding*.jsx` | Modified | Continuidad del onboarding para hogar compartido. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Regresión en permisos al ajustar “paid_by_member” | Med | Tests de integración owner/member en creación de gasto. |
| Confusión temporal por cambios de copy | Low | Cambios graduales y consistentes en pantallas core primero. |
| Dependencia de backend para consenso real | High | Diseñar UX progresiva con estados “provisional” explícitos. |

## Rollback Plan

Revertir cambios frontend por módulo (modal, period status, historial y páginas) y restaurar copy/validaciones previas con git revert del change `ux-shared-finance-ritual-v1`.

## Dependencies

- Confirmar contratos backend actuales de `expenses`, `settlement` y permisos por rol.

## Success Criteria

- [ ] El flujo de gasto evita ambigüedad entre variable/fixed/MSI y reduce errores de captura.
- [ ] Los bloqueos por periodo se entienden sin soporte adicional.
- [ ] El historial indica claramente cuándo los datos son provisionales.
- [ ] La UX se percibe consistente con finanzas compartidas en pareja/roomies.
