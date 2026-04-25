# Tasks: UX Shared Finance Ritual v1

## Phase 1: Foundation (Spec-to-UI contracts)

- [x] 1.1 Definir matriz de copy objetivo (ES) en `frontend/src/pages/{DashboardPage,ExpensesPage,BalancesPage,RulesPage}.jsx` y `frontend/src/components/AppHeader.jsx`.
- [x] 1.2 Establecer reglas UX de periodo en `frontend/src/hooks/useDashboardUxState.js` (open/review/closed + texto accionable).
- [x] 1.3 Definir contrato de guardrail de `expenseType` en `frontend/src/components/ExpenseModal.jsx` (sin `fixed` en alta general).

## Phase 2: Core Implementation (Business-first UX)

- [x] 2.1 Modificar `frontend/src/components/ExpenseModal.jsx` para remover/bloquear opción `fixed` y agregar CTA hacia `/onboarding/fixed-expenses`.
- [x] 2.2 Ajustar elegibilidad y microcopy de “pagado por” en `frontend/src/components/ExpenseModal.jsx` (owner/member con contexto claro).
- [x] 2.3 Unificar mensajes de bloqueo por periodo en `frontend/src/hooks/useHouseholdData.jsx` para create/save/delete.
- [x] 2.4 Actualizar `frontend/src/components/PeriodStatusRibbon.jsx` para comunicar estado + impacto en acciones.
- [x] 2.5 Ajustar `frontend/src/hooks/useDashboardUxState.js` para textos consistentes y orientados a decisión.
- [x] 2.6 Mejorar señalización provisional en `frontend/src/hooks/useHistoricalPeriods.js` (motivo claro y metadata coherente).
- [x] 2.7 Mostrar indicador provisional + siguiente paso en `frontend/src/components/HistorySection.jsx`.
- [x] 2.8 Unificar CTAs y copy de flujo colaborativo en `frontend/src/pages/DashboardPage.jsx`.
- [x] 2.9 Unificar CTAs y copy de flujo colaborativo en `frontend/src/pages/ExpensesPage.jsx`.
- [x] 2.10 Unificar CTAs y copy de flujo colaborativo en `frontend/src/pages/BalancesPage.jsx` y `frontend/src/pages/RulesPage.jsx`.
- [x] 2.11 Ajustar navegación/terminología en `frontend/src/components/AppHeader.jsx`.

## Phase 3: Integration / Wiring

- [x] 3.1 Verificar navegación completa desde modal de gasto hacia setup fijo (`/onboarding/fixed-expenses`).
- [x] 3.2 Validar coherencia de bloqueos de mutación entre `Dashboard`, `Expenses`, `Balances` y `Rules`.
- [x] 3.3 Validar que histórico provisional se refleje igual en panel histórico y vistas consumidoras.

## Phase 4: Testing (TDD RED → GREEN → REFACTOR)

- [x] 4.1 RED: ampliar `frontend/src/components/__tests__/PeriodStatusRibbon.test.jsx` con escenarios de copy/estado bloqueado.
- [x] 4.2 GREEN: implementar cambios en ribbon/ux-state hasta pasar pruebas.
- [x] 4.3 RED: ampliar `frontend/src/components/__tests__/DashboardPage.integration.test.jsx` para bloqueo por periodo y CTA colaborativos.
- [x] 4.4 GREEN: implementar cambios en páginas core hasta pasar pruebas.
- [x] 4.5 RED: agregar tests en `frontend/src/components/__tests__/ExpenseModal.test.jsx` para guardrail de `fixed` y mensajes de “pagado por”.
- [x] 4.6 GREEN: implementar cambios de modal hasta pasar pruebas.
- [x] 4.7 REFACTOR: limpiar duplicación de mensajes/labels en componentes y hooks sin cambiar comportamiento.

## Phase 5: Verification / Cleanup

- [x] 5.1 Ejecutar `cd frontend && npm test -- --run` y corregir regresiones de la suite afectada.
- [x] 5.2 Ejecutar `cd frontend && npm run lint` y resolver warnings/errors introducidos.
- [x] 5.3 Revisar que todos los escenarios de specs (`shared-finance-ux-ritual`, `period-governance-ux`) estén cubiertos.
