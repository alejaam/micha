# Design: UX Shared Finance Ritual v1

## Technical Approach

Implementación incremental en frontend, centrada en comportamiento:  
1) guardrails de captura en `ExpenseModal`,  
2) gobernanza de periodo en `useHouseholdData` + `PeriodStatusRibbon`,  
3) claridad de histórico provisional en `useHistoricalPeriods` + paneles consumidores,  
4) unificación de copy en páginas core.  
Esto implementa las specs `shared-finance-ux-ritual` y `period-governance-ux` sin romper rutas ni contratos API existentes.

## Architecture Decisions

| Decision | Options | Tradeoff | Choice |
|---|---|---|---|
| Guardrail para `fixed` | Permitir en modal / ocultar en modal y redirigir a setup | Permitir genera ambigüedad de negocio | Ocultar/bloquear `fixed` en modal y guiar a `/onboarding/fixed-expenses` |
| Bloqueo por periodo | Bloquear solo backend / bloquear backend + UX explícita | Solo backend da mala UX | Mantener backend como fuente de verdad y agregar mensajes accionables en UI |
| Historial provisional | Silenciar fallback / mostrar etiqueta provisional | Silenciar confunde decisiones financieras | Mostrar estado provisional + motivo + siguiente paso recomendado |
| Estrategia de cambio | Refactor completo / cambios quirúrgicos | Refactor completo aumenta riesgo | Cambios quirúrgicos en componentes y hooks actuales |

## Data Flow

`AppShell(periodStatus)` → `useHouseholdData(isMutationLocked)` → páginas (`Dashboard/Expenses/Balances/Rules`) → `ExpenseModal`  
`useHistoricalPeriods(getSettlement)` → `selectedPeriodSnapshot + provisionalReason` → `HistorySection` / gráficos

```text
User Action
   │
   ▼
Page CTA/FAB ──→ useHouseholdData.handleCreate/Save/Delete
   │                         │
   │                         ├─ if locked: UI error + stop
   │                         └─ if open: API mutation + reload
   ▼
ExpenseModal guardrails (type/paid_by) ──→ payload validado

Historical view ──→ useHistoricalPeriods ──→ API settlement
                                  │
                                  └─ fallback provisional + explicit hint
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `frontend/src/components/ExpenseModal.jsx` | Modify | Eliminar/inhabilitar `fixed` en flujo general, copy en español, claridad de elegibilidad en “pagado por”. |
| `frontend/src/hooks/useHouseholdData.jsx` | Modify | Unificar mensajes de bloqueo por `review/closed` y recomendación de siguiente acción. |
| `frontend/src/components/PeriodStatusRibbon.jsx` | Modify | Mensajes de estado orientados a decisión de usuario. |
| `frontend/src/hooks/useDashboardUxState.js` | Modify | Texto del mapa de estado consistente con negocio y idioma. |
| `frontend/src/hooks/useHistoricalPeriods.js` | Modify | Estandarizar `provisionalReason` y metadatos de confiabilidad. |
| `frontend/src/components/HistorySection.jsx` | Modify | Mostrar indicador provisional y CTA de refresco/validación. |
| `frontend/src/pages/{DashboardPage,ExpensesPage,BalancesPage,RulesPage}.jsx` | Modify | Unificar copy core y priorizar CTA del ritual colaborativo. |
| `frontend/src/components/AppHeader.jsx` | Modify | Ajustar labels de navegación/control a terminología de negocio en español. |
| `frontend/src/components/__tests__/DashboardPage.integration.test.jsx` | Modify | Actualizar assertions de copy y flujos bloqueados. |
| `frontend/src/components/__tests__/PeriodStatusRibbon.test.jsx` | Modify | Validar mensajes nuevos de estados de periodo. |
| `frontend/src/pages/__tests__/OnboardingFixedExpensesPage.test.jsx` | Modify | Garantizar guía de fixed en setup dedicado. |

## Interfaces / Contracts

No se crean endpoints nuevos. Se mantiene contrato con:
- `createExpense(...)` (`frontend/src/api.js`)
- `getSettlement(...)` (`frontend/src/api.js`)

Contrato UI interno esperado:

```ts
type PeriodLockState = 'open' | 'review' | 'closed'
type ProvisionalMeta = { isProvisional: boolean; provisionalReason: string }
```

Reglas:
- `PeriodLockState !== 'open'` ⇒ mutaciones bloqueadas en UI.
- `expenseType === 'fixed'` en alta general ⇒ no permitido; redirigir a setup dedicado.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Guardrails de `ExpenseModal`, estado de `buildRibbonState`, razón provisional | Vitest de componentes/hooks con mocks de API |
| Integration | Flujo bloqueado por periodo, copy consistente, navegación a setup fijo | RTL tests en páginas (`Dashboard`, `Expenses`, `Rules`) |
| E2E | No aplica en esta fase | Sin suite E2E instalada; cubrir con integración + smoke manual |

## Migration / Rollout

No migration required. Rollout directo en frontend.  
Si hay regresión de copy/flujo, rollback por módulo vía git revert del change.

## Open Questions

- [ ] ¿El owner debe poder registrar gastos para cualquier miembro en todos los estados o solo en `open`?
- [ ] ¿Qué CTA exacto de “siguiente acción” prefieren para datos históricos provisionales?
